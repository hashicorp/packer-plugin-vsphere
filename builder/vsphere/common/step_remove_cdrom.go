// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type RemoveCDRomConfig

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type RemoveCDRomConfig struct {
	// Remove CD-ROM devices from template. Defaults to `false`.
	RemoveCdrom bool `mapstructure:"remove_cdrom"`
}

type StepRemoveCDRom struct {
	Config *RemoveCDRomConfig
}

func (s *StepRemoveCDRom) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	// Eject media from CD-ROM devices.
	ui.Say("Ejecting CD-ROM media...")
	err := vm.EjectCdroms()
	if err != nil {
		state.Put("error", fmt.Errorf("error ejecting cdrom media: %v", err))
		return multistep.ActionHalt
	}

	// Remove all CD-ROM devices from the image.
	if s.Config.RemoveCdrom == true {
		ui.Say("Removing CD-ROM devices...")
		err := vm.RemoveCdroms()
		if err != nil {
			state.Put("error", fmt.Errorf("error removing cdrom: %v", err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepRemoveCDRom) Cleanup(state multistep.StateBag) {
	// no cleanup
}
