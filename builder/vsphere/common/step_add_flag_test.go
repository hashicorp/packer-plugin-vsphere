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

func TestFlagConfig_Prepare(t *testing.T) {
	tc := []struct {
		name           string
		config         *FlagConfig
		hardwareConfig *HardwareConfig
		fail           bool
		expectedErrMsg string
	}{
		{
			name:           "Should not fail for empty config",
			config:         new(FlagConfig),
			hardwareConfig: new(HardwareConfig),
			fail:           false,
			expectedErrMsg: "",
		},
		{
			name: "VbsEnabled but VvtdEnabled not set",
			config: &FlagConfig{
				VbsEnabled: true,
			},
			hardwareConfig: &HardwareConfig{
				Firmware: "efi-secure",
				NestedHV: true,
			},
			fail:           true,
			expectedErrMsg: "`vvtd_enabled` must be set to `true` when `vbs_enabled` is set to `true`",
		},
		{
			name: "VbsEnabled but NestedHV not set",
			config: &FlagConfig{
				VbsEnabled:  true,
				VvtdEnabled: true,
			},
			hardwareConfig: &HardwareConfig{
				Firmware: "efi-secure",
			},
			fail:           true,
			expectedErrMsg: "`nestedhv` must be set to `true` when `vbs_enabled` is set to `true`",
		},
		{
			name: "VbsEnabled but Firmware not set to efi-secure",
			config: &FlagConfig{
				VbsEnabled:  true,
				VvtdEnabled: true,
			},
			hardwareConfig: &HardwareConfig{
				NestedHV: true,
				Firmware: "efi",
			},
			fail:           true,
			expectedErrMsg: "`firmware` must be set to `efi-secure` when `vbs_enabled` is set to `true`",
		},
		{
			name: "VbsEnabled and all required fields set",
			config: &FlagConfig{
				VbsEnabled:  true,
				VvtdEnabled: true,
			},
			hardwareConfig: &HardwareConfig{
				NestedHV: true,
				Firmware: "efi-secure",
			},
			fail:           false,
			expectedErrMsg: "",
		},
	}

	for _, c := range tc {
		errs := c.config.Prepare(c.hardwareConfig)
		if c.fail {
			if len(errs) == 0 {
				t.Fatalf("Config prepare should fail")
			}
			if errs[0].Error() != c.expectedErrMsg {
				t.Fatalf("Expected error message: %s but was '%s'", c.expectedErrMsg, errs[0].Error())
			}
		} else {
			if len(errs) != 0 {
				t.Fatalf("Config prepare should not fail")
			}
		}
	}
}

func TestStepAddFlag_Run(t *testing.T) {
	tc := []struct {
		name           string
		state          *multistep.BasicStateBag
		step           *StepAddFlag
		vmMock         *driver.VirtualMachineMock
		expectedAction multistep.StepAction
		expectedVmMock *driver.VirtualMachineMock
		fail           bool
		errMessage     string
	}{
		{
			name:  "Add Flag",
			state: basicStateBag(nil),
			step: &StepAddFlag{
				FlagConfig: FlagConfig{
					VbsEnabled:  true,
					VvtdEnabled: true,
				},
			},
			vmMock:         new(driver.VirtualMachineMock),
			expectedAction: multistep.ActionContinue,
			expectedVmMock: &driver.VirtualMachineMock{
				AddFlagCalled:            true,
				AddFlagCalledTimes:       1,
				AddFlagVbsEnabledValues:  true,
				AddFlagVvtdEnabledValues: true,
			},
			fail:       false,
			errMessage: "",
		},
		{
			name:  "Fail to add flag",
			state: basicStateBag(nil),
			step: &StepAddFlag{
				FlagConfig: FlagConfig{
					VbsEnabled:  true,
					VvtdEnabled: false,
				},
			},
			vmMock: &driver.VirtualMachineMock{
				AddFlagErr: fmt.Errorf("AddFlag error"),
			},
			expectedAction: multistep.ActionHalt,
			expectedVmMock: &driver.VirtualMachineMock{
				AddFlagCalled:            true,
				AddFlagCalledTimes:       1,
				AddFlagVbsEnabledValues:  true,
				AddFlagVvtdEnabledValues: false,
			},
			fail:       true,
			errMessage: fmt.Sprintf("error adding virtual machine flag: %v", fmt.Errorf("AddFlag error")),
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			c.state.Put("vm", c.vmMock)
			if action := c.step.Run(context.TODO(), c.state); action != c.expectedAction {
				t.Fatalf("unexpected action %v", action)
			}
			err, ok := c.state.Get("error").(error)
			if ok {
				if err.Error() != c.errMessage {
					t.Fatalf("unexpected error %s", err.Error())
				}
			} else {
				if c.fail {
					t.Fatalf("expected to fail but it didn't")
				}
			}

			if diff := cmp.Diff(c.vmMock, c.expectedVmMock,
				cmpopts.IgnoreInterfaces(struct{ error }{})); diff != "" {
				t.Fatalf("unexpected VirtualMachine calls: %s", diff)
			}
		})
	}
}
