package supervisor_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hashicorp/packer-plugin-sdk/multistep"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func TestConnectSupervisor_Prepare(t *testing.T) {
	// Check when non-existing kubeconfig file is provided.
	config := &supervisor.ConnectSupervisorConfig{
		KubeconfigPath: "non-existing-file",
	}
	if err := config.Prepare(); err == nil {
		t.Fatalf("Prepare should fail by non-existing kubeconfig file")
	}

	// Check when an invalid kubeconfig file is provided.
	fakeFile, err := os.CreateTemp(t.TempDir(), "invalid-kubeconfig-file")
	if err != nil {
		t.Fatalf("Failed to create an invalid kubeconfig file: %v", err)
	}
	defer fakeFile.Close()
	config.KubeconfigPath = fakeFile.Name()
	if err := config.Prepare(); err == nil {
		t.Fatalf("Prepare should fail by an invalid kubeconfig file")
	}

	// Check kubeconfig path value when the KUBECONFIG env var is set.
	config.KubeconfigPath = ""
	validKubeconfigPath := getTestKubeconfigFile(t, "test-ns").Name()
	os.Setenv(clientcmd.RecommendedConfigPathEnvVar, validKubeconfigPath)
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Prepare should NOT fail: %v", errs)
	}
	if config.KubeconfigPath != validKubeconfigPath {
		t.Fatalf("config.KubeconfigPath should be '%s', but got '%s'",
			validKubeconfigPath, config.KubeconfigPath)
	}

	// Check if Supervisor namespace is set from the given kubeconfig file context.
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
