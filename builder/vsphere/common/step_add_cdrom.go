// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CDRomConfig

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type CDRomConfig struct {
	// Which controller to use. Example: `sata`. Defaults to `ide`.
	CdromType string `mapstructure:"cdrom_type"`
	// A list of paths to ISO files in either a datastore or a content library that will be mounted to the VM.
	//
	// Usage example (HCL):
	//
	// ```hcl
	// iso_paths = [
	//   "[nfs] iso/ubuntu-server-amd64.iso",
	//   "Packer/ubuntu-server-amd64/ubuntu-server-amd64.iso"
	// ]
	// ```
	//
	// Two ISOs are referenced:
	// 1. An ISO in the "_iso_" folder of the "_nfs_" datastore with the file name of "_ubuntu-server-amd64.iso_".
	// 2. An ISO in the "_Packer_" content library with the item name of "_ubuntu-server-amd64_".
	//
	// -> **Note:** All files in a content library have an associated item name.
	// To determine the file name, view the datastore backing the content library or use the `govc` vSphere CLI.
	ISOPaths []string `mapstructure:"iso_paths"`
}

type StepAddCDRom struct {
	Config *CDRomConfig
}

func (c *CDRomConfig) Prepare() []error {
	var errs []error

	if c.CdromType != "" && c.CdromType != "ide" && c.CdromType != "sata" {
		errs = append(errs, fmt.Errorf("'cdrom_type' must be 'ide' or 'sata'"))
	}

	return errs
}

func (s *StepAddCDRom) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	if s.Config.CdromType == "sata" {
		if _, err := vm.FindSATAController(); err == driver.ErrNoSataController {
			ui.Say("Adding SATA controller...")
			if err := vm.AddSATAController(); err != nil {
				state.Put("error", fmt.Errorf("error adding SATA controller: %v", err))
				return multistep.ActionHalt
			}
		}
	}

	ui.Say("Mounting ISO images...")
	if path, ok := state.GetOk("iso_remote_path"); ok {
		if err := vm.AddCdrom(s.Config.CdromType, path.(string)); err != nil {
			state.Put("error", fmt.Errorf("error mounting an image '%v': %v", path, err))
			return multistep.ActionHalt
		}
	}

	// Add our custom CD, if it exists
	if cd_path, _ := state.Get("cd_path").(string); cd_path != "" {
		s.Config.ISOPaths = append(s.Config.ISOPaths, cd_path)
	}

	if len(s.Config.ISOPaths) > 0 {
		for _, path := range s.Config.ISOPaths {
			if err := vm.AddCdrom(s.Config.CdromType, path); err != nil {
				state.Put("error", fmt.Errorf("error mounting an image '%v': %v", path, err))
				return multistep.ActionHalt
			}
		}
	}
	return multistep.ActionContinue
}

func (s *StepAddCDRom) Cleanup(state multistep.StateBag) {}
