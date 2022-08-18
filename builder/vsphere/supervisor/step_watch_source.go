//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type WatchSourceConfig

package supervisor

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

const (
	defaultWatchTimeoutSec = 300

	stateKeyVMIP      = "vm_ip"
	stateKeyConnectIP = "ip"
)

type WatchSourceConfig struct {
	TimeoutSecond int64 `mapstructure:"watch_source_timeout_sec"`
}

func (c *WatchSourceConfig) Prepare() []error {
	if c.TimeoutSecond == 0 {
		c.TimeoutSecond = defaultWatchTimeoutSec
	}

	return nil
}

type StepWatchSource struct {
	Config *WatchSourceConfig
}

func (s *StepWatchSource) getVMIngressIP(
	ctx context.Context, logger *PackerLogger, kubeClient *kubernetes.Clientset, name, ns string) (string, error) {
	logger.Info("Getting source VM ingress IP from its K8s Service object")

	svc, err := kubeClient.CoreV1().Services(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		logger.Error("Failed to get source VM's Service object")
		return "", err
	}

	ingress := svc.Status.LoadBalancer.Ingress
	if len(ingress) == 0 || ingress[0].IP == "" {
		return "", fmt.Errorf("source VM's Service ingress IP is empty")
	}

	logger.Info("Successfully get the source VM ingress IP: %s", ingress[0].IP)
	return ingress[0].IP, nil
}

func (s *StepWatchSource) waitForVMReady(
	ctx context.Context, logger *PackerLogger, dynamicClient dynamic.Interface, name, ns string) (string, error) {
	logger.Info("Establishing a watch to the source VM object")

	vmWatch, err := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "vmoperator.vmware.com",
		Version:  "v1alpha1",
		Resource: "virtualmachines",
	}).Namespace(ns).Watch(ctx, metav1.ListOptions{
		FieldSelector:  fmt.Sprintf("metadata.name=%s", name),
		TimeoutSeconds: &s.Config.TimeoutSecond,
	})

	if err != nil {
		logger.Error("Failed to establish watch for source VM object")
		return "", err
	}

	for {
		select {
		case event := <-vmWatch.ResultChan():
			if event.Object == nil {
				return "", fmt.Errorf("timed out watching for source VM object to be ready")
			}

			obj := event.Object.(*unstructured.Unstructured)
			vmIP, found, err := unstructured.NestedString(obj.Object, "status", "vmIp")
			if err != nil {
				logger.Error("Failed to get the source VM IP")
				return "", err
			}

			if found && vmIP != "" {
				logger.Info("Successfully get the source VM IP: %s", vmIP)
				return vmIP, nil
			}

			// VM is not ready. Provide additional logging based on the current VM power state.
			vmPowerState, _, _ := unstructured.NestedString(obj.Object, "status", "powerState")
			if vmPowerState == "poweredOn" {
				logger.Info("Source VM is powered on, waiting for an IP to be assigned")
			} else {
				logger.Info("Source VM is NOT powered on yet, continue watching")
			}
		}
	}
}

func (s *StepWatchSource) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	logger := state.Get("logger").(*PackerLogger)
	logger.Info("Waiting for the source VM to be up and ready...")

	var err error
	defer func() {
		if err != nil {
			state.Put("error", err)
		}
	}()

	kubeClient := state.Get(stateKeyKubeClient).(*kubernetes.Clientset)
	dynamicClient := state.Get(stateKeyDynamicClient).(dynamic.Interface)
	if kubeClient == nil || dynamicClient == nil {
		err = fmt.Errorf("required K8s clients are nil from the StateBag")
		return multistep.ActionHalt
	}

	sourceName := state.Get(stateKeySourceName).(string)
	sourceNamespace := state.Get(stateKeySourceNamespace).(string)
	if sourceName == "" || sourceNamespace == "" {
		err = fmt.Errorf("required source name or namespace is empty from the StateBag")
		return multistep.ActionHalt
	}

	// Wait for the source VM to power up and have an IP assigned.
	vmIP, err := s.waitForVMReady(ctx, logger, dynamicClient, sourceName, sourceNamespace)
	if err != nil {
		return multistep.ActionHalt
	}
	state.Put(stateKeyVMIP, vmIP)

	ingressIP, err := s.getVMIngressIP(ctx, logger, kubeClient, sourceName, sourceNamespace)
	if err != nil {
		return multistep.ActionHalt
	}
	state.Put(stateKeyConnectIP, ingressIP)

	logger.Info("Source VM is now up and ready for customization")
	return multistep.ActionContinue
}

func (s *StepWatchSource) Cleanup(state multistep.StateBag) {}
