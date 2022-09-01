package supervisor_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
	restfake "k8s.io/client-go/rest/fake"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

var (
	testCreatedVMObj        *vmopv1alpha1.VirtualMachine
	testCreatedVMServiceObj *vmopv1alpha1.VirtualMachineService
)

func parseObjFromReq(t *testing.T, req *http.Request, obj interface{}) {
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	if err := json.Unmarshal(reqBody, obj); err != nil {
		t.Fatal("expected no error, got:", err)
	}
}

func mockRESTClientForCreateSourceCleanup(t *testing.T, sourceName, namespace string) *restfake.RESTClient {
	expectedReqPathFmt := "/apis/vmoperator.vmware.com/v1alpha1/namespaces/%s/%s/%s"
	expectedReqPath := ""

	return &restfake.RESTClient{
		Client: restfake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch req.Method {
			case "DELETE":
				// Check 'virtualmachineservices' first as it also contains 'virtualmachines'.
				if strings.Contains(req.URL.Path, "virtualmachineservices") {
					expectedReqPath = fmt.Sprintf(expectedReqPathFmt, namespace, "virtualmachineservices", sourceName)
				} else if strings.Contains(req.URL.Path, "virtualmachines") {
					expectedReqPath = fmt.Sprintf(expectedReqPathFmt, namespace, "virtualmachines", sourceName)
				} else {
					t.Fatalf("received unexpected resource: %s", req.URL.Path)
				}

				// Verify the client is deleting the correct resource.
				if req.URL.Path != expectedReqPath {
					t.Fatalf("Expected request path '%s' but got '%s'", expectedReqPath, req.URL.Path)
				}

			default:
				t.Fatalf("received unexpected method: %s", req.Method)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		}),
	}
}

func mockRESTClientForCreateSourceRun(t *testing.T) *restfake.RESTClient {
	// Reset the VM and VMService objects and parse them from the REST client requests.
	testCreatedVMObj = &vmopv1alpha1.VirtualMachine{}
	testCreatedVMServiceObj = &vmopv1alpha1.VirtualMachineService{}

	return &restfake.RESTClient{
		Client: restfake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch req.Method {
			case "POST":
				// Check 'virtualmachineservices' first as it also contains 'virtualmachines'.
				if strings.Contains(req.URL.Path, "virtualmachineservices") {
					parseObjFromReq(t, req, testCreatedVMServiceObj)
				} else if strings.Contains(req.URL.Path, "virtualmachines") {
					parseObjFromReq(t, req, testCreatedVMObj)
				} else {
					t.Fatalf("received unexpected resource: %s", req.URL.Path)
				}

			default:
				t.Fatalf("received unexpected method: %s", req.Method)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		}),
	}
}

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
	if config.SourceName != supervisor.DefaultSourceName {
		t.Fatalf("Expected default SourceName %s, got %s", supervisor.DefaultSourceName, config.SourceName)
	}
}

func TestCreateSource_Run(t *testing.T) {
	// Set up required config and state for running the step.
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
	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)
	testKubeClientSet := kubefake.NewSimpleClientset()
	state.Put(supervisor.StateKeyKubeClientSet, testKubeClientSet)
	state.Put(supervisor.StateKeyKubeRestClient, mockRESTClientForCreateSourceRun(t))
	state.Put(supervisor.StateKeyK8sNamespace, "test-ns")

	action := step.Run(context.TODO(), state)
	if action == multistep.ActionHalt {
		if rawErr, ok := state.GetOk("error"); ok {
			t.Errorf("Error from running the step: %s", rawErr.(error))
		}
		t.Fatal("Step should NOT halt")
	}

	// Check if the source Secret (VM-Metadata) object is created as expected.
	testCreatedSecretObj, err := testKubeClientSet.CoreV1().Secrets("test-ns").
		Get(context.TODO(), "test-source", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get the expected Secret object, err: %s", err.Error())
	}
	if testCreatedSecretObj == nil || testCreatedSecretObj.StringData["user-data"] == "" {
		t.Errorf("Expected Secret object to contain user-data, got: %v", testCreatedSecretObj)
	}

	// Check if the source VM object is created as expected.
	if testCreatedVMObj == nil {
		t.Fatal("Source VM object is nil")
	}
	if testCreatedVMObj.Name != "test-source" {
		t.Errorf("Expected source VM name 'test-vm', got '%s'", testCreatedVMObj.Name)
	}
	if testCreatedVMObj.Namespace != "test-ns" {
		t.Errorf("Expected source VM namespace 'test-ns', got '%s'", testCreatedVMObj.Namespace)
	}
	if testCreatedVMObj.Spec.ImageName != "test-image" {
		t.Errorf("Expected source VM image name 'test-image', got '%s'", testCreatedVMObj.Spec.ImageName)
	}
	if testCreatedVMObj.Spec.ClassName != "test-class" {
		t.Errorf("Expected source VM class name 'test-class', got '%s'", testCreatedVMObj.Spec.ClassName)
	}
	if testCreatedVMObj.Spec.StorageClass != "test-storage-class" {
		t.Errorf("Expected source VM storage class 'test-storage-class', got '%s'", testCreatedVMObj.Spec.StorageClass)
	}
	selectorLabelVal := testCreatedVMObj.Labels[supervisor.VMSelectorLabelKey]
	if selectorLabelVal != "test-source" {
		t.Errorf("Expected source VM label '%s' to be 'test-source', got '%s'", supervisor.VMSelectorLabelKey, selectorLabelVal)
	}

	// Check if the source VMService object is created as expected.
	if testCreatedVMServiceObj == nil {
		t.Fatal("source VMService object is nil")
	}
	if testCreatedVMServiceObj.Name != "test-source" {
		t.Errorf("Expected source VMService name 'test-source', got '%s'", testCreatedVMServiceObj.Name)
	}
	if testCreatedVMServiceObj.Namespace != "test-ns" {
		t.Errorf("Expected source VMService namespace 'test-ns', got '%s'", testCreatedVMServiceObj.Namespace)
	}
	if testCreatedVMServiceObj.Spec.Type != "LoadBalancer" {
		t.Errorf("Expected source VMService type 'LoadBalancer', got '%s'", testCreatedVMServiceObj.Spec.Type)
	}
	ports := testCreatedVMServiceObj.Spec.Ports
	if len(ports) == 0 || ports[0].Port != 123 || ports[0].TargetPort != 123 {
		t.Errorf("Expected source VMService Port/TargetPort 123, got %v", ports)
	}
	selectorMap := testCreatedVMServiceObj.Spec.Selector
	if val, ok := selectorMap[supervisor.VMSelectorLabelKey]; !ok || val != "test-source" {
		t.Errorf("Expected source VMService selector '%s' to be 'test-source', got '%s'",
			supervisor.VMSelectorLabelKey, val)
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
		"Creating source VM and its required objects in the connected Supervisor cluster...",
		"Initializing a source K8s Secret object for providing VM metadata",
		"Applying the source K8s Secret object with the kube CoreV1Client",
		"Created the source K8s Secret object",
		"Initializing a source VirtualMachine object for customization",
		"Creating the source VirtualMachine object with the kube REST client",
		"Created the source VirtualMachine object",
		"Initializing a source VMService object for setting up communication",
		"Creating the VMService object with the kube REST client",
		"Created the source VMService object",
		"Successfully created all required objects in the Supervisor cluster",
	}
	checkOutputLines(t, testWriter, expectedOutput)
}

func TestCreateSource_Cleanup(t *testing.T) {
	// Test when 'keep_source' config is set to true (should skip cleanup).
	config := &supervisor.CreateSourceConfig{
		SourceName: "test-source",
		KeepSource: true,
	}
	step := &supervisor.StepCreateSource{
		Config:       config,
		K8sNamespace: "test-ns",
	}
	testWriter := &bytes.Buffer{}
	state := newBasicTestState(testWriter)
	step.Cleanup(state)

	expectedOutput := []string{"Skip cleaning up the previously created source objects as configured"}
	checkOutputLines(t, testWriter, expectedOutput)

	// Test when 'keep_source' config is false (should delete all created source objects).
	step.Config.KeepSource = false
	step.Config.SourceName = "test-source"
	step.KubeRestClient = mockRESTClientForCreateSourceCleanup(t, "test-source", "test-ns")
	fakeSecretObj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      step.Config.SourceName,
			Namespace: step.K8sNamespace,
		},
	}
	testKubeClientSet := kubefake.NewSimpleClientset(fakeSecretObj)
	step.KubeClientSet = testKubeClientSet
	state.Put(supervisor.StateKeyVMCreated, true)
	state.Put(supervisor.StateKeyVMServiceCreated, true)
	state.Put(supervisor.StateKeyVMMetadataSecretCreated, true)
	step.Cleanup(state)

	// Check if the Secret object got deleted from the Kube ClientSet.
	// The other objects deletion is checked in the mock REST client.
	_, err := testKubeClientSet.CoreV1().Secrets(step.K8sNamespace).
		Get(context.TODO(), step.Config.SourceName, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		t.Fatal("Expected source Secret object to be deleted")
	}

	// Check the output lines from the step runs.
	expectedOutput = []string{
		"Cleaning up the previously created source objects from Supervisor cluster...",
		"Deleting the source VirtualMachineService object",
		"Deleted the source VirtualMachineService object",
		"Deleting the source VirtualMachine object",
		"Deleted the source VirtualMachine object",
		"Deleting source VMMetadata Secret object",
		"Deleted the source VMMetadata Secret object",
	}
	checkOutputLines(t, testWriter, expectedOutput)
}
