// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CDRomConfig

package common

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/object"
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

// validateISOPaths validates that all ISO files specified in ISOPaths exist on their respective datastores.
func (c *CDRomConfig) validateISOPaths(d driver.Driver) []error {
	var errs []error

	if len(c.ISOPaths) == 0 {
		return errs
	}

	if d == nil {
		errs = append(errs, fmt.Errorf("driver is not available for ISO validation"))
		return errs
	}

	for _, isoPath := range c.ISOPaths {
		if strings.TrimSpace(isoPath) == "" {
			errs = append(errs, fmt.Errorf("ISO path cannot be empty or whitespace-only"))
			continue
		}

		if err := c.validateContentLibraryPath(isoPath, d); err == nil {
			continue
		} else {
			if strings.Contains(err.Error(), "content library") && !strings.Contains(err.Error(), "not a content library path format") {
				errs = append(errs, err)
				continue
			}
		}

		if err := c.validateDatastorePath(isoPath, d); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// validateDatastorePath validates a datastore ISO path format and file existence.
func (c *CDRomConfig) validateDatastorePath(isoPath string, d driver.Driver) error {
	dsIsoPath := driver.NewDatastoreIsoPath(isoPath)

	if !dsIsoPath.Validate() {
		return fmt.Errorf("invalid datastore path format: '%s'", isoPath)
	}

	dsPath := object.DatastorePath{}
	if !dsPath.FromString(isoPath) {
		return fmt.Errorf("unable to parse datastore path: '%s'", isoPath)
	}

	datastore, err := d.FindDatastore(dsPath.Datastore, "")
	if err != nil {
		return fmt.Errorf("unable to access datastore '%s' for ISO validation: %s", dsPath.Datastore, err)
	}

	if datastore == nil {
		return fmt.Errorf("datastore '%s' returned nil reference during ISO validation", dsPath.Datastore)
	}

	filePath := dsIsoPath.GetFilePath()
	if !datastore.FileExists(filePath) {
		return fmt.Errorf("ISO file not found: '%s'", isoPath)
	}

	return nil
}

// validateContentLibraryPath validates a content library ISO path format and file existence.
func (c *CDRomConfig) validateContentLibraryPath(isoPath string, d driver.Driver) error {
	pathParts := strings.Split(strings.TrimLeft(isoPath, "/"), "/")
	if len(pathParts) != 3 {
		return fmt.Errorf("not a content library path format")
	}
	for i, part := range pathParts {
		if strings.TrimSpace(part) == "" {
			switch i {
			case 0:
				return fmt.Errorf("content library name cannot be empty in path: '%s'", isoPath)
			case 1:
				return fmt.Errorf("content library item name cannot be empty in path: '%s'", isoPath)
			case 2:
				return fmt.Errorf("content library file name cannot be empty in path: '%s'", isoPath)
			}
		}
	}

	datastorePath, err := d.FindContentLibraryFileDatastorePath(isoPath)
	if err != nil {
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "not found") {
			if strings.Contains(errorMsg, "library") {
				return fmt.Errorf("content library not found: '%s'", pathParts[0])
			}
			if strings.Contains(errorMsg, "item") {
				return fmt.Errorf("content library item not found: '%s' in library '%s'", pathParts[1], pathParts[0])
			}
		}
		if strings.Contains(errorMsg, "not identified as a Content Library path") {
			return fmt.Errorf("invalid content library path format: '%s'", isoPath)
		}
		if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "network") {
			return fmt.Errorf("unable to connect to vCenter for content library validation: %s", err)
		}
		return fmt.Errorf("content library path resolution failed for '%s': %s", isoPath, err)
	}

	if strings.TrimSpace(datastorePath) == "" {
		return fmt.Errorf("content library path resolution returned empty datastore path for '%s'", isoPath)
	}

	if datastorePath != isoPath {
		if err := c.validateDatastorePath(datastorePath, d); err != nil {
			return fmt.Errorf("content library file validation failed for '%s' (resolved to '%s'): %s", isoPath, datastorePath, err)
		}
		return nil
	}

	return fmt.Errorf("content library path resolution did not return a datastore path for '%s'", isoPath)
}

func (c *CDRomConfig) Prepare(k *ReattachCDRomConfig, d driver.Driver) []error {
	var errs []error

	if len(c.ISOPaths) > 0 && d != nil {
		if validationErrs := c.validateISOPaths(d); len(validationErrs) > 0 {
			errs = append(errs, validationErrs...)
		}
	}

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
	if d, ok := state.GetOk("driver"); ok && len(s.Config.ISOPaths) > 0 {
		if driver, driverOk := d.(driver.Driver); driverOk {
			if validationErrs := s.Config.validateISOPaths(driver); len(validationErrs) > 0 {
				var errMsg string
				for i, err := range validationErrs {
					if i > 0 {
						errMsg += "; "
					}
					errMsg += err.Error()
				}
				state.Put("error", fmt.Errorf("ISO validation failed: %s", errMsg))
				return multistep.ActionHalt
			}
		}
	}

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
		s.Config.ISOPaths = append([]string{path.(string)}, s.Config.ISOPaths...)
	}

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
