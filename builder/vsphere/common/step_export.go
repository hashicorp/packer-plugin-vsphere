// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ExportConfig

package common

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// You can export an image in Open Virtualization Format (OVF) to the Packer host.
//
// Example usage:
//
// In JSON:
// ```json
// ...
//
//	"vm_name": "example-ubuntu",
//
// ...
//
//	"export": {
//	  "force": true,
//	  "output_directory": "./output-artifacts"
//	},
//
// ```
// In HCL2:
// ```hcl
//
//	# ...
//	vm_name = "example-ubuntu"
//	# ...
//	export {
//	  force = true
//	  output_directory = "./output-artifacts"
//	}
//
// ```
// The above configuration would create the following files:
//
// ```text
// ./output-artifacts/example-ubuntu-disk-0.vmdk
// ./output-artifacts/example-ubuntu.mf
// ./output-artifacts/example-ubuntu.ovf
// ```
type ExportConfig struct {
	// Name of the exported image in Open Virtualization Format (OVF).
	// The name of the virtual machine with the `.ovf` extension is used if this option is not specified.
	Name string `mapstructure:"name"`
	// Forces the export to overwrite existing files. Defaults to false.
	// If set to false, the export will fail if the files already exists.
	Force bool `mapstructure:"force"`
	// Include additional image files that are that are associated with the virtual machine. Defaults to false.
	// For example, `.nvram` and `.log` files.
	ImageFiles bool `mapstructure:"image_files"`
	// Generate a manifest file with the specified hash algorithm. Defaults to `sha256`.
	// Available options include `none`, `sha1`, `sha256`, and `sha512`. Use `none` for no manifest.
	Manifest string `mapstructure:"manifest"`
	// Path to the directory where the exported image will be saved.
	OutputDir OutputConfig `mapstructure:",squash"`
	// Advanced image export options. Options can include:
	// * mac - MAC address is exported for each Ethernet device.
	// * uuid - UUID is exported for the virtual machine.
	// * extraconfig - Extra configuration options are exported for the virtual machine.
	// * nodevicesubtypes - Resource subtypes for CD/DVD drives, floppy drives, and serial and parallel ports are not exported.
	//
	// For example, adding the following export config option outputs the MAC addresses for each Ethernet device in the OVF descriptor:
	//
	// In JSON:
	// ```json
	// ...
	//   "export": {
	//     "options": ["mac"]
	//   },
	// ```
	// In HCL2:
	// ```hcl
	// ...
	//   export {
	//     options = ["mac"]
	//   }
	// ```
	Options []string `mapstructure:"options"`
}

// Supported hash algorithms.
var sha = map[string]func() hash.Hash{
	"none":   nil,
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"sha512": sha512.New,
}

func (c *ExportConfig) Prepare(ctx *interpolate.Context, lc *LocationConfig, pc *common.PackerConfig) []error {
	var errs *packersdk.MultiError

	errs = packersdk.MultiErrorAppend(errs, c.OutputDir.Prepare(ctx, pc)...)

	// Check if the hash algorithm is supported.
	switch c.Manifest {
	case "":
		c.Manifest = "sha256"
	case "none", "sha1", "sha256", "sha512":
		// Supported hash algorithms; do nothing.
	default:
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("unsupported hash: %s. available options include 'none', 'sha1', 'sha256', and 'sha512'", c.Manifest))
	}

	// Default the name to the name of the virtual machine if not specified.
	if c.Name == "" {
		c.Name = lc.VMName
	}
	target := getTarget(c.OutputDir.OutputDir, c.Name)
	if !c.Force {
		if _, err := os.Stat(target); err == nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("file already exists: %s", target))
		}
	}

	if err := os.MkdirAll(c.OutputDir.OutputDir, c.OutputDir.DirPerm); err != nil {
		errs = packersdk.MultiErrorAppend(errs, errors.Wrap(err, "unable to make directory for export"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs.Errors
	}

	return nil
}

// Returns the target path for the exported image in Open Virtualization Format (OVF).
func getTarget(dir string, name string) string {
	return filepath.Join(dir, name+".ovf")
}

type StepExport struct {
	Name       string
	Force      bool
	ImageFiles bool
	Manifest   string
	OutputDir  string
	Options    []string
	mf         bytes.Buffer
}

func (s *StepExport) Cleanup(multistep.StateBag) {
}

func (s *StepExport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)

	// Start exporting the virtual machine image to Open Virtualization Format (OVF).
	ui.Say("Exporting to Open Virtualization Format (OVF)...")
	lease, err := vm.Export()
	if err != nil {
		state.Put("error", errors.Wrap(err, "error exporting virtual machine"))
		return multistep.ActionHalt
	}

	info, err := lease.Wait(ctx, nil)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	u := lease.StartUpdater(ctx, info)
	defer u.Done()

	cdp := types.OvfCreateDescriptorParams{
		Name: s.Name,
	}

	m := vm.NewOvfManager()
	if len(s.Options) > 0 {
		exportOptions, err := vm.GetOvfExportOptions(m)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		var unknown []string
		for _, option := range s.Options {
			found := false
			for _, exportOpt := range exportOptions {
				if exportOpt.Option == option {
					found = true
					break
				}
			}
			if !found {
				unknown = append(unknown, option)
			}
			cdp.ExportOption = append(cdp.ExportOption, option)
		}

		// Only print error message. The unknown options are ignored by vCenter Server.
		if len(unknown) > 0 {
			ui.Error(fmt.Sprintf("unknown export options %s", strings.Join(unknown, ",")))
		}
	}

	for _, i := range info.Items {
		if !s.include(&i) {
			continue
		}

		if !strings.HasPrefix(i.Path, s.Name) {
			i.Path = s.Name + "-" + i.Path
		}

		file := i.File()

		// Download the virtual machine image in Open Virtualization Format (OVF).
		ui.Say(fmt.Sprintf("Downloading %s...", file.Path))
		size, err := s.Download(ctx, lease, i)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		// Set the file size in the Open Virtualization Format descriptor.
		file.Size = size

		// Export the virtual machine image in Open Virtualization Format (OVF).
		ui.Say(fmt.Sprintf("Exporting %s...", file.Path))
		cdp.OvfFiles = append(cdp.OvfFiles, file)
	}

	if err = lease.Complete(ctx); err != nil {
		state.Put("error", errors.Wrap(err, "unable to complete lease"))
		return multistep.ActionHalt
	}

	desc, err := vm.CreateDescriptor(m, cdp)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to create descriptor"))
		return multistep.ActionHalt
	}

	target := getTarget(s.OutputDir, s.Name)
	file, err := os.Create(target)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to create file"))
		return multistep.ActionHalt
	}

	var w io.Writer = file
	h, ok := s.newHash()
	if ok {
		w = io.MultiWriter(file, h)
	}

	// Write the Open Virtualization Format descriptor.
	ui.Say(fmt.Sprintf("Writing OVF descriptor %s...", s.Name+".ovf"))
	_, err = io.WriteString(w, desc.OvfDescriptor)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to write ovf descriptor"))
		return multistep.ActionHalt
	}

	if err = file.Close(); err != nil {
		state.Put("error", errors.Wrap(err, "unable to close ovf descriptor"))
		return multistep.ActionHalt
	}

	// Manifest file will not be created. Continue to the next step.
	if s.Manifest == "none" {
		return multistep.ActionContinue
	}

	// Create a manifest file with the specified hash algorithm.
	ui.Say(fmt.Sprintf("Creating %s manifest %s...", strings.ToUpper(s.Manifest), s.Name+".mf"))
	s.addHash(filepath.Base(target), h)

	file, err = os.Create(filepath.Join(s.OutputDir, s.Name+".mf"))
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to create manifest"))
		return multistep.ActionHalt
	}

	_, err = io.Copy(file, &s.mf)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to write to manifest"))
		return multistep.ActionHalt
	}

	err = file.Close()
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to close the manifest"))
		return multistep.ActionHalt
	}

	// Completed exporting the virtual machine image to Open Virtualization Format (OVF).
	ui.Say("Completed export to Open Virtualization Format (OVF).")
	return multistep.ActionContinue
}

func (s *StepExport) include(item *nfc.FileItem) bool {
	if s.ImageFiles {
		return true
	}
	return filepath.Ext(item.Path) == ".vmdk"
}

func (s *StepExport) newHash() (hash.Hash, bool) {
	// Check if the hash function is nil to handle the 'none' case.
	if h, ok := sha[s.Manifest]; ok && h != nil {
		return h(), true
	}
	return nil, false
}

func (s *StepExport) addHash(p string, h hash.Hash) {
	_, _ = fmt.Fprintf(&s.mf, "%s(%s)= %x\n", strings.ToUpper(s.Manifest), p, h.Sum(nil))
}

func (s *StepExport) Download(ctx context.Context, lease *nfc.Lease, item nfc.FileItem) (int64, error) {
	path := filepath.Join(s.OutputDir, item.Path)
	opts := soap.Download{}

	if h, ok := s.newHash(); ok {
		opts.Writer = h
		defer s.addHash(item.Path, h)
	}

	err := lease.DownloadFile(ctx, path, item, opts)
	if err != nil {
		return 0, err
	}

	f, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return f.Size(), err
}
