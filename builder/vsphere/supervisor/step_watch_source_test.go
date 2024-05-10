// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package supervisor_test

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	imgregv1 "github.com/vmware-tanzu/image-registry-operator-api/api/v1alpha1"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func TestWatchSource_Prepare(t *testing.T) {
	config := &supervisor.WatchSourceConfig{}
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Prepare should NOT fail: %v", errs)
	}
	if config.WatchSourceTimeoutSec != supervisor.DefaultWatchTimeoutSec {
		t.Fatalf("Default timeout should be %d, but got %d", supervisor.DefaultWatchTimeoutSec, config.WatchSourceTimeoutSec)
	}
}

func TestWatchSource_Run(t *testing.T) {
	// Initialize the step with required configs.
	config := &supervisor.WatchSourceConfig{
		WatchSourceTimeoutSec: 60,
	}
	step := &supervisor.StepWatchSource{
		Config: config,
	}

	// Set up required state for running this step.
	testNamespace := "test-ns"
	testSourceName := "test-source"
	testVMIP := "1.2.3.4"
	testIngressIP := "5.6.7.8"
	vmObj := newFakeVMObj(testNamespace, testSourceName, testVMIP)
	vmServiceObj := newFakeVMServiceObj(testNamespace, testSourceName)
	kubeClient := newFakeKubeClient(vmObj, vmServiceObj)

	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)
	state.Put(supervisor.StateKeyKubeClient, kubeClient)
	state.Put(supervisor.StateKeySupervisorNamespace, testNamespace)
	state.Put(supervisor.StateKeySourceName, testSourceName)
	state.Put(supervisor.StateKeyVMServiceCreated, true)

	// Run this step in a new goroutine as it contains a blocking 'watch' process.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		action := step.Run(context.TODO(), state)
		if action == multistep.ActionHalt {
			if rawErr, ok := state.GetOk("error"); ok {
				t.Errorf("Error from running the step: %s", rawErr.(error))
			}
			t.Error("Step should NOT halt")
			return
		}

		// Check if all the required states are set correctly after the step is run.
		vmIP := state.Get(supervisor.StateKeyVMIP)
		if vmIP != testVMIP {
			t.Errorf("State %q should be %q, but got %q", supervisor.StateKeyCommunicateIP, testVMIP, vmIP)
		}
		connectIP := state.Get(supervisor.StateKeyCommunicateIP)
		if connectIP != testIngressIP {
			t.Errorf("State %q should be %q, but got %q", supervisor.StateKeyCommunicateIP, testIngressIP, connectIP)
		}

		// Check the output lines from the step runs.
		expectedOutput := []string{
			"Waiting for the source VM to be powered-on and accessible...",
			"Source VM is NOT powered-on yet, continue watching...",
			"Source VM is powered-on, waiting for an IP to be assigned...",
			fmt.Sprintf("Successfully obtained the source VM IP: %s", testVMIP),
			"Getting source VM ingress IP from the VMService object",
			fmt.Sprintf("Successfully retrieved the source VM ingress IP: %s", testIngressIP),
			"Source VM is now ready in Supervisor cluster",
		}
		checkOutputLines(t, testWriter, expectedOutput)
	}()

	// Wait for the watch to be established from Builder before updating the fake VM resource below.
	for i := 0; i < step.Config.WatchSourceTimeoutSec; i++ {
		supervisor.Mu.Lock()
		if supervisor.IsWatchingVM {
			supervisor.Mu.Unlock()
			break
		}
		supervisor.Mu.Unlock()
		time.Sleep(time.Second)
	}

	// Update the VM resource in the order of poweredOff => poweredOn => IP assigned.
	// In this way, we can test out the VM watch functionality and all output messages.
	ctx := context.TODO()
	opt := &client.UpdateOptions{}

	vmObj.Status.PowerState = vmopv1alpha1.VirtualMachinePoweredOff
	_ = kubeClient.Update(ctx, vmObj, opt)

	vmObj.Status.PowerState = vmopv1alpha1.VirtualMachinePoweredOn
	_ = kubeClient.Update(ctx, vmObj, opt)

	vmObj.Status.VmIp = testVMIP
	_ = kubeClient.Update(ctx, vmObj, opt)

	vmServiceObj.Status.LoadBalancer.Ingress[0].IP = testIngressIP
	_ = kubeClient.Update(ctx, vmServiceObj, opt)

	wg.Wait()
}

func newFakeVMObj(namespace, name, vmIP string) *vmopv1alpha1.VirtualMachine {
	return &vmopv1alpha1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func TestStepWatchSource_Cleanup(t *testing.T) {
	step := &supervisor.StepWatchSource{
		Namespace: testNamespace,
	}
	importedImage := &imgregv1.ContentLibraryItem{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testCLItemName,
			Namespace: testNamespace,
		},
	}
	fakeClient := newFakeKubeClient(importedImage)
	step.KubeWatchClient = fakeClient

	ctx := context.TODO()
	objKey := client.ObjectKey{
		Name:      testCLItemName,
		Namespace: testNamespace,
	}
	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)

	// 1. Test when 'clean_imported_image' config is not set.
	step.Cleanup(state)

	if err := fakeClient.Get(ctx, objKey, &imgregv1.ContentLibraryItem{}); err != nil {
		t.Fatal("The ContentLibraryItem object should still exist")
	}

	// 2. Test when 'clean_imported_image' is set but 'imported_image_name' is not set.
	state.Put(supervisor.StateKeyCleanImportedImage, true)
	step.Cleanup(state)

	if err := fakeClient.Get(ctx, objKey, &imgregv1.ContentLibraryItem{}); err != nil {
		t.Fatal("The ContentLibraryItem object should still exist")
	}
	// Check the output lines from the step runs.
	expectedOutput := []string{
		fmt.Sprintf("Skip cleaning imported image since config %s is not set", supervisor.StateKeyImportedImageName),
	}
	checkOutputLines(t, testWriter, expectedOutput)

	// 3. Test when 'clean_imported_image' and 'imported_image_name' are set to be true.
	state.Put(supervisor.StateKeyImportedImageName, testCLItemName)
	step.Cleanup(state)

	// Check if the imported image object is deleted from the cluster.
	if err := fakeClient.Get(ctx, objKey, &imgregv1.ContentLibraryItem{}); !errors.IsNotFound(err) {
		t.Fatal("Expected the ContentLibraryItem object to be deleted")
	}

	// Check the output lines from the step runs.
	expectedOutput = []string{
		fmt.Sprintf("Deleting the imported ContentLibraryItem object %s in namespace %s.", testCLItemName, testNamespace),
		fmt.Sprintf("Successfully deleted the ContentLibraryItem object %s in namespace %s.", testCLItemName, testNamespace),
	}
	checkOutputLines(t, testWriter, expectedOutput)
}

func newFakeVMServiceObj(namespace, name string) *vmopv1alpha1.VirtualMachineService {
	return &vmopv1alpha1.VirtualMachineService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Status: vmopv1alpha1.VirtualMachineServiceStatus{
			LoadBalancer: vmopv1alpha1.LoadBalancerStatus{
				Ingress: []vmopv1alpha1.LoadBalancerIngress{
					{
						IP: "",
					},
				},
			},
		},
	}
}
