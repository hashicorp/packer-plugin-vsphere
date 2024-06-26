// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type StepConvertToTemplate struct {
	ConvertToTemplate bool
}

func (s *StepConvertToTemplate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)

	if s.ConvertToTemplate {
		ui.Say("Converting virtual machine to template...")
		err := vm.ConvertToTemplate()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConvertToTemplate) Cleanup(state multistep.StateBag) {}
