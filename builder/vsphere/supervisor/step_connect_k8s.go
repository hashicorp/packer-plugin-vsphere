//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ConnectK8sConfig

package supervisor

import (
	"context"
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

const (
	stateKeyKubeClient    = "kube_client"
	stateKeyDynamicClient = "dynamic_client"
)

type ConnectK8sConfig struct {
	KubeconfigPath string `mapstructure:"kubeconfig_path"`
}

func (c *ConnectK8sConfig) Prepare() []error {
	var errs []error
	if c.KubeconfigPath == "" {
		errs = append(errs, fmt.Errorf("'kubeconfig_path' is required for connecting to the Kubernetes cluster"))
	}

	return errs
}

type StepConnectK8s struct {
	Config *ConnectK8sConfig
}

func (s *StepConnectK8s) getKubeClients() (*kubernetes.Clientset, dynamic.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", s.Config.KubeconfigPath)
	if err != nil {
		return nil, nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return kubeClient, dynamicClient, nil
}

func (s *StepConnectK8s) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	logger := state.Get("logger").(*PackerLogger)
	logger.Info("Connecting to Supervisor K8s cluster...")

	kubeClient, dynamicClient, err := s.getKubeClients()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put(stateKeyKubeClient, kubeClient)
	state.Put(stateKeyDynamicClient, dynamicClient)

	logger.Info("Successfully connected to the Supervisor cluster")
	return multistep.ActionContinue
}

func (s *StepConnectK8s) Cleanup(multistep.StateBag) {}
