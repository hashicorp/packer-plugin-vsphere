package supervisor_test

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func TestCreateSource_Prepare(t *testing.T) {
	// Check error output when missing the required config.
	config := &supervisor.CreateSourceConfig{}
	var actualErrs []error
	if actualErrs = config.Prepare(); len(actualErrs) == 0 {
		t.Fatalf("Prepare should fail by missing required configs, got empty")
	}

	expectedErrs := []error{
		fmt.Errorf("'image_name' is required for creating the source VM"),
		fmt.Errorf("'class_name' is required for creating the source VM"),
		fmt.Errorf("'storage_class' is required for creating the source VM"),
	}
	if !reflect.DeepEqual(actualErrs, expectedErrs) {
		t.Fatalf("Expected errs %v, got %v", expectedErrs, actualErrs)
	}

	// Check default values for the optional configs.
	config = &supervisor.CreateSourceConfig{
		ImageName:    "fake-image",
		ClassName:    "fake-class",
		StorageClass: "fake-storage-class",
	}
	if actualErrs = config.Prepare(); len(actualErrs) != 0 {
		t.Fatalf("Prepare should NOT fail: %v", actualErrs)
	}
	if !strings.HasPrefix(config.SourceName, supervisor.DefaultSourceNamePrefix) {
		t.Errorf("Expected default SourceName has prefix %s, got %s",
			supervisor.DefaultSourceNamePrefix, config.SourceName)
	}
}

func TestCreateSource_Run(t *testing.T) {
	// Initialize the step with required configs.
	config := &supervisor.CreateSourceConfig{
		ImageName:    "test-image",
		ClassName:    "test-class",
		StorageClass: "test-storage-class",
		SourceName:   "test-source",
	}
	commConfig := &communicator.Config{
		Type: "ssh",
		SSH: communicator.SSH{
			SSHUsername: "test-username",
			SSHPort:     123,
		},
	}
	step := &supervisor.StepCreateSource{
		Config:             config,
		CommunicatorConfig: commConfig,
	}

	// Set up required state for running this step.
	testNamespace := "test-namespace"
	kubeClient := newFakeKubeClient()
	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)
	state.Put(supervisor.StateKeyKubeClient, kubeClient)
	state.Put(supervisor.StateKeySupervisorNamespace, testNamespace)

	ctx := context.TODO()
	action := step.Run(ctx, state)
	if action == multistep.ActionHalt {
		if rawErr, ok := state.GetOk("error"); ok {
			t.Errorf("Error from running the step: %s", rawErr.(error))
		}
		t.Fatal("Step should NOT halt")
	}

	// Check if the K8s Secret object is created with expected spec.
	objKey := client.ObjectKey{
		Namespace: testNamespace,
		Name:      config.SourceName,
	}
	secretObj := &corev1.Secret{}
	if err := kubeClient.Get(ctx, objKey, secretObj); err != nil {
		t.Fatalf("Failed to get the expected Secret object, err: %s", err.Error())
	}
	if secretObj.StringData["user-data"] == "" {
		t.Errorf("Expected the Secret object to be created with user-data, got: %v", secretObj)
	}

	// Check if the source VM object is created with expected spec.
	vmObj := &vmopv1alpha1.VirtualMachine{}
	if err := kubeClient.Get(ctx, objKey, vmObj); err != nil {
		t.Fatalf("Failed to get the expected VM object, err: %s", err.Error())
	}
	if vmObj.Name != "test-source" {
		t.Errorf("Expected VM name to be 'test-vm', got '%s'", vmObj.Name)
	}
	if vmObj.Namespace != "test-namespace" {
		t.Errorf("Expected VM namespace to be 'test-namespace', got '%s'", vmObj.Namespace)
	}
	if vmObj.Spec.ImageName != "test-image" {
		t.Errorf("Expected VM image name to be 'test-image', got '%s'", vmObj.Spec.ImageName)
	}
	if vmObj.Spec.ClassName != "test-class" {
		t.Errorf("Expected VM class name to be 'test-class', got '%s'", vmObj.Spec.ClassName)
	}
	if vmObj.Spec.StorageClass != "test-storage-class" {
		t.Errorf("Expected VM storage class to be 'test-storage-class', got '%s'", vmObj.Spec.StorageClass)
	}
	selectorLabelVal := vmObj.Labels[supervisor.VMSelectorLabelKey]
	if selectorLabelVal != "test-source" {
		t.Errorf("Expected source VM label '%s' to be 'test-source', got '%s'", supervisor.VMSelectorLabelKey, selectorLabelVal)
	}

	// Check if the source VMService object is created with expected spec.
	vmServiceObj := &vmopv1alpha1.VirtualMachineService{}
	if err := kubeClient.Get(ctx, objKey, vmServiceObj); err != nil {
		t.Fatalf("Failed to get the expected VMService object, err: %s", err.Error())
	}
	if vmServiceObj.Name != "test-source" {
		t.Errorf("Expected VMService name to be 'test-source', got '%s'", vmServiceObj.Name)
	}
	if vmServiceObj.Namespace != "test-namespace" {
		t.Errorf("Expected VMService namespace to be 'test-namespace', got '%s'", vmServiceObj.Namespace)
	}
	if vmServiceObj.Spec.Type != "LoadBalancer" {
		t.Errorf("Expected VMService type to be 'LoadBalancer', got '%s'", vmServiceObj.Spec.Type)
	}
	ports := vmServiceObj.Spec.Ports
	if len(ports) == 0 || ports[0].Port != 123 || ports[0].TargetPort != 123 {
		t.Errorf("Expected VMService Port and TargetPort to be '123', got %v", ports)
	}
	selectorMap := vmServiceObj.Spec.Selector
	if val, ok := selectorMap[supervisor.VMSelectorLabelKey]; !ok || val != "test-source" {
		t.Errorf("Expected VMService selector '%s' to be 'test-source', got '%s'", supervisor.VMSelectorLabelKey, val)
	}

	// Check if all the required states are set correctly after the step is run.
	sourceName := state.Get(supervisor.StateKeySourceName)
	if sourceName != "test-source" {
		t.Errorf("State '%s' should be 'test-source', but got '%s'", supervisor.StateKeySourceName, sourceName)
	}
	if state.Get(supervisor.StateKeyVMCreated) != true {
		t.Errorf("State '%s' should be 'true'", supervisor.StateKeyVMCreated)
	}
	if state.Get(supervisor.StateKeyVMServiceCreated) != true {
		t.Errorf("State '%s' should be 'true'", supervisor.StateKeyVMServiceCreated)
	}
	if state.Get(supervisor.StateKeyVMMetadataSecretCreated) != true {
		t.Errorf("State '%s' should be 'true'", supervisor.StateKeyVMMetadataSecretCreated)
	}

	// Check the output lines from the step runs.
	expectedOutput := []string{
		"Creating required source objects in Supervisor cluster...",
		"Creating a K8s Secret object for providing source VM metadata",
		"Successfully created the K8s Secret object",
		"Creating a source VirtualMachine object",
		"Successfully created the VirtualMachine object",
		"Creating a VirtualMachineService object for network connection",
		"Successfully created the VirtualMachineService object",
		"Finished creating all required source objects in Supervisor cluster",
	}
	checkOutputLines(t, testWriter, expectedOutput)
}

func TestCreateSource_Cleanup(t *testing.T) {
	// Test when 'keep_input_artifact' config is set to true (should skip cleanup).
	config := &supervisor.CreateSourceConfig{
		KeepInputArtifact: true,
	}
	step := &supervisor.StepCreateSource{
		Config: config,
	}
	testWriter := &bytes.Buffer{}
	state := newBasicTestState(testWriter)
	step.Cleanup(state)

	expectedOutput := []string{"Skip cleaning up the source objects as specified in config"}
	checkOutputLines(t, testWriter, expectedOutput)

	// Test when 'keep_input_artifact' config is false (should delete all created source objects).
	step.Config.KeepInputArtifact = false
	step.Config.SourceName = "test-source"
	step.Namespace = "test-namespace"

	sourceObjMeta := metav1.ObjectMeta{
		Name:      "test-source",
		Namespace: "test-namespace",
	}
	kubeClient := newFakeKubeClient(
		&vmopv1alpha1.VirtualMachine{
			ObjectMeta: sourceObjMeta,
		},
		&vmopv1alpha1.VirtualMachineService{
			ObjectMeta: sourceObjMeta,
		},
		&corev1.Secret{
			ObjectMeta: sourceObjMeta,
		},
	)
	step.KubeClient = kubeClient

	state.Put(supervisor.StateKeyVMCreated, true)
	state.Put(supervisor.StateKeyVMServiceCreated, true)
	state.Put(supervisor.StateKeyVMMetadataSecretCreated, true)
	step.Cleanup(state)

	// Check if the source objects are deleted from the cluster.
	ctx := context.TODO()
	objKey := client.ObjectKey{
		Name:      "test-source",
		Namespace: "test-namespace",
	}
	if err := kubeClient.Get(ctx, objKey, &corev1.Secret{}); !errors.IsNotFound(err) {
		t.Fatal("Expected the Secret object to be deleted")
	}
	if err := kubeClient.Get(ctx, objKey, &vmopv1alpha1.VirtualMachine{}); !errors.IsNotFound(err) {
		t.Fatal("Expected the VirtualMachine object to be deleted")
	}
	if err := kubeClient.Get(ctx, objKey, &vmopv1alpha1.VirtualMachineService{}); !errors.IsNotFound(err) {
		t.Fatal("Expected the VirtualMachineService object to be deleted")
	}

	// Check the output lines from the step runs.
	expectedOutput = []string{
		"Deleting the VirtualMachineService object from Supervisor cluster",
		"Successfully deleted the VirtualMachineService object",
		"Deleting the VirtualMachine object from Supervisor cluster",
		"Successfully deleted the VirtualMachine object",
		"Deleting the K8s Secret object from Supervisor cluster",
		"Successfully deleted the K8s Secret object",
	}
	checkOutputLines(t, testWriter, expectedOutput)
}
