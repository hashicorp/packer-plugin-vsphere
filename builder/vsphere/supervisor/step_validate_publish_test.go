// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package supervisor_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	imgregv1a1 "github.com/vmware-tanzu/image-registry-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func TestValidatePublish_Prepare(t *testing.T) {
	config := &supervisor.ValidatePublishConfig{}
	if actualErr := config.Prepare(); len(actualErr) != 0 {
		t.Fatalf("Prepare should not fail: %v", actualErr)
	}
}

func TestValidatePublish_Run(t *testing.T) {
	// Test with `publish_location_name` not set.
	step := &supervisor.StepValidatePublish{
		Config: &supervisor.ValidatePublishConfig{
			PublishLocationName: "",
		},
	}

	ctx := context.TODO()
	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)

	action := step.Run(ctx, state)
	if action == multistep.ActionHalt {
		if rawErr, ok := state.GetOk("error"); ok {
			t.Errorf("Error from running the step: %s", rawErr.(error))
		}
		t.Fatal("Step should NOT halt")
	}

	expectedOutput := []string{
		"Validating VM publish location...",
		"VM publish step will be skipped as the `publish_location_name` config is not set",
	}
	checkOutputLines(t, testWriter, expectedOutput)

	// Test with non-existing "publish_location_name".
	testPublishLocationName := "test-publish-location"
	testNamespace := "test-namespace"
	kubeClient := newFakeKubeClient()
	step.Config.PublishLocationName = testPublishLocationName
	state.Put(supervisor.StateKeySupervisorNamespace, testNamespace)
	state.Put(supervisor.StateKeyKubeClient, kubeClient)

	action = step.Run(ctx, state)
	if action != multistep.ActionHalt {
		t.Fatal("Step should halt")
	}

	expectedError := fmt.Sprintf("%q not found", testPublishLocationName)
	if rawErr, ok := state.GetOk("error"); ok {
		if !strings.Contains(rawErr.(error).Error(), expectedError) {
			t.Errorf("Expected error contains %v, but got %v", expectedError, rawErr.(error).Error())
		}
	}

	expectedOutput = []string{
		"Validating VM publish location...",
	}
	checkOutputLines(t, testWriter, expectedOutput)

	// Test with existing but non-writable "publish_location_name".
	clObj := &imgregv1a1.ContentLibrary{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      testPublishLocationName,
		},
		Spec: imgregv1a1.ContentLibrarySpec{
			Writable: false,
		},
	}
	kubeClient = newFakeKubeClient(clObj)
	state.Put(supervisor.StateKeyKubeClient, kubeClient)

	action = step.Run(ctx, state)
	if action != multistep.ActionHalt {
		t.Fatal("Step should halt")
	}

	expectedError = fmt.Sprintf("The specified publish location %q is not writable", testPublishLocationName)
	if rawErr, ok := state.GetOk("error"); ok {
		if rawErr.(error).Error() != expectedError {
			t.Errorf("Expected error is %v, but got %v", expectedError, rawErr.(error).Error())
		}
	}

	expectedOutput = []string{
		"Validating VM publish location...",
	}
	checkOutputLines(t, testWriter, expectedOutput)

	// Test with valid (existing and writable) "publish_location_name".
	clObj.Spec.Writable = true
	kubeClient = newFakeKubeClient(clObj)
	state.Put(supervisor.StateKeyKubeClient, kubeClient)

	action = step.Run(ctx, state)
	if action != multistep.ActionContinue {
		t.Fatal("Step should continue")
	}

	expectedOutput = []string{
		"Validating VM publish location...",
		"VM publish location is valid",
	}
	checkOutputLines(t, testWriter, expectedOutput)
}
