// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CDRomConfig

package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type CDRomConfig struct {
	// The type of controller to use for the CD-ROM device. Defaults to `ide`.
	//
	// The available options for this setting are: `ide` and `sata`.
	CdromType string `mapstructure:"cdrom_type"`
	// A list of paths to ISO files in either a datastore or a content library
	// that will be attached to the virtual machine.
	//
	// HCL Example:
	//
	// ```hcl
	// iso_paths = [
	//   "[nfs] iso/ubuntu-server-amd64.iso",
	//   "Example Content Library/ubuntu-server-amd64/ubuntu-server-amd64.iso"
	// ]
	// ```
	//
	// JSON Example:
	//
	// ```json
	// "iso_paths": [
	//   "[nfs] iso/ubuntu-server-amd64.iso",
	//   "Example Content Library/ubuntu-server-amd64/ubuntu-server-amd64.iso"
	// ]
	// ```
	//
	// Two ISOs are referenced:
	//
	// 1. An ISO in the "_iso_" folder of the "_nfs_" datastore with the file
	//   name of "_ubuntu-server-amd64.iso_". "_ubuntu-server-amd64.iso_".
	// 2. An ISO in the "_Example Content Library_" content library with the
	//   item name of "_ubuntu-server-amd64_".
	//
	// -> **Note:** All files in a content library have an associated item name.
	// To determine the file name, view the datastore backing the content
	// library or use the `govc` vSphere CLI.
	ISOPaths []string `mapstructure:"iso_paths"`
}

type StepAddCDRom struct {
	Config *CDRomConfig
}

func (c *CDRomConfig) Prepare(k *ReattachCDRomConfig) []error {
	var errs []error

	// `cdrom_type` must be either 'ide' or 'sata'.
	if c.CdromType != "" && c.CdromType != "ide" && c.CdromType != "sata" {
		errs = append(errs, fmt.Errorf("'cdrom_type' must be 'ide' or 'sata'"))
	}

	// `reattach_cdroms` should be between 1 and 4 to keep the CD-ROM devices
	// without any attached media. If `reattach_cdroms` is set to 0, it is
	// ignored and the step is skipped.
	if k.ReattachCDRom < 0 || k.ReattachCDRom > 4 {
		errs = append(errs, fmt.Errorf("'reattach_cdroms' should be between 1 and 4,\n"+
			"  if set to 0, `reattach_cdroms` is ignored and the step is skipped"))
	}
	return errs
}

func (s *StepAddCDRom) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	if s.Config.CdromType == "sata" {
		if _, err := vm.FindSATAController(); err != nil {
			if !errors.Is(err, driver.ErrNoSataController) {
				state.Put("error", fmt.Errorf("unexpected error finding SATA controller: %w", err))
				return multistep.ActionHalt
			}

			ui.Say("Adding SATA controller...")
			if err := vm.AddSATAController(); err != nil {
				state.Put("error", fmt.Errorf("error adding SATA controller: %w", err))
				return multistep.ActionHalt
			}
		}
	}

	if path, ok := state.GetOk("iso_remote_path"); ok {
		// The order matters: docs say "iso_url" should go first, so make sure
		// to prepend it.
		s.Config.ISOPaths = append([]string{path.(string)}, s.Config.ISOPaths...)
	}

	// Add our custom CD, if it exists.
	if cdPath, _ := state.Get("cd_path").(string); cdPath != "" {
		s.Config.ISOPaths = append(s.Config.ISOPaths, cdPath)
	}

	ui.Say("Mounting ISO images...")
	// Due to a limitation in govmomi, creating multiple CD-ROMs simultaneously
	// and then mounting ISO files can lead to incorrect UnitNumbers.
	// To avoid this issue, create and mount them individually.
	for _, path := range s.Config.ISOPaths {
		if path == "" {
			state.Put("error", fmt.Errorf("invalid path: empty string"))
			return multistep.ActionHalt
		}
		if err := vm.AddCdrom(s.Config.CdromType, path); err != nil {
			state.Put("error", fmt.Errorf("error mounting an image '%v': %v", path, err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepAddCDRom) Cleanup(state multistep.StateBag) {}
