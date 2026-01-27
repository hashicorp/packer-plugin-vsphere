// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type FlagConfig

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/vim25/types"
)

type FlagConfig struct {
	// Enable Virtualization Based Security option for virtual machine. Defaults to `false`.
	// Requires `vvtd_enabled` and `NestedHV` to be set to `true`.
	// Requires `firmware` to be set to `efi-secure`.
	VbsEnabled bool `mapstructure:"vbs_enabled"`
	// Enable IO/MMU option for virtual machine. Defaults to `false`.
	VvtdEnabled bool `mapstructure:"vvtd_enabled"`
}

func (c *FlagConfig) Prepare(h *HardwareConfig) []error {
	var errs []error

	if h == nil {
		return append(errs, fmt.Errorf("no hardware config provided"))
	}

	if c.VbsEnabled {
		if !c.VvtdEnabled {
			errs = append(errs, fmt.Errorf("`vvtd_enabled` must be set to `true` when `vbs_enabled` is set to `true`"))
		}

		if !h.NestedHV {
			errs = append(errs, fmt.Errorf("`nestedhv` must be set to `true` when `vbs_enabled` is set to `true`"))
		}

		if h.Firmware != "efi-secure" {
			errs = append(errs, fmt.Errorf("`firmware` must be set to `efi-secure` when `vbs_enabled` is set to `true`"))
		}
	}

	return errs
}

type StepAddFlag struct {
	FlagConfig FlagConfig
}

func (s *StepAddFlag) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	var info *types.VirtualMachineFlagInfo

	if s.FlagConfig.VbsEnabled || s.FlagConfig.VvtdEnabled {
		info = &types.VirtualMachineFlagInfo{}

		if s.FlagConfig.VbsEnabled {
			info.VbsEnabled = &s.FlagConfig.VbsEnabled
		}

		if s.FlagConfig.VvtdEnabled {
			info.VvtdEnabled = &s.FlagConfig.VvtdEnabled
		}

		ui.Say("Adding virtual machine flags...")
		if err := vm.AddFlag(ctx, info); err != nil {
			state.Put("error", fmt.Errorf("error adding virtual machine flag: %v", err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepAddFlag) Cleanup(state multistep.StateBag) {
	// Nothing to clean up.
}
