//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ConnectSupervisorConfig

package supervisor

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/pkg/errors"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
)

const (
	StateKeySupervisorNamespace = "supervisor_namespace"
	StateKeyKubeClient          = "kube_client"
)

type ConnectSupervisorConfig struct {
	// The path to kubeconfig file for accessing to the vSphere Supervisor cluster. Defaults to the value of `KUBECONFIG` envvar or `$HOME/.kube/config` if the envvar is not set.
	KubeconfigPath string `mapstructure:"kubeconfig_path"`
	// The Supervisor namespace to deploy the source VM. Defaults to the current context's namespace in kubeconfig.
	SupervisorNamespace string `mapstructure:"supervisor_namespace"`
}

func (c *ConnectSupervisorConfig) Prepare() []error {
	// Set the kubeconfig path from KUBECONFIG env var or the default home path if not provided.
	if c.KubeconfigPath == "" {
		if val := os.Getenv(clientcmd.RecommendedConfigPathEnvVar); val != "" {
			c.KubeconfigPath = val
		} else {
			c.KubeconfigPath = clientcmd.RecommendedHomeFile
		}
	}

	// Set the Supervisor namespace from current context if not provided.
	if c.SupervisorNamespace == "" {
		data, err := os.ReadFile(c.KubeconfigPath)
		if err != nil {
			return []error{errors.Wrap(err, "failed to read kubeconfig file")}
		}
		kubeConfig, err := clientcmd.NewClientConfigFromBytes(data)
		if err != nil {
			return []error{errors.Wrap(err, "failed to parse kubeconfig file")}
		}
		ns, _, err := kubeConfig.Namespace()
		if err != nil {
			return []error{errors.Wrap(err, "failed to get current context's namespace in kubeconfig file")}
		}

		c.SupervisorNamespace = ns
	}

	return nil
}

type StepConnectSupervisor struct {
	Config *ConnectSupervisorConfig
}

func (s *StepConnectSupervisor) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	logger := state.Get("logger").(*PackerLogger)
	logger.Info("Connecting to Supervisor cluster...")

	kubeClient, err := InitKubeClientFunc(s)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put(StateKeyKubeClient, kubeClient)
	state.Put(StateKeySupervisorNamespace, s.Config.SupervisorNamespace)

	logger.Info("Successfully connected to Supervisor cluster")
	return multistep.ActionContinue
}

func (s *StepConnectSupervisor) Cleanup(multistep.StateBag) {}

// Setting this function as a variable so that it can be mocked in test.
var InitKubeClientFunc = func(s *StepConnectSupervisor) (client.WithWatch, error) {
	config, err := clientcmd.BuildConfigFromFlags("", s.Config.KubeconfigPath)
	if err != nil {
		return nil, err
	}

	// The Supervisor builder will interact with both vmoperator and corev1 resources.
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = vmopv1alpha1.AddToScheme(scheme)

	// Initialize a WithWatch client as we need to watch the status of the source VM.
	return client.NewWithWatch(config, client.Options{Scheme: scheme})
}
