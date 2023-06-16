// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package supervisor_test

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func TestPublishSource_Prepare(t *testing.T) {
	config := &supervisor.PublishSourceConfig{}
	if actualErrs := config.Prepare(); len(actualErrs) != 0 {
		t.Fatalf("Prepare should NOT fail: %v", actualErrs)
	}

	if config.WatchPublishTimeoutSec != supervisor.DefaultWatchPublishTimeoutSec {
		t.Fatalf("Default timeout should be %d, but got %d",
			supervisor.DefaultWatchPublishTimeoutSec, config.WatchPublishTimeoutSec)
	}
}

func TestStepPublishSource_Run_Skip(t *testing.T) {
	// Initialize the step without `publish_location_name` set.
	config := &supervisor.PublishSourceConfig{
		WatchPublishTimeoutSec: 5,
	}
	step := &supervisor.StepPublishSource{
		Config: config,
	}

	// Set up required state for running this step.
	state := newBasicTestState(new(bytes.Buffer))
	state.Put(supervisor.StateKeyPublishLocationName, "")
	state.Put(supervisor.StateKeySourceName, "test-source")
	state.Put(supervisor.StateKeySupervisorNamespace, "test-ns")
	state.Put(supervisor.StateKeyKubeClient, newFakeKubeClient())
	state.Put(supervisor.StateKeyKeepInputArtifact, true)

	action := step.Run(context.TODO(), state)
	if action != multistep.ActionContinue {
		if rawErr, ok := state.GetOk("error"); ok {
			t.Errorf("Error from running the step: %s", rawErr.(error))
		}
		t.Fatal("Step should continue")
	}
}

func TestStepPublishSource_Run(t *testing.T) {
	// Initialize the step with `publish_location_name` set.
	config := &supervisor.PublishSourceConfig{
		WatchPublishTimeoutSec: 5,
	}
	step := &supervisor.StepPublishSource{
		Config: config,
	}

	testSourceName := "test-source-name"
	testImageName := "test-image-name"
	testPublishLocationName := "test-publish-location-name"
	testNamespace := "test-namespace"
	testPublishRequestName := "test-publish-request-name"
	VMPublishReqObj := newFakeVMPubReqObj(testNamespace, testPublishRequestName, testPublishLocationName)
	testKubeClient := newFakeKubeClient(VMPublishReqObj)

	// Set up required state for running this step.
	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)
	state.Put(supervisor.StateKeyPublishLocationName, testPublishLocationName)
	state.Put(supervisor.StateKeySourceName, testSourceName)
	state.Put(supervisor.StateKeySupervisorNamespace, testNamespace)
	state.Put(supervisor.StateKeyKubeClient, testKubeClient)
	state.Put(supervisor.StateKeyKeepInputArtifact, true)

	ctx := context.TODO()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		action := step.Run(ctx, state)
		if action == multistep.ActionHalt {
			if rawErr, ok := state.GetOk("error"); ok {
				t.Errorf("Error from running the step: %s", rawErr.(error))
			}
			t.Errorf("Step should NOT halt")
		}

		// check if the VirtualMachinePublishRequest object is created with the expected spec.
		objKey := client.ObjectKey{
			Name:      testPublishRequestName,
			Namespace: testNamespace,
		}
		if err := testKubeClient.Get(ctx, objKey, VMPublishReqObj); err != nil {
			t.Errorf("Failed to get the expected VirtualMachinePublishRequest object, err: %s", err.Error())
		}
		if VMPublishReqObj.Name != testPublishRequestName {
			t.Errorf("Expected VirtualMachinePublishRequest name to be '%s', got '%s'",
				testPublishRequestName, VMPublishReqObj.Name)
		}
		if VMPublishReqObj.Namespace != testNamespace {
			t.Errorf("Expected VirtualMachinePublishRequest namespace to be '%s', got '%s'",
				testNamespace, VMPublishReqObj.Namespace)
		}
		if VMPublishReqObj.Spec.Target.Location.Name != testPublishLocationName {
			t.Errorf("Expected VirtualMachinePublishRequest target location to be '%s', got '%s'",
				testPublishLocationName, VMPublishReqObj.Spec.Target.Location.Name)
		}

		expectedOutput := []string{
			"Publishing the source VM to \"test-publish-location-name\"",
			"Creating a VirtualMachinePublishRequest object",
			"Successfully created the VirtualMachinePublishRequest object",
			"Waiting for the VM publish request to complete...",
			"Successfully published the VM to image \"test-image-name\"",
			"Finished publishing the source VM",
		}
		checkOutputLines(t, testWriter, expectedOutput)
	}()

	// Wait for the watch to be established from Builder before updating the fake VirtualMachinePublishRequest resource below.
	for i := 0; i < step.Config.WatchPublishTimeoutSec; i++ {
		supervisor.Mu.Lock()
		if supervisor.IsWatchingVMPublish {
			supervisor.Mu.Unlock()
			break
		}
		supervisor.Mu.Unlock()
		time.Sleep(time.Second)
	}

	VMPublishReqObj.Status.Ready = false
	if err := testKubeClient.Update(ctx, VMPublishReqObj); err != nil {
		t.Errorf("Failed to update the VirtualMachinePublishRequest object status ready, err: %s", err.Error())
	}

	VMPublishReqObj.Status.Ready = true
	VMPublishReqObj.Status.ImageName = testImageName
	if err := testKubeClient.Update(ctx, VMPublishReqObj); err != nil {
		t.Errorf("Failed to update the VirtualMachinePublishRequest object status image name, err: %s", err.Error())
	}

	wg.Wait()
}

func TestStepPublishSource_Cleanup(t *testing.T) {
	// Test when 'keep_input_artifact' config is set to true (should skip cleanup).
	step := &supervisor.StepPublishSource{}
	step.KeepInputArtifact = true
	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)
	state.Put(supervisor.StateKeyVMPublishRequestCreated, true)
	step.Cleanup(state)

	expectedOutput := []string{"Skip cleaning up the VirtualMachinePublishRequest object as specified in config"}
	checkOutputLines(t, testWriter, expectedOutput)

	// Test when 'keep_input_artifact' config is false (should delete the VirtualMachinePublishRequest object).
	step.KeepInputArtifact = false
	step.SourceName = "test-source"
	step.Namespace = "test-namespace"
	vmPubReq := &vmopv1alpha1.VirtualMachinePublishRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-source",
			Namespace: "test-namespace",
		},
	}
	fakeClient := newFakeKubeClient(vmPubReq)
	step.KubeWatchClient = fakeClient
	state.Put(supervisor.StateKeyKeepInputArtifact, true)
	state.Put(supervisor.StateKeyVMPublishRequestCreated, true)
	step.Cleanup(state)

	// Check if the source objects are deleted from the cluster.
	ctx := context.TODO()
	objKey := client.ObjectKey{
		Name:      "test-source",
		Namespace: "test-namespace",
	}
	if err := fakeClient.Get(ctx, objKey, &vmopv1alpha1.VirtualMachinePublishRequest{}); !errors.IsNotFound(err) {
		t.Fatal("Expected the VirtualMachinePublishRequest object to be deleted")
	}

	// Check the output lines from the step runs.
	expectedOutput = []string{
		"Deleting the VirtualMachinePublishRequest object from Supervisor cluster",
		"Successfully deleted the VirtualMachinePublishRequest object",
	}
	checkOutputLines(t, testWriter, expectedOutput)
}

func newFakeVMPubReqObj(ns, name, publishLocation string) *vmopv1alpha1.VirtualMachinePublishRequest {
	return &vmopv1alpha1.VirtualMachinePublishRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: vmopv1alpha1.VirtualMachinePublishRequestSpec{
			Target: vmopv1alpha1.VirtualMachinePublishRequestTarget{
				Location: vmopv1alpha1.VirtualMachinePublishRequestTargetLocation{
					Name: publishLocation,
				},
			},
		},
	}
}
