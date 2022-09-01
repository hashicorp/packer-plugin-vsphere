package supervisor_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func getTestKubeconfigFile(t *testing.T, namespace string) *os.File {
	fakeKubeconfigDataFmt := `
apiVersion: v1
clusters:
- cluster:
    server: test-server
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    namespace: %s
  name: test-context
current-context: test-context
kind: Config
`
	// The directory will be automatically removed when the test ends.
	tmpDir := t.TempDir()
	fakeFile, err := os.CreateTemp(tmpDir, "fake-test-file")
	if err != nil {
		t.Fatalf("Failed to create a fake kubeconfig file: %s", err)
	}
	defer fakeFile.Close()

	_, err = io.WriteString(fakeFile, fmt.Sprintf(fakeKubeconfigDataFmt, namespace))
	if err != nil {
		t.Fatalf("Failed to write to the fake kubeconfig file: %s", err)
	}

	return fakeFile
}

func TestConnectK8s_Prepare(t *testing.T) {
	// Check kubeconfig path when the KUBECONFIG env var is set.
	os.Setenv(clientcmd.RecommendedConfigPathEnvVar, "test-path")
	config := &supervisor.ConnectK8sConfig{
		K8sNamespace: "fake", // avoid reading the config file
	}
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Prepare should NOT fail: %v", errs)
	}
	if config.KubeconfigPath != "test-path" {
		t.Errorf("KubeconfigPath should be 'test-path' but got '%s'", config.KubeconfigPath)
	}

	// Check kubeconfig path when the KUBECONFIG env var is NOT set.
	config.KubeconfigPath = ""
	os.Unsetenv(clientcmd.RecommendedConfigPathEnvVar)
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Prepare should NOT fail: %s", errs[0])
	}
	if config.KubeconfigPath != clientcmd.RecommendedHomeFile {
		t.Errorf("KubeconfigPath should be '%s', but got '%s'", clientcmd.RecommendedHomeFile, config.KubeconfigPath)
	}

	// Check k8s namespace from the given kubeconfig file context.
	testFile := getTestKubeconfigFile(t, "test-ns")
	config.KubeconfigPath = testFile.Name()
	config.K8sNamespace = ""
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Prepare should NOT fail: %s", errs[0])
	}
	if config.K8sNamespace != "test-ns" {
		t.Errorf("K8sNamespace should be 'test-ns' but got '%s'", config.K8sNamespace)
	}
}

func TestConnectK8s_Run(t *testing.T) {
	// Set up required config and state for running the step.
	testFile := getTestKubeconfigFile(t, "test-ns")
	config := &supervisor.ConnectK8sConfig{
		KubeconfigPath: testFile.Name(),
		K8sNamespace:   "test-ns",
	}
	step := supervisor.StepConnectK8s{
		Config: config,
	}
	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)

	action := step.Run(context.TODO(), state)
	if action == multistep.ActionHalt {
		if rawErr, ok := state.GetOk("error"); ok {
			t.Errorf("Error from running the step: %s", rawErr.(error))
		}
		t.Fatalf("step should NOT halt")
	}

	// Check if all the required states are set after the step is run.
	if err := supervisor.CheckRequiredStates(state,
		supervisor.StateKeyKubeClientSet,
		supervisor.StateKeyKubeDynamicClient,
		supervisor.StateKeyK8sNamespace,
	); err != nil {
		t.Fatalf("Missing required states: %s", err)
	}

	// Check if the k8s namespace value is set correctly in the state.
	k8sNamespace := state.Get(supervisor.StateKeyK8sNamespace)
	if k8sNamespace != "test-ns" {
		t.Errorf("State '%s' should be 'test-ns', but got '%s'", supervisor.StateKeyK8sNamespace, k8sNamespace)
	}

	// Check the output lines from the step runs.
	expectedLines := []string{
		"Connecting to Supervisor K8s cluster...",
		"Successfully connected to the Supervisor cluster",
	}
	checkOutputLines(t, testWriter, expectedLines)
}
