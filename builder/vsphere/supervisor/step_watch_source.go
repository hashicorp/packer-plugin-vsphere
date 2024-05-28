// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type WatchSourceConfig

package supervisor

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultWatchTimeoutSec = 1800

	StateKeyVMIP          = "vm_ip"
	StateKeyCommunicateIP = "ip"
)

var (
	Mu           sync.Mutex
	IsWatchingVM bool
)

type WatchSourceConfig struct {
	// The timeout in seconds to wait for the source VM to be ready. Defaults to `1800`.
	WatchSourceTimeoutSec int `mapstructure:"watch_source_timeout_sec"`
}

func (c *WatchSourceConfig) Prepare() []error {
	if c.WatchSourceTimeoutSec == 0 {
		c.WatchSourceTimeoutSec = DefaultWatchTimeoutSec
	}

	return nil
}

type StepWatchSource struct {
	Config *WatchSourceConfig

	SourceName, Namespace string
	KubeWatchClient       client.WithWatch
}

func (s *StepWatchSource) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	logger := state.Get("logger").(*PackerLogger)
	logger.Info("Waiting for the source VM to be powered-on and accessible...")

	var err error
	defer func() {
		if err != nil {
			state.Put("error", err)
		}
	}()

	if err = s.initStep(state); err != nil {
		return multistep.ActionHalt
	}

	timedCtx, cancel := context.WithTimeout(ctx, time.Duration(s.Config.WatchSourceTimeoutSec)*time.Second)
	defer cancel()

	vmIP := ""
	vmIP, err = s.waitForVMReady(timedCtx, logger)
	if err != nil {
		return multistep.ActionHalt
	}
	state.Put(StateKeyVMIP, vmIP)

	// Only get the VM ingress IP if the VM service has been created (i.e. communicator is not 'none').
	if state.Get(StateKeyVMServiceCreated) == true {
		ingressIP := ""
		ingressIP, err = s.getVMIngressIP(timedCtx, logger)
		if err != nil {
			return multistep.ActionHalt
		}
		state.Put(StateKeyCommunicateIP, ingressIP)
	}

	logger.Info("Source VM is now ready in Supervisor cluster")
	return multistep.ActionContinue
}

func (s *StepWatchSource) Cleanup(state multistep.StateBag) {}

func (s *StepWatchSource) initStep(state multistep.StateBag) error {
	if err := CheckRequiredStates(state,
		StateKeyKubeClient,
		StateKeySupervisorNamespace,
		StateKeySourceName,
	); err != nil {
		return err
	}

	var (
		ok                    bool
		sourceName, namespace string
		kubeWatchClient       client.WithWatch
	)

	if sourceName, ok = state.Get(StateKeySourceName).(string); !ok {
		return fmt.Errorf("failed to cast %s to type string", StateKeySourceName)
	}
	if namespace, ok = state.Get(StateKeySupervisorNamespace).(string); !ok {
		return fmt.Errorf("failed to cast %s to type string", StateKeySupervisorNamespace)
	}
	if kubeWatchClient, ok = state.Get(StateKeyKubeClient).(client.WithWatch); !ok {
		return fmt.Errorf("failed to cast %s to type client.WithWatch", StateKeyKubeClient)
	}

	s.SourceName = sourceName
	s.Namespace = namespace
	s.KubeWatchClient = kubeWatchClient

	return nil
}

func (s *StepWatchSource) waitForVMReady(ctx context.Context, logger *PackerLogger) (string, error) {
	vmWatch, err := s.KubeWatchClient.Watch(ctx, &vmopv1alpha1.VirtualMachineList{}, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", s.SourceName),
		Namespace:     s.Namespace,
	})

	if err != nil {
		logger.Error("Failed to watch the VM object in Supervisor cluster")
		return "", err
	}

	defer func() {
		vmWatch.Stop()

		Mu.Lock()
		IsWatchingVM = false // This is used when mocking the watch process in test.
		Mu.Unlock()
	}()

	Mu.Lock()
	IsWatchingVM = true
	Mu.Unlock()

	for {
		select {
		case event := <-vmWatch.ResultChan():
			if event.Object == nil {
				continue
			}

			vmObj, ok := event.Object.(*vmopv1alpha1.VirtualMachine)
			if !ok {
				continue
			}

			vmIP := vmObj.Status.VmIp
			if vmIP != "" && net.ParseIP(vmIP) != nil && net.ParseIP(vmIP).To4() != nil {
				logger.Info("Successfully obtained the source VM IP: %s", vmIP)
				return vmIP, nil
			}

			// If the code reaches here, then the VM object is not ready yet.
			// Provide additional logging based on the current VM power state.
			vmPowerState := vmObj.Status.PowerState
			if vmPowerState == vmopv1alpha1.VirtualMachinePoweredOn {
				logger.Info("Source VM is powered-on, waiting for an IP to be assigned...")
			} else {
				logger.Info("Source VM is NOT powered-on yet, continue watching...")
			}

		case <-ctx.Done():
			return "", fmt.Errorf("timed out watching for source VM object to be ready")
		}
	}
}

func (s *StepWatchSource) getVMIngressIP(ctx context.Context, logger *PackerLogger) (string, error) {
	logger.Info("Getting source VM ingress IP from the VMService object")

	vmServiceObj := &vmopv1alpha1.VirtualMachineService{}
	vmServiceObjKey := client.ObjectKey{
		Namespace: s.Namespace,
		Name:      s.SourceName,
	}

	var vmIngressIP string
	err := retry.Config{
		RetryDelay: func() time.Duration {
			return 5 * time.Second
		},
		ShouldRetry: func(err error) bool {
			return !errors.Is(err, context.DeadlineExceeded)
		},
	}.Run(ctx, func(ctx context.Context) error {

		if err := s.KubeWatchClient.Get(ctx, vmServiceObjKey, vmServiceObj); err != nil {
			logger.Error("Failed to get the VMService object in Supervisor cluster")
			return err
		}

		ingress := vmServiceObj.Status.LoadBalancer.Ingress
		if len(ingress) == 0 || ingress[0].IP == "" {
			logger.Info("VMService object's ingress IP is empty, continue checking...")
			return errors.New("vmservice object's ingress IP address is empty")
		}

		logger.Info("Successfully retrieved the source VM ingress IP: %s", ingress[0].IP)
		vmIngressIP = ingress[0].IP
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("timed out checking for VMService object's ingress IP")
	}

	return vmIngressIP, nil
}
