// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func CleanupVM(state multistep.StateBag) {
	st := state.Get("vm")
	if st == nil {
		return
	}
	vm := st.(driver.VirtualMachine)

	if vmDriver, ok := vm.(*driver.VirtualMachineDriver); ok {
		// Make sure we get VM metadata before destroying it
		state.Put("metadata", GetVMMetadata(vmDriver, state))
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	_, destroy := state.GetOk("destroy_vm")
	if !cancelled && !halted && !destroy {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Destroying VM...")
	err := vm.Destroy()
	if err != nil {
		ui.Errorf("%s", err)
	}
}
