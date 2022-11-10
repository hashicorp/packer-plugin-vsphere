//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type PublishSourceConfig

package supervisor

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultWatchPublishTimeoutSec = 600

	StateKeyVMPublishRequestCreated = "vm_pub_req_created"
)

var IsWatchingVMPublish bool

type PublishSourceConfig struct {
	// The name of the published VM image. If not specified, the vm-operator API will set a default name.
	PublishImageName string `mapstructure:"publish_image_name"`
	// The timeout in seconds to wait for the VM to be published. Defaults to `600`.
	WatchPublishTimeoutSec int `mapstructure:"watch_publish_timeout_sec"`
}

func (c *PublishSourceConfig) Prepare() []error {
	if c.WatchPublishTimeoutSec == 0 {
		c.WatchPublishTimeoutSec = DefaultWatchPublishTimeoutSec
	}

	return nil
}

type StepPublishSource struct {
	Config *PublishSourceConfig

	PublishLocationName, SourceName, Namespace string
	KeepInputArtifact                          bool
	KubeWatchClient                            client.WithWatch
}

func (s *StepPublishSource) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	logger := state.Get("logger").(*PackerLogger)

	var err error
	defer func() {
		if err != nil {
			state.Put("error", err)
		}
	}()

	if err = s.initStep(state); err != nil {
		return multistep.ActionHalt
	}

	// Skip publishing if the publish location name is not specified.
	if s.PublishLocationName == "" {
		return multistep.ActionContinue
	}

	logger.Info("Publishing the source VM to %q", s.PublishLocationName)

	if err = s.createVMPublishRequest(ctx, logger); err != nil {
		return multistep.ActionHalt
	}
	state.Put(StateKeyVMPublishRequestCreated, true)

	if err = s.watchVMPublish(ctx, logger); err != nil {
		return multistep.ActionHalt
	}

	logger.Info("Finished publishing the source VM")

	return multistep.ActionContinue
}

func (s *StepPublishSource) Cleanup(state multistep.StateBag) {
	if state.Get(StateKeyVMPublishRequestCreated) == false {
		// Either the publish step was skipped or the object was not created successfully.
		// Skip deleting the VirtualMachinePublishRequest object.
		return
	}

	logger := state.Get("logger").(*PackerLogger)
	if s.KeepInputArtifact {
		logger.Info("Skip cleaning up the VirtualMachinePublishRequest object as specified in config")
		return
	}

	logger.Info("Deleting the VirtualMachinePublishRequest object from Supervisor cluster")
	ctx := context.Background()
	vmPubReqObj := &vmopv1alpha1.VirtualMachinePublishRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.SourceName,
			Namespace: s.Namespace,
		},
	}
	if err := s.KubeWatchClient.Delete(ctx, vmPubReqObj); err != nil {
		logger.Error("Failed to delete the VirtualMachinePublishRequest object")
	} else {
		logger.Info("Successfully deleted the VirtualMachinePublishRequest object")
	}
}

func (s *StepPublishSource) initStep(state multistep.StateBag) error {
	if err := CheckRequiredStates(state,
		StateKeyPublishLocationName,
		StateKeySourceName,
		StateKeySupervisorNamespace,
		StateKeyKubeClient,
		StateKeyKeepInputArtifact,
	); err != nil {
		return err
	}

	var ok bool
	if s.PublishLocationName, ok = state.Get(StateKeyPublishLocationName).(string); !ok {
		return fmt.Errorf("failed to cast %s to type string", StateKeyPublishLocationName)
	}
	if s.SourceName, ok = state.Get(StateKeySourceName).(string); !ok {
		return fmt.Errorf("failed to cast %s to type string", StateKeySourceName)
	}
	if s.Namespace, ok = state.Get(StateKeySupervisorNamespace).(string); !ok {
		return fmt.Errorf("failed to cast %s to type string", StateKeySupervisorNamespace)
	}
	if s.KubeWatchClient, ok = state.Get(StateKeyKubeClient).(client.WithWatch); !ok {
		return fmt.Errorf("failed to cast %s to type client.WithWatch", StateKeyKubeClient)
	}
	if s.KeepInputArtifact, ok = state.Get(StateKeyKeepInputArtifact).(bool); !ok {
		return fmt.Errorf("failed to cast %s to type bool", StateKeyKeepInputArtifact)
	}

	return nil
}

func (s *StepPublishSource) createVMPublishRequest(ctx context.Context, logger *PackerLogger) error {
	logger.Info("Creating a VirtualMachinePublishRequest object")

	vmPublishReq := &vmopv1alpha1.VirtualMachinePublishRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.SourceName,
			Namespace: s.Namespace,
		},
		Spec: vmopv1alpha1.VirtualMachinePublishRequestSpec{
			Target: vmopv1alpha1.VirtualMachinePublishRequestTarget{
				Location: vmopv1alpha1.VirtualMachinePublishRequestTargetLocation{
					Name: s.PublishLocationName,
				},
			},
		},
	}

	// Set the PublishImageName if provided in configs.
	if s.Config.PublishImageName != "" {
		vmPublishReq.Spec.Target.Item.Name = s.Config.PublishImageName
	}

	if err := s.KubeWatchClient.Create(ctx, vmPublishReq); err != nil {
		logger.Error("Failed to create the VirtualMachinePublishRequest object")
		return err
	}

	logger.Info("Successfully created the VirtualMachinePublishRequest object")
	return nil
}

func (s *StepPublishSource) watchVMPublish(ctx context.Context, logger *PackerLogger) error {
	vmPublishReqWatch, err := s.KubeWatchClient.Watch(ctx, &vmopv1alpha1.VirtualMachinePublishRequestList{}, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", s.SourceName),
		Namespace:     s.Namespace,
	})

	if err != nil {
		logger.Error("Failed to watch the VirtualMachinePublishRequest object in Supervisor cluster")
		return err
	}

	timedCtx, cancel := context.WithTimeout(ctx, time.Duration(s.Config.WatchPublishTimeoutSec)*time.Second)

	defer func() {
		vmPublishReqWatch.Stop()
		cancel()

		Mu.Lock()
		IsWatchingVMPublish = false
		Mu.Unlock()
	}()

	Mu.Lock()
	IsWatchingVMPublish = true
	Mu.Unlock()

	for {
		select {
		case event := <-vmPublishReqWatch.ResultChan():
			if event.Object == nil {
				return fmt.Errorf("watch VirtualMachinePublishRequest event object is nil")
			}

			vmPublishReqObj, ok := event.Object.(*vmopv1alpha1.VirtualMachinePublishRequest)
			if !ok {
				return fmt.Errorf("failed to convert the watch VirtualMachinePublishRequest event object")
			}

			if !vmPublishReqObj.Status.Ready {
				logger.Info("Waiting for the VM publish request to complete...")
			} else {
				logger.Info("Successfully published the VM to image %q", vmPublishReqObj.Status.ImageName)
				return nil
			}

		case <-timedCtx.Done():
			return fmt.Errorf("timed out watching for VirtualMachinePublishRequest object to complete")
		}
	}
}
