// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CreateSourceConfig

package supervisor

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
)

const (
	DefaultSourceNamePrefix = "packer-vsphere-supervisor"
	VMSelectorLabelKey      = DefaultSourceNamePrefix + "-selector"

	StateKeySourceName              = "source_name"
	StateKeyVMCreated               = "vm_created"
	StateKeyVMServiceCreated        = "vm_service_created"
	StateKeyVMMetadataSecretCreated = "vm_metadata_secret_created"
)

type CreateSourceConfig struct {
	// Name of the source virtual machine (VM) image.
	ImageName string `mapstructure:"image_name" required:"true"`
	// Name of the VM class that describes virtual hardware settings.
	ClassName string `mapstructure:"class_name" required:"true"`
	// Name of the storage class that configures storage-related attributes.
	StorageClass string `mapstructure:"storage_class" required:"true"`

	// Name of the source VM. Defaults to `packer-vsphere-supervisor-built-source`.
	SourceName string `mapstructure:"source_name"`
	// Name of the network type to attach to the source VM's network interface. Defaults to empty.
	NetworkType string `mapstructure:"network_type"`
	// Name of the network to attach to the source VM's network interface. Defaults to empty.
	NetworkName string `mapstructure:"network_name"`
	// Preserve the created objects even after importing them to the vSphere endpoint. Defaults to `false`.
	KeepInputArtifact bool `mapstructure:"keep_input_artifact"`
}

func (c *CreateSourceConfig) Prepare() []error {
	var errs []error

	if c.ImageName == "" {
		errs = append(errs, fmt.Errorf("'image_name' is required for creating the source VM"))
	}
	if c.ClassName == "" {
		errs = append(errs, fmt.Errorf("'class_name' is required for creating the source VM"))
	}
	if c.StorageClass == "" {
		errs = append(errs, fmt.Errorf("'storage_class' is required for creating the source VM"))
	}

	if c.SourceName == "" {
		c.SourceName = fmt.Sprintf("%s-%s", DefaultSourceNamePrefix, rand.String(5))
	}

	return errs
}

type StepCreateSource struct {
	Config             *CreateSourceConfig
	CommunicatorConfig *communicator.Config

	Namespace  string
	KubeClient client.Client
}

func (s *StepCreateSource) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	logger := state.Get("logger").(*PackerLogger)
	logger.Info("Creating required source objects in Supervisor cluster...")

	var err error
	defer func() {
		if err != nil {
			state.Put("error", err)
		}
	}()

	if err = s.initStep(state); err != nil {
		return multistep.ActionHalt
	}

	if err = s.createVMMetadataSecret(ctx, logger); err != nil {
		return multistep.ActionHalt
	}
	state.Put(StateKeyVMMetadataSecretCreated, true)

	if err = s.createVM(ctx, logger); err != nil {
		return multistep.ActionHalt
	}
	state.Put(StateKeyVMCreated, true)

	if err = s.createVMService(ctx, logger); err != nil {
		return multistep.ActionHalt
	}
	state.Put(StateKeyVMServiceCreated, true)

	// Make the source name retrievable in later step.
	state.Put(StateKeySourceName, s.Config.SourceName)

	logger.Info("Finished creating all required source objects in Supervisor cluster")
	return multistep.ActionContinue
}

func (s *StepCreateSource) Cleanup(state multistep.StateBag) {
	logger := state.Get("logger").(*PackerLogger)

	if s.Config.KeepInputArtifact {
		logger.Info("Skip cleaning up the source objects as specified in config")
		return
	}

	ctx := context.Background()
	objMeta := metav1.ObjectMeta{
		Name:      s.Config.SourceName,
		Namespace: s.Namespace,
	}
	if state.Get(StateKeyVMServiceCreated) == true {
		logger.Info("Deleting the VirtualMachineService object from Supervisor cluster")
		vmServiceObj := &vmopv1alpha1.VirtualMachineService{
			ObjectMeta: objMeta,
		}
		if err := s.KubeClient.Delete(ctx, vmServiceObj); err != nil {
			logger.Error("Failed to delete the VirtualMachineService object")
		} else {
			logger.Info("Successfully deleted the VirtualMachineService object")
		}
	}

	if state.Get(StateKeyVMCreated) == true {
		logger.Info("Deleting the VirtualMachine object from Supervisor cluster")
		vmObj := &vmopv1alpha1.VirtualMachine{
			ObjectMeta: objMeta,
		}
		if err := s.KubeClient.Delete(ctx, vmObj); err != nil {
			logger.Error("Failed to delete the VirtualMachine object")
		} else {
			logger.Info("Successfully deleted the VirtualMachine object")
		}
	}

	if state.Get(StateKeyVMMetadataSecretCreated) == true {
		logger.Info("Deleting the K8s Secret object from Supervisor cluster")
		secretObj := &corev1.Secret{
			ObjectMeta: objMeta,
		}
		err := s.KubeClient.Delete(ctx, secretObj)
		if err != nil {
			logger.Error("Failed to delete the K8s Secret object: %s", err)
		} else {
			logger.Info("Successfully deleted the K8s Secret object")
		}
	}
}

func (s *StepCreateSource) initStep(state multistep.StateBag) error {
	if err := CheckRequiredStates(state,
		StateKeyKubeClient,
		StateKeySupervisorNamespace,
	); err != nil {
		return err
	}

	var (
		ok         bool
		namespace  string
		kubeClient client.Client
	)

	if namespace, ok = state.Get(StateKeySupervisorNamespace).(string); !ok {
		return fmt.Errorf("failed to cast %q from state bag as type 'string'", StateKeySupervisorNamespace)
	}
	if kubeClient, ok = state.Get(StateKeyKubeClient).(client.Client); !ok {
		return fmt.Errorf("failed to cast %q from state bag as type 'client.Client'", StateKeyKubeClient)
	}

	s.Namespace = namespace
	s.KubeClient = kubeClient
	return nil
}

func (s *StepCreateSource) createVMMetadataSecret(ctx context.Context, logger *PackerLogger) error {
	logger.Info("Creating a K8s Secret object for providing source VM metadata")

	cloudInitFmt := `#cloud-config
ssh_pwauth: true
users:
  - name: %s
    plain_text_passwd: %s
    lock_passwd: false
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
    - %s
`
	cloudInitStr := fmt.Sprintf(cloudInitFmt,
		s.CommunicatorConfig.SSHUsername,
		s.CommunicatorConfig.SSHPassword,
		strings.TrimSpace(string(s.CommunicatorConfig.SSHPublicKey)),
	)
	stringData := map[string]string{
		"user-data": cloudInitStr,
	}

	kubeSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Config.SourceName,
			Namespace: s.Namespace,
		},
		StringData: stringData,
	}
	err := s.KubeClient.Create(ctx, kubeSecret)
	if err != nil {
		logger.Error("Failed to create the K8s Secret object")
		return err
	}

	logger.Info("Successfully created the K8s Secret object")
	return nil
}

func (s *StepCreateSource) createVM(ctx context.Context, logger *PackerLogger) error {
	logger.Info("Creating a source VirtualMachine object")

	vm := &vmopv1alpha1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Config.SourceName,
			Namespace: s.Namespace,
			Labels: map[string]string{
				VMSelectorLabelKey: s.Config.SourceName,
			},
		},
		Spec: vmopv1alpha1.VirtualMachineSpec{
			ImageName:    s.Config.ImageName,
			ClassName:    s.Config.ClassName,
			StorageClass: s.Config.StorageClass,
			PowerState:   vmopv1alpha1.VirtualMachinePoweredOn,
			VmMetadata: &vmopv1alpha1.VirtualMachineMetadata{
				SecretName: s.Config.SourceName,
				Transport:  vmopv1alpha1.VirtualMachineMetadataCloudInitTransport,
			},
		},
	}

	// Set up network interface configs if provided in configs.
	if s.Config.NetworkType != "" || s.Config.NetworkName != "" {
		vm.Spec.NetworkInterfaces = []vmopv1alpha1.VirtualMachineNetworkInterface{
			{
				NetworkType: s.Config.NetworkType,
				NetworkName: s.Config.NetworkName,
			},
		}
	}

	err := s.KubeClient.Create(ctx, vm)
	if err != nil {
		logger.Error("Failed to create the VirtualMachine object")
		return err
	}

	logger.Info("Successfully created the VirtualMachine object")
	return nil
}

func (s *StepCreateSource) createVMService(ctx context.Context, logger *PackerLogger) error {
	logger.Info("Creating a VirtualMachineService object for network connection")

	sshPort := int32(s.CommunicatorConfig.SSHPort)
	vmServiceObj := &vmopv1alpha1.VirtualMachineService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Config.SourceName,
			Namespace: s.Namespace,
		},
		Spec: vmopv1alpha1.VirtualMachineServiceSpec{
			Type: vmopv1alpha1.VirtualMachineServiceTypeLoadBalancer,
			Ports: []vmopv1alpha1.VirtualMachineServicePort{
				{
					Name:       "ssh",
					Protocol:   "TCP",
					Port:       sshPort,
					TargetPort: sshPort,
				},
			},
			Selector: map[string]string{
				VMSelectorLabelKey: s.Config.SourceName,
			},
		},
	}

	err := s.KubeClient.Create(ctx, vmServiceObj)
	if err != nil {
		logger.Error("Failed to create the VirtualMachineService object")
		return err
	}

	logger.Info("Successfully created the VirtualMachineService object")
	return nil
}
