// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type RemoveNetworkAdapterConfig

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type RemoveNetworkAdapterConfig struct {
	// Remove all network adapters from the virtual machine image. Defaults to `false`.
	RemoveNetworkAdapter bool `mapstructure:"remove_network_adapter"`
}

type StepRemoveNetworkAdapter struct {
	Config *RemoveNetworkAdapterConfig
}

func (s *StepRemoveNetworkAdapter) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.Config.RemoveNetworkAdapter {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	ui.Say("Removing network adapters...")
	err := vm.RemoveNetworkAdapters()

	if err != nil {
		state.Put("error", fmt.Errorf("error removing network adapters: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRemoveNetworkAdapter) Cleanup(state multistep.StateBag) {
	// no cleanup
}
