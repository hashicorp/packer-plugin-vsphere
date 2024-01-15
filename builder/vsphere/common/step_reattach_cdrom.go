// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ReattachCDRomConfig

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type ReattachCDRomConfig struct {
	// Reattach one or more configured CD-ROM devices. Range: 1-4.
	// You can reattach up to 4 CD-ROM devices to the final build artifact.
	// If set to 0, `reattach_cdroms` is ignored and the step is skipped.
	// When set to a value in the range, `remove_cdrom` is ignored and
	// the CD-ROM devices are kept without any attached media.
	ReattachCDRom int `mapstructure:"reattach_cdroms"`
}

type StepReattachCDRom struct {
	Config      *ReattachCDRomConfig
	CDRomConfig *CDRomConfig
}

func (s *StepReattachCDRom) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	var err error

	// Check if `reattach_cdroms` is set.
	ReattachCDRom := s.Config.ReattachCDRom
	if ReattachCDRom == 0 {
		return multistep.ActionContinue
	}
	if ReattachCDRom < 1 || ReattachCDRom > 4 {
		err := fmt.Errorf("'reattach_cdroms' should be between 1 and 4. if set to 0, `reattach_cdroms` is ignored and the step is skipped")
		state.Put("error", fmt.Errorf("error reattach cdrom: %v", err))
		return multistep.ActionHalt
	}

	// Keep the CD-ROM devices on the image without any attached media.
	ui.Say("Reattaching CD-ROM devices...")
	// Remove all CD-ROM devices from the image.
	err = vm.RemoveCdroms()
	if err != nil {
		state.Put("error", fmt.Errorf("error removing cdrom prior to reattaching: %v", err))
		return multistep.ActionHalt
	}

	// If the CD-ROM device type is SATA, make sure SATA controller is present.
	if s.CDRomConfig.CdromType == "sata" {
		if _, err := vm.FindSATAController(); err == driver.ErrNoSataController {
			ui.Say("Adding SATA controller...")
			if err := vm.AddSATAController(); err != nil {
				state.Put("error", fmt.Errorf("error adding sata controller: %v", err))
				return multistep.ActionHalt
			}
		}
	}

	ui.Say("Adding CD-ROM device...")
	// Add the CD-ROM devices to the image based on the value of `reattach_cdroms`.
	// A valid ISO path is required for this step. The media will subsequently be ejected.
	for i := 0; i < ReattachCDRom; i++ {
		err = vm.AddCdrom(s.CDRomConfig.CdromType, s.CDRomConfig.ISOPaths[0])
		if err != nil {
			state.Put("error", fmt.Errorf("error adding cdrom %d: %v", i, err))
			return multistep.ActionHalt
		}
	}

	ui.Say("Ejecting CD-ROM media...")
	// Eject media from CD-ROM devices.
	err = vm.EjectCdroms()
	if err != nil {
		state.Put("error", fmt.Errorf("error ejecting cdrom media: %v", err))
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *StepReattachCDRom) Cleanup(state multistep.StateBag) {
	// no cleanup
}
