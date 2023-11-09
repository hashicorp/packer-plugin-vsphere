// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package supervisor_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	imgregv1a1 "github.com/vmware-tanzu/image-registry-operator-api/api/v1alpha1"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func TestCheckRequiredStates(t *testing.T) {
	state := newBasicTestState(nil)
	err := supervisor.CheckRequiredStates(state, "logger")
	if err != nil {
		t.Errorf("Expected no error but got: %s", err.Error())
	}

	state.Put("test-key-1", "test-val-1")
	state.Put("test-key-2", "test-val-2")
	err = supervisor.CheckRequiredStates(state, "test-key-1", "test-key-2")
	if err != nil {
		t.Errorf("Expected no error but got: %s", err.Error())
	}

	expectErr := supervisor.CheckRequiredStates(state, "test-key-non-exist")
	if expectErr == nil {
		t.Errorf("Expected error but got nil")
	}
}

// Utility functions that are used in multiple test files.

func newBasicTestState(writer *bytes.Buffer) *multistep.BasicStateBag {
	state := new(multistep.BasicStateBag)
	ui := &packersdk.BasicUi{
		Writer: writer,
	}
	state.Put("logger", &supervisor.PackerLogger{UI: ui})

	return state
}

func checkOutputLines(t *testing.T, writer *bytes.Buffer, expectedLines []string) {
	for _, expected := range expectedLines {
		if actual := readLine(t, writer); actual != expected {
			t.Fatalf("Expected output %q but got %q", expected, actual)
		}
	}
}

func readLine(t *testing.T, writer *bytes.Buffer) string {
	actual, err := writer.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read line from writer, err: %s", err.Error())
	}

	// Skip "continue checking" line as it can be printed from the retry.
	if strings.Contains(actual, "continue checking") {
		return readLine(t, writer)
	}

	return strings.TrimSpace(actual)
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

func newFakeKubeClient(initObjs ...client.Object) client.WithWatch {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = vmopv1alpha1.AddToScheme(scheme)
	_ = imgregv1a1.AddToScheme(scheme)

	return fake.NewClientBuilder().WithObjects(initObjs...).WithScheme(scheme).Build()
}
