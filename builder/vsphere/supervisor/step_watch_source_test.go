package supervisor_test

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	daynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func newUnstructuredVM(namespace, name, vmIP string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "vmoperator.vmware.com/v1alpha1",
			"kind":       "VirtualMachine",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
		},
	}
}

func newFakeLBServiceObj(namespace, name, ip string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{
						IP: ip,
					},
				},
			},
		},
	}
}

func TestWatchSource_Prepare(t *testing.T) {
	// Check default values for optional config.
	config := &supervisor.WatchSourceConfig{}
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Prepare should NOT fail: %v", errs)
	}
	if config.TimeoutSecond != supervisor.DefaultWatchTimeoutSec {
		t.Fatalf("Default timeout should be %d, but got %d", supervisor.DefaultWatchTimeoutSec, config.TimeoutSecond)
	}
}

func TestWatchSource_Run(t *testing.T) {
	// Set up required config and state for running the step.
	config := &supervisor.WatchSourceConfig{
		TimeoutSecond: 60,
	}
	step := &supervisor.StepWatchSource{
		Config: config,
	}
	testNamespace := "test-ns"
	testSourceName := "test-source"
	testVMIP := "1.2.3.4"
	testIngressIP := "5.6.7.8"
	fakeLBServiceObj := newFakeLBServiceObj(testNamespace, testSourceName, testIngressIP)
	fakeKubClientSet := kubefake.NewSimpleClientset(fakeLBServiceObj)
	fakeVMUnstructured := newUnstructuredVM(testNamespace, testSourceName, testVMIP)
	fakeDynamicClient := daynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), fakeVMUnstructured)
	testWriter := new(bytes.Buffer)
	state := newBasicTestState(testWriter)
	state.Put(supervisor.StateKeyKubeClientSet, fakeKubClientSet)
	state.Put(supervisor.StateKeyKubeDynamicClient, fakeDynamicClient)
	state.Put(supervisor.StateKeyK8sNamespace, testNamespace)
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
			t.Errorf("State '%s' should be '%s', but got '%s'", supervisor.StateKeyConnectIP, testVMIP, vmIP)
		}
		ingressIP := state.Get(supervisor.StateKeyConnectIP)
		if ingressIP != testIngressIP {
			t.Errorf("State '%s' should be '%s', but got '%s'", supervisor.StateKeyConnectIP, testIngressIP, ingressIP)
		}

		// Check the output lines from the step runs.
		expectedOutput := []string{
			"Waiting for the source VM to be up and ready...",
			"Establishing a watch to the source VM object",
			"Source VM is NOT powered on yet, continue watching",
			"Source VM is powered on, waiting for an IP to be assigned",
			fmt.Sprintf("Successfully get the source VM IP: %s", testVMIP),
			"Getting source VM ingress IP from its K8s Service object",
			fmt.Sprintf("Successfully get the source VM ingress IP: %s", testIngressIP),
			"Source VM is now up and ready for customization",
		}
		checkOutputLines(t, testWriter, expectedOutput)
	}()

	// Wait for the watch to be established before updating the fake VM resource below.
	for i := 0; i < int(step.Config.TimeoutSecond); i++ {
		supervisor.Mu.Lock()
		if supervisor.IsWatchingVM {
			supervisor.Mu.Unlock()
			break
		}
		supervisor.Mu.Unlock()
		time.Sleep(time.Second)
	}

	// Update the VM resource with powerOff => powerOn => VM IP assigned.
	// In this way, we can test out the watch functionality and output messages.
	ctx := context.TODO()
	opt := metav1.UpdateOptions{}
	vmResource := schema.GroupVersionResource{
		Group:    "vmoperator.vmware.com",
		Version:  "v1alpha1",
		Resource: "virtualmachines",
	}
	fakeVMUnstructured.Object["status"] = map[string]interface{}{
		"powerState": "poweredOff",
	}
	fakeDynamicClient.Resource(vmResource).Namespace(testNamespace).Update(ctx, fakeVMUnstructured, opt)

	fakeVMUnstructured.Object["status"] = map[string]interface{}{
		"powerState": "poweredOn",
	}
	fakeDynamicClient.Resource(vmResource).Namespace(testNamespace).Update(ctx, fakeVMUnstructured, opt)

	fakeVMUnstructured.Object["status"] = map[string]interface{}{
		"vmIp": testVMIP,
	}
	fakeDynamicClient.Resource(vmResource).Namespace(testNamespace).Update(ctx, fakeVMUnstructured, opt)

	wg.Wait()
}
