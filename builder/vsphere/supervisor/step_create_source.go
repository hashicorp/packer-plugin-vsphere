// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CreateSourceConfig

package supervisor

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator/api/v1alpha1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultSourceNamePrefix = "packer-vsphere-supervisor"
	VMSelectorLabelKey      = DefaultSourceNamePrefix + "-selector"

	StateKeySourceName              = "source_name"
	StateKeyVMCreated               = "vm_created"
	StateKeyVMServiceCreated        = "vm_service_created"
	StateKeyVMMetadataSecretCreated = "vm_metadata_secret_created"
	StateKeyKeepInputArtifact       = "keep_input_artifact"

	ProviderCloudInit  = string(vmopv1alpha1.VirtualMachineMetadataCloudInitTransport)
	ProviderSysprep    = string(vmopv1alpha1.VirtualMachineMetadataSysprepTransport)
	ProviderVAppConfig = string(vmopv1alpha1.VirtualMachineMetadataVAppConfigTransport)
)

type CreateSourceConfig struct {
	// Name of the source virtual machine (VM) image.
	ImageName string `mapstructure:"image_name" required:"true"`
	// Name of the VM class that describes virtual hardware settings.
	ClassName string `mapstructure:"class_name" required:"true"`
	// Name of the storage class that configures storage-related attributes.
	StorageClass string `mapstructure:"storage_class" required:"true"`

	// Name of the source VM. Defaults to `packer-vsphere-supervisor-<random-suffix>`.
	SourceName string `mapstructure:"source_name"`
	// Name of the network type to attach to the source VM's network interface. Defaults to empty.
	NetworkType string `mapstructure:"network_type"`
	// Name of the network to attach to the source VM's network interface. Defaults to empty.
	NetworkName string `mapstructure:"network_name"`
	// Preserve all the created objects in Supervisor cluster after the build finishes. Defaults to `false`.
	KeepInputArtifact bool `mapstructure:"keep_input_artifact"`
	// Name of the bootstrap provider to use for configuring the source VM.
	// Supported values are `CloudInit`, `Sysprep`, and `vAppConfig`. Defaults to `CloudInit`.
	BootstrapProvider string `mapstructure:"bootstrap_provider"`
	// Path to a file with bootstrap configuration data. Required if `bootstrap_provider` is not set to `CloudInit`.
	// Defaults to a basic cloud config that sets up the user account from the SSH communicator config.
	BootstrapDataFile string `mapstructure:"bootstrap_data_file"`
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

	bp := c.BootstrapProvider
	if bp == "" {
		c.BootstrapProvider = ProviderCloudInit
	} else if bp != ProviderCloudInit && bp != ProviderSysprep && bp != ProviderVAppConfig {
		errs = append(errs, fmt.Errorf("'bootstrap_provider' must be one of %q, %q, %q",
			ProviderCloudInit, ProviderSysprep, ProviderVAppConfig))
	} else if bp != ProviderCloudInit && c.BootstrapDataFile == "" {
		errs = append(errs, fmt.Errorf("'bootstrap_data_file' is required when 'bootstrap_provider' is %q", bp))
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

	if s.CommunicatorConfig.Type == "none" {
		logger.Info("Skip creating VirtualMachineService as communicator type is 'none'")
	} else {
		if err = s.createVMService(ctx, logger); err != nil {
			return multistep.ActionHalt
		}
		state.Put(StateKeyVMServiceCreated, true)
	}

	// Make the source_name and keep_input_artifact retrievable in later step.
	state.Put(StateKeySourceName, s.Config.SourceName)
	state.Put(StateKeyKeepInputArtifact, s.Config.KeepInputArtifact)

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
	logger.Info("Creating a K8s Secret object for providing source VM bootstrap data...")

	stringData, err := s.getBootstrapStringData(ctx, logger)
	if err != nil {
		logger.Error("Failed to get the bootstrap data")
		return err
	}

	kubeSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Config.SourceName,
			Namespace: s.Namespace,
		},
		StringData: stringData,
	}

	if err := s.KubeClient.Create(ctx, kubeSecret); err != nil {
		logger.Error("Failed to create the K8s Secret object")
		return err
	}

	logger.Info("Successfully created the K8s Secret object")
	return nil
}

func (s *StepCreateSource) getBootstrapStringData(ctx context.Context, logger *PackerLogger) (map[string]string, error) {
	if s.Config.BootstrapDataFile != "" {
		logger.Info("Loading bootstrap data from file: %s", s.Config.BootstrapDataFile)
		content, err := os.ReadFile(s.Config.BootstrapDataFile)
		if err != nil {
			return nil, err
		}
		var bootstrapData map[string]string
		err = yaml.Unmarshal(content, &bootstrapData)
		return bootstrapData, err
	}

	logger.Info("Using default cloud-init user data as the 'bootstrap_data_file' is not specified")

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
	defaultData := map[string]string{
		"user-data": cloudInitStr,
	}

	return defaultData, nil
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
				Transport:  vmopv1alpha1.VirtualMachineMetadataTransport(s.Config.BootstrapProvider),
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

	if err := s.KubeClient.Create(ctx, vm); err != nil {
		logger.Error("Failed to create the VirtualMachine object")
		return err
	}

	logger.Info("Successfully created the VirtualMachine object")
	return nil
}

func (s *StepCreateSource) createVMService(ctx context.Context, logger *PackerLogger) error {
	logger.Info("Creating a VirtualMachineService object for network connection")

	var commPort int
	switch s.CommunicatorConfig.Type {
	case "ssh":
		commPort = s.CommunicatorConfig.SSHPort
	case "winrm":
		commPort = s.CommunicatorConfig.WinRMPort
	default:
		return fmt.Errorf("unsupported communicator type: %q", s.CommunicatorConfig.Type)
	}

	vmServiceObj := &vmopv1alpha1.VirtualMachineService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Config.SourceName,
			Namespace: s.Namespace,
		},
		Spec: vmopv1alpha1.VirtualMachineServiceSpec{
			Type: vmopv1alpha1.VirtualMachineServiceTypeLoadBalancer,
			Ports: []vmopv1alpha1.VirtualMachineServicePort{
				{
					Name:       s.CommunicatorConfig.Type,
					Protocol:   "TCP",
					Port:       int32(commPort),
					TargetPort: int32(commPort),
				},
			},
			Selector: map[string]string{
				VMSelectorLabelKey: s.Config.SourceName,
			},
		},
	}

	if err := s.KubeClient.Create(ctx, vmServiceObj); err != nil {
		logger.Error("Failed to create the VirtualMachineService object")
		return err
	}

	logger.Info("Successfully created the VirtualMachineService object")
	return nil
}
