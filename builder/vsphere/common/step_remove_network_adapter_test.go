// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func TestStepRemoveNetworkAdapter_Run(t *testing.T) {
	tc := []struct {
		name           string
		step           *StepRemoveNetworkAdapter
		expectedAction multistep.StepAction
		vmMock         *driver.VirtualMachineMock
		expectedVmMock *driver.VirtualMachineMock
		errMessage     string
	}{
		{
			name: "Successfully remove network adapters.",
			step: &StepRemoveNetworkAdapter{
				Config: &RemoveNetworkAdapterConfig{
					RemoveNetworkAdapter: true,
				},
			},
			expectedAction: multistep.ActionContinue,
			vmMock: &driver.VirtualMachineMock{
				RemoveNetworkAdaptersCalled: true,
			},
			expectedVmMock: &driver.VirtualMachineMock{
				RemoveNetworkAdaptersCalled: true,
			},
		},
		{
			name: "Fail to remove network adapters.",
			step: &StepRemoveNetworkAdapter{
				Config: &RemoveNetworkAdapterConfig{
					RemoveNetworkAdapter: true,
				},
			},
			expectedAction: multistep.ActionHalt,
			vmMock: &driver.VirtualMachineMock{
				RemoveNetworkAdaptersCalled: true,
				RemoveNetworkAdaptersErr:    fmt.Errorf("failed to remove network adapters"),
			},
			expectedVmMock: &driver.VirtualMachineMock{
				RemoveNetworkAdaptersCalled: true,
			},
			errMessage: "error removing network adapters: failed to remove network adapters",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			state := basicStateBag(nil)
			state.Put("vm", c.vmMock)

			if action := c.step.Run(context.TODO(), state); action != c.expectedAction {
				t.Fatalf("unexpected action %v", action)
			}
			err, ok := state.Get("error").(error)
			if ok {
				if err.Error() != c.errMessage {
					t.Fatalf("unexpected error %s", err.Error())
				}
			} else if c.errMessage != "" {
				t.Fatalf("expected to fail with %s but it didn't", c.errMessage)
			}

			if diff := cmp.Diff(c.vmMock, c.expectedVmMock,
				cmpopts.IgnoreInterfaces(struct{ error }{})); diff != "" {
				t.Fatalf("unexpected VirtualMachine calls: %s", diff)
			}
		})
	}
}
