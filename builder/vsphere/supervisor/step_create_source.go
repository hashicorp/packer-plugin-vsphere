//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CreateSourceConfig

package supervisor

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
)

const (
	defaultSourceName  = "packer-supervisor-built-source"
	vmSelectorLabelKey = defaultSourceName + "-selector"

	stateKeyVMCreated               = "vm_created"
	stateKeyVMServiceCreated        = "vm_service_created"
	stateKeyVMMetadataSecretCreated = "vm_metadata_secret_created"
	stateKeySourceName              = "source_name"
	stateKeySourceNamespace         = "source_namespace"
)

var (
	vmopAPIVersion = vmopv1alpha1.SchemeGroupVersion.String()
)

type CreateSourceConfig struct {
	// Required configs.
	Namespace    string `mapstructure:"namespace"`
	ImageName    string `mapstructure:"image_name"`
	ClassName    string `mapstructure:"class_name"`
	StorageClass string `mapstructure:"storage_class"`

	// Optional configs.
	SourceName  string `mapstructure:"source_name"`
	NetworkType string `mapstructure:"network_type"`
	NetworkName string `mapstructure:"network_name"`
	KeepSource  bool   `mapstructure:"keep_source"`
}

func (c *CreateSourceConfig) Prepare() []error {
	var errs []error
	if c.Namespace == "" {
		errs = append(errs, fmt.Errorf("'namespace' is required for creating the source VM"))
	}
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
		c.SourceName = defaultSourceName
	}

	return errs
}

type StepCreateSource struct {
	Config             *CreateSourceConfig
	CommunicatorConfig *communicator.Config
}

func (s *StepCreateSource) createVMService(ctx context.Context, logger *PackerLogger, kubeClient *kubernetes.Clientset) error {
	logger.Info("Initializing a source VMService object for setting up communication")

	sshPort := int32(s.CommunicatorConfig.SSHPort)
	vmService := vmopv1alpha1.VirtualMachineService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: vmopAPIVersion,
			Kind:       "VirtualMachineService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Config.SourceName,
			Namespace: s.Config.Namespace,
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
				vmSelectorLabelKey: s.Config.SourceName,
			},
		},
	}

	vmServiceJSON, err := json.Marshal(vmService)
	if err != nil {
		return err
	}

	logger.Info("Creating the VMService object with the kube REST client")
	data, err := kubeClient.RESTClient().
		Post().
		AbsPath(fmt.Sprintf("/apis/%s", vmopAPIVersion)).
		Namespace(s.Config.Namespace).
		Resource("virtualmachineservices").
		Body(vmServiceJSON).
		DoRaw(ctx)

	if err != nil {
		logger.Error(
			"Failed to create source VMService object\nResponse from K8s API-Server: %s",
			string(data),
		)
		return err
	}

	logger.Info("Created the source VMService object")
	return nil
}

func (s *StepCreateSource) createVM(ctx context.Context, logger *PackerLogger, kubeClient *kubernetes.Clientset) error {
	logger.Info("Initializing a source VirtualMachine object for customization")

	vm := &vmopv1alpha1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: vmopAPIVersion,
			Kind:       "VirtualMachine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Config.SourceName,
			Namespace: s.Config.Namespace,
			Labels: map[string]string{
				vmSelectorLabelKey: s.Config.SourceName,
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

	// Set up network interface configs if provided by users.
	if s.Config.NetworkType != "" || s.Config.NetworkName != "" {
		vm.Spec.NetworkInterfaces = []vmopv1alpha1.VirtualMachineNetworkInterface{
			{
				NetworkType: s.Config.NetworkType,
				NetworkName: s.Config.NetworkName,
			},
		}
	}

	vmJSON, err := json.Marshal(vm)
	if err != nil {
		return err
	}

	logger.Info("Creating the source VirtualMachine object with the kube REST client")
	data, err := kubeClient.RESTClient().
		Post().
		AbsPath(fmt.Sprintf("/apis/%s", vmopAPIVersion)).
		Namespace(s.Config.Namespace).
		Resource("virtualmachines").
		Body(vmJSON).
		DoRaw(ctx)

	if err != nil {
		logger.Error(
			"Failed to create source VirtualMachine object\nResponse from K8s API-Server: %s",
			string(data),
		)
		return err
	}

	logger.Info("Created the source VirtualMachine object")
	return nil
}

func (s *StepCreateSource) createVMMetadataSecret(ctx context.Context, logger *PackerLogger, kubeClient *kubernetes.Clientset) error {
	logger.Info("Initializing a source K8s Secret object for providing VM metadata")
	cloudInitFmt := `#cloud-config
users:
  - name: %s
    lock_passwd: false
    plain_text_passwd: %s
    ssh_authorized_keys:
    - %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
write_files:
  - content: |
      Packer Plugin Says Hello World
    path: /helloworld
`
	cloudInitStr := fmt.Sprintf(cloudInitFmt,
		s.CommunicatorConfig.SSHUsername,
		s.CommunicatorConfig.SSHPassword,
		s.CommunicatorConfig.SSHPublicKey,
	)
	data := map[string]string{
		"user-data": cloudInitStr,
	}
	kubeSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Config.SourceName,
			Namespace: s.Config.Namespace,
		},
		StringData: data,
	}

	logger.Info("Applying the source K8s Secret object with the kube CoreV1Client")
	_, err := kubeClient.CoreV1().Secrets(s.Config.Namespace).Create(ctx, kubeSecret, metav1.CreateOptions{})
	if err != nil {
		logger.Error("Failed to create source Secret object")
		return err
	}

	logger.Info("Created the source K8s Secret object")
	return nil
}

func (s *StepCreateSource) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	logger := state.Get("logger").(*PackerLogger)
	logger.Info("Creating source VM and its required objects in the connected Supervisor cluster...")

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

	if err = s.createVMMetadataSecret(ctx, logger, kubeClient); err != nil {
		return multistep.ActionHalt
	}
	state.Put(stateKeyVMMetadataSecretCreated, true)

	if err = s.createVM(ctx, logger, kubeClient); err != nil {
		return multistep.ActionHalt
	}
	state.Put(stateKeyVMCreated, true)

	if err = s.createVMService(ctx, logger, kubeClient); err != nil {
		return multistep.ActionHalt
	}
	state.Put(stateKeyVMServiceCreated, true)

	// Storing the created source info to be used in the next step.
	state.Put(stateKeySourceNamespace, s.Config.Namespace)
	state.Put(stateKeySourceName, s.Config.SourceName)

	logger.Info("Successfully created all required objects in the Supervisor cluster")
	return multistep.ActionContinue
}

func (s *StepCreateSource) Cleanup(state multistep.StateBag) {
	logger := state.Get("logger").(*PackerLogger)

	if s.Config.KeepSource {
		logger.Info("Skip cleaning up the previously created source objects as configured")
		return
	}

	logger.Info("Cleaning up the previously created source objects from Supervisor cluster...")
	kubeClient := state.Get(stateKeyKubeClient).(*kubernetes.Clientset)
	if kubeClient == nil {
		logger.Error("kube client is nil from the StateBag")
		return
	}

	if state.Get(stateKeyVMServiceCreated) == true {
		logger.Info("Deleting the source VirtualMachineService object")
		data, err := kubeClient.RESTClient().
			Delete().
			AbsPath(fmt.Sprintf("/apis/%s", vmopAPIVersion)).
			Namespace(s.Config.Namespace).
			Resource("virtualmachineservices").
			Name(s.Config.SourceName).
			DoRaw(context.Background())
		if err != nil {
			logger.Error("Failed to delete source VirtualMachineService object: %s", string(data))
		} else {
			logger.Info("Deleted the source VirtualMachineService object")
		}
	}

	if state.Get(stateKeyVMCreated) == true {
		logger.Info("Deleting the source VirtualMachine object")
		data, err := kubeClient.RESTClient().
			Delete().
			AbsPath(fmt.Sprintf("/apis/%s", vmopAPIVersion)).
			Namespace(s.Config.Namespace).
			Resource("virtualmachines").
			Name(s.Config.SourceName).
			DoRaw(context.Background())
		if err != nil {
			logger.Error("Failed to delete source VirtualMachine object: %v", string(data))
		} else {
			logger.Info("Deleted the source VirtualMachine object")
		}
	}

	if state.Get(stateKeyVMMetadataSecretCreated) == true {
		logger.Info("Deleting source VMMetadata Secret object")
		err := kubeClient.CoreV1().Secrets(s.Config.Namespace).Delete(
			context.Background(), s.Config.SourceName, metav1.DeleteOptions{})
		if err != nil {
			logger.Error("Failed to delete source VMMetadata K8s Secret object: %s", err)
		} else {
			logger.Info("Deleted the source VMMetadata Secret object")
		}
	}
}
