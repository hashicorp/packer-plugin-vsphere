package supervisor_test

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func newFakeKubeClient(initObjs ...client.Object) client.WithWatch {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = vmopv1alpha1.AddToScheme(scheme)

	return fake.NewClientBuilder().WithObjects(initObjs...).WithScheme(scheme).Build()
}

func newFakeVMObj(namespace, name, vmIP string) *vmopv1alpha1.VirtualMachine {
	return &vmopv1alpha1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func newFakeVMServiceObj(namespace, name, ingressIP string) *vmopv1alpha1.VirtualMachineService {
	return &vmopv1alpha1.VirtualMachineService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Status: vmopv1alpha1.VirtualMachineServiceStatus{
			LoadBalancer: vmopv1alpha1.LoadBalancerStatus{
				Ingress: []vmopv1alpha1.LoadBalancerIngress{
					{
						IP: ingressIP,
					},
				},
			},
		},
	}
}

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
	vmServiceObj := newFakeVMServiceObj(testNamespace, testSourceName, testIngressIP)
	kubeClient := newFakeKubeClient(vmObj, vmServiceObj)

	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)
	state.Put(supervisor.StateKeyKubeClient, kubeClient)
	state.Put(supervisor.StateKeySupervisorNamespace, testNamespace)
	state.Put(supervisor.StateKeySourceName, testSourceName)

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
			t.Errorf("State '%s' should be '%s', but got '%s'", supervisor.StateKeyCommunicateIP, testVMIP, vmIP)
		}
		connectIP := state.Get(supervisor.StateKeyCommunicateIP)
		if connectIP != testIngressIP {
			t.Errorf("State '%s' should be '%s', but got '%s'", supervisor.StateKeyCommunicateIP, testIngressIP, connectIP)
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

	wg.Wait()
}
