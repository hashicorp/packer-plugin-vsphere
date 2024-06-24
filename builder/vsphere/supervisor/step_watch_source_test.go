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
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func TestWatchSource_Prepare(t *testing.T) {
	config := &supervisor.WatchSourceConfig{}
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("unexpected failure: expected success, but failed: %v", errs[0])
	}
	if config.WatchSourceTimeoutSec != supervisor.DefaultWatchTimeoutSec {
		t.Fatalf("Default timeout should be %d, but returned %d", supervisor.DefaultWatchTimeoutSec, config.WatchSourceTimeoutSec)
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
				t.Errorf("unexpected error: %s", rawErr.(error))
			}
			t.Errorf("unexpected action: expected '%#v', but returned '%#v'", multistep.ActionContinue, action)
			return
		}

		// Check if all the required states are set correctly after the step is run.
		vmIP := state.Get(supervisor.StateKeyVMIP)
		if vmIP != testVMIP {
			t.Errorf("State %q should be %q, but returned %q", supervisor.StateKeyCommunicateIP, testVMIP, vmIP)
		}
		connectIP := state.Get(supervisor.StateKeyCommunicateIP)
		if connectIP != testIngressIP {
			t.Errorf("State %q should be %q, but returned %q", supervisor.StateKeyCommunicateIP, testIngressIP, connectIP)
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
