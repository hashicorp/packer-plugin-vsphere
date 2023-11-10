// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type RemoveCDRomConfig

package common

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type RemoveCDRomConfig struct {
	// Remove CD-ROM devices from template. Defaults to `false`.
	RemoveCdrom bool `mapstructure:"remove_cdrom"`
	KeepOneCdrom bool `mapstructure:"keep_one_cdrom"`
}

type StepRemoveCDRom struct {
	Config *RemoveCDRomConfig
}

func (s *StepRemoveCDRom) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	ui.Say("Eject CD-ROM drives...")
	err := vm.EjectCdroms()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if s.Config.RemoveCdrom == true {
		ui.Say("Deleting CD-ROM drives...")
		err := vm.RemoveCdroms()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	if s.Config.KeepOneCdrom == true {
		if _, err := vm.FindSATAController(); err == driver.ErrNoSataController {
			ui.Say("Adding SATA controller...")
			if err := vm.AddSATAController(); err != nil {
				state.Put("error", err)
				return multistep.ActionHalt
			}
		}

		ui.Say("Adding SATA CD-ROM drive...")
		err2 := vm.AddCdrom("sata","[] /usr/lib/vmware/isoimages/windows.iso")
		if err2 != nil {
			state.Put("error", err2)
			return multistep.ActionHalt
		}

		ui.Say("Ejecting ISO on SATA CD-ROM drive...")
		err3 := vm.EjectCdroms()
		if err3 != nil {
			state.Put("error", err3)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepRemoveCDRom) Cleanup(state multistep.StateBag) {}
