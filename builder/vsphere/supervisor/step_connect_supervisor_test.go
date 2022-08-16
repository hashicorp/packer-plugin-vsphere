package supervisor_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hashicorp/packer-plugin-sdk/multistep"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func TestConnectSupervisor_Prepare(t *testing.T) {
	// Check kubeconfig path when the KUBECONFIG env var is set.
	os.Setenv(clientcmd.RecommendedConfigPathEnvVar, "test-path")
	config := &supervisor.ConnectSupervisorConfig{
		SupervisorNamespace: "fake", // avoid reading the config file
	}
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Prepare should NOT fail: %v", errs)
	}
	if config.KubeconfigPath != "test-path" {
		t.Errorf("config.KubeconfigPath should be 'test-path', but got '%s'", config.KubeconfigPath)
	}

	// Check kubeconfig path when the KUBECONFIG env var is NOT set.
	config.KubeconfigPath = ""
	os.Unsetenv(clientcmd.RecommendedConfigPathEnvVar)
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Prepare should NOT fail: %s", errs[0])
	}
	if config.KubeconfigPath != clientcmd.RecommendedHomeFile {
		t.Errorf("config.KubeconfigPath should be '%s', but got '%s'", clientcmd.RecommendedHomeFile, config.KubeconfigPath)
	}

	// Check Supervisor namespace from the given kubeconfig file context.
	testFile := getTestKubeconfigFile(t, "test-ns")
	config.KubeconfigPath = testFile.Name()
	config.SupervisorNamespace = ""
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Prepare should NOT fail: %s", errs[0])
	}
	if config.SupervisorNamespace != "test-ns" {
		t.Errorf("Supervisor namespace should be 'test-ns' but got '%s'", config.SupervisorNamespace)
	}
}

func TestConnectSupervisor_Run(t *testing.T) {
	// Set up required config for running the step.
	testFile := getTestKubeconfigFile(t, "test-ns")
	config := &supervisor.ConnectSupervisorConfig{
		KubeconfigPath:      testFile.Name(),
		SupervisorNamespace: "test-ns",
	}
	step := supervisor.StepConnectSupervisor{
		Config: config,
	}
	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)

	// Mock the InitKubeClientFunc as controller-client always requires a valid kubeconfig to initialize.
	originalFunc := supervisor.InitKubeClientFunc
	defer func() {
		supervisor.InitKubeClientFunc = originalFunc
	}()
	supervisor.InitKubeClientFunc = func(s *supervisor.StepConnectSupervisor) (client.WithWatch, error) {
		return client.WithWatch(nil), nil
	}

	action := step.Run(context.TODO(), state)
	if action == multistep.ActionHalt {
		if rawErr, ok := state.GetOk("error"); ok {
			t.Errorf("Error from running the step: %s", rawErr.(error))
		}
		t.Fatalf("step should NOT halt")
	}

	// Check if all the required states are set after the step is run.
	if err := supervisor.CheckRequiredStates(state,
		supervisor.StateKeyKubeClient,
		supervisor.StateKeySupervisorNamespace,
	); err != nil {
		t.Fatalf("Missing required states: %s", err)
	}

	// Check if the Supervisor namespace value is set correctly in the state.
	namespace := state.Get(supervisor.StateKeySupervisorNamespace)
	if namespace != "test-ns" {
		t.Errorf("State '%s' should be 'test-ns', but got '%s'", supervisor.StateKeySupervisorNamespace, namespace)
	}

	// Check the output lines from the step runs.
	expectedLines := []string{
		"Connecting to Supervisor cluster...",
		"Successfully connected to Supervisor cluster",
	}
	checkOutputLines(t, testWriter, expectedLines)
}

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
