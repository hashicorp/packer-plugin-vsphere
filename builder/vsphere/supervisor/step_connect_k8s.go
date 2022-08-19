//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ConnectK8sConfig

package supervisor

import (
	"context"
	"fmt"
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

const (
	stateKeyKubeClient    = "kube_client"
	stateKeyDynamicClient = "dynamic_client"
	stateKeyK8sNamespace  = "k8s_namespace"
)

type ConnectK8sConfig struct {
	KubeconfigPath string `mapstructure:"kubeconfig_path"`
	K8sNamespace   string `mapstructure:"k8s_namespace"`
}

func (c *ConnectK8sConfig) Prepare() []error {
	// Set the kubeconfig path from KUBECONFIG env var or the default path if not provided.
	if c.KubeconfigPath == "" {
		if val := os.Getenv(clientcmd.RecommendedConfigPathEnvVar); val != "" {
			c.KubeconfigPath = val
		} else {
			c.KubeconfigPath = clientcmd.RecommendedHomeFile
		}
	}

	// Set the K8s namespace from current context if not provided.
	if c.K8sNamespace == "" {
		data, err := os.ReadFile(c.KubeconfigPath)
		if err != nil {
			return []error{fmt.Errorf("failed to read kubeconfig file: %s", err)}
		}
		kubeConfig, err := clientcmd.NewClientConfigFromBytes(data)
		if err != nil {
			return []error{fmt.Errorf("failed to parse kubeconfig file: %s", err)}
		}
		ns, _, err := kubeConfig.Namespace()
		if err != nil {
			return []error{fmt.Errorf("failed to get namespace from current context: %s", err)}
		}

		c.K8sNamespace = ns
	}

	return nil
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
	state.Put(stateKeyK8sNamespace, s.Config.K8sNamespace)

	logger.Info("Successfully connected to the Supervisor cluster")
	return multistep.ActionContinue
}

func (s *StepConnectK8s) Cleanup(multistep.StateBag) {}
