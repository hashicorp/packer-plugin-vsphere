// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ContentLibraryDestinationConfig
package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/vapi/vcenter"
)

// Create a content library item in a content library whose content is a VM
// template or an OVF template created from the virtual machine image after
// the build is complete.
//
// The template is stored in an existing or newly created library item.
type ContentLibraryDestinationConfig struct {
	// The name of the content library in which the new content library item
	// containing the template will be created or updated. The content library
	// must be of type Local to allow deploying virtual machines.
	Library string `mapstructure:"library"`
	// The name of the content library item that will be created or updated.
	// For VM templates, the name of the item should be different from
	// [vm_name](#vm_name) and the default is [vm_name](#vm_name) + timestamp
	// when not set. VM templates will always be imported to a new library item.
	// For OVF templates, the name defaults to [vm_name](#vm_name) when not set,
	// and if an item with the same name already exists it will be then updated
	// with the new OVF template, otherwise a new item will be created.
	//
	// ~> **Note:** It's not possible to update existing content library items
	// with a new VM template. If updating an existing content library item is
	// necessary, use an OVF template instead by setting the [ovf](#ovf) option
	// as `true`.
	Name string `mapstructure:"name"`
	// A description for the content library item that will be created.
	// Defaults to "Packer imported [vm_name](#vm_name) VM template".
	Description string `mapstructure:"description"`
	// The cluster where the VM template will be placed.
	// If `cluster` and `resource_pool` are both specified, `resource_pool` must
	// belong to cluster. If `cluster` and `host` are both specified, the ESXi
	// host must be a member of the cluster. This option is not used when
	// importing OVF templates. Defaults to [`cluster`](#cluster).
	Cluster string `mapstructure:"cluster"`
	// The virtual machine folder where the VM template will be placed.
	// This option is not used when importing OVF templates. Defaults to
	// the same folder as the source virtual machine.
	Folder string `mapstructure:"folder"`
	// The ESXi host where the virtual machine template will be placed.
	// If `host` and `resource_pool` are both specified, `resource_pool` must
	// belong to host. If `host` and `cluster` are both specified, `host` must
	// be a member of the cluster. This option is not used when importing OVF
	// templates. Defaults to [`host`](#host).
	Host string `mapstructure:"host"`
	// The resource pool where the virtual machine template will be placed.
	// Defaults to [`resource_pool`](#resource_pool). If [`resource_pool`](#resource_pool)
	// is unset, the system will attempt to choose a suitable resource pool
	// for the VM template.
	ResourcePool string `mapstructure:"resource_pool"`
	// The datastore for the virtual machine template's configuration and log
	// files. This option is not used when importing OVF templates.
	// Defaults to the storage backing associated with the content library.
	Datastore string `mapstructure:"datastore"`
	// Destroy the virtual machine after the import to the content library.
	// Defaults to `false`.
	Destroy bool `mapstructure:"destroy"`
	// Import an OVF template to the content library item. Defaults to `false`.
	Ovf bool `mapstructure:"ovf"`
	// Skip the import to the content library item. Useful during a build test
	// stage. Defaults to `false`.
	SkipImport bool `mapstructure:"skip_import"`
	// Flags to use for OVF package creation. The supported flags can be
	// obtained using ExportFlag.list. If unset, no flags will be used.
	// Known values: `EXTRA_CONFIG`, `PRESERVE_MAC`.
	OvfFlags []string `mapstructure:"ovf_flags"`
}

func (c *ContentLibraryDestinationConfig) Prepare(lc *LocationConfig) []error {
	var errs *packersdk.MultiError

	if c.Library == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("a library name must be provided"))
	}

	if c.Ovf {
		if c.Name == "" {
			c.Name = lc.VMName
		}
	} else {
		if c.Name == lc.VMName {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("the content library destination name must be different from the VM name"))
		}

		if c.Name == "" {
			// Add timestamp to the name to differentiate from the original VM
			// otherwise vSphere won't be able to create the template which will be imported
			name, err := interpolate.Render(lc.VMName+"{{timestamp}}", nil)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs,
					fmt.Errorf("unable to parse content library VM template name: %s", err))
			}
			c.Name = name
		}
		if c.Cluster == "" {
			c.Cluster = lc.Cluster
		}
		if c.Host == "" {
			c.Host = lc.Host
		}
		if c.ResourcePool == "" {
			c.ResourcePool = lc.ResourcePool
		}
	}
	if c.Description == "" {
		c.Description = fmt.Sprintf("Packer imported %s VM template", lc.VMName)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs.Errors
	}

	return nil
}

type StepImportToContentLibrary struct {
	ContentLibConfig *ContentLibraryDestinationConfig
}

func (s *StepImportToContentLibrary) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	if s.ContentLibConfig.SkipImport {
		ui.Say("Skipping import...")
		return multistep.ActionContinue
	}

	vm := state.Get("vm").(*driver.VirtualMachineDriver)
	var err error

	ui.Say("Clearing boot order...")
	err = vm.SetBootOrder([]string{"-"})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	vmTypeLabel := "VM"
	if s.ContentLibConfig.Ovf {
		vmTypeLabel = "VM OVF"
	}
	ui.Sayf("Importing %s template %s to Content Library '%s' as the item '%s' with the description '%s'...",
		vmTypeLabel, s.ContentLibConfig.Name, s.ContentLibConfig.Library, s.ContentLibConfig.Name, s.ContentLibConfig.Description)

	if s.ContentLibConfig.Ovf {
		err = s.importOvfTemplate(vm)
	} else {
		err = s.importVmTemplate(vm)
	}

	if err != nil {
		ui.Errorf("Failed to import template %s: %s", s.ContentLibConfig.Name, err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Add a tracer to the state to track if the Destroy parameter was used.
	if s.ContentLibConfig.Destroy {
		state.Put("destroy_vm", s.ContentLibConfig.Destroy)
	}

	// For HCP Packer metadata, save the content library item UUID in state.
	itemUuid, err := vm.FindContentLibraryItemUUID(s.ContentLibConfig.Library, s.ContentLibConfig.Name)
	if err != nil {
		ui.Errorf("Failed to get content library item uuid: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	} else {
		state.Put("content_library_item_uuid", itemUuid)
	}

	// For HCP Packer metadata, save the content library datastore name in state.
	datastores, err := vm.FindContentLibraryTemplateDatastoreName(s.ContentLibConfig.Library)
	if err != nil {
		ui.Errorf("Failed to get content library datastore name: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	} else {
		state.Put("content_library_datastore", datastores)
	}

	return multistep.ActionContinue
}

func (s *StepImportToContentLibrary) importOvfTemplate(vm *driver.VirtualMachineDriver) error {
	ovf := vcenter.OVF{
		Spec: vcenter.CreateSpec{
			Name:        s.ContentLibConfig.Name,
			Description: s.ContentLibConfig.Description,
			Flags:       s.ContentLibConfig.OvfFlags,
		},
		Target: vcenter.LibraryTarget{
			LibraryID: s.ContentLibConfig.Library,
		},
	}
	return vm.ImportOvfToContentLibrary(ovf)
}

func (s *StepImportToContentLibrary) importVmTemplate(vm *driver.VirtualMachineDriver) error {
	template := vcenter.Template{
		Name:        s.ContentLibConfig.Name,
		Description: s.ContentLibConfig.Description,
		Library:     s.ContentLibConfig.Library,
		Placement: &vcenter.Placement{
			Cluster:      s.ContentLibConfig.Cluster,
			Folder:       s.ContentLibConfig.Folder,
			Host:         s.ContentLibConfig.Host,
			ResourcePool: s.ContentLibConfig.ResourcePool,
		},
	}

	if s.ContentLibConfig.Datastore != "" {
		template.VMHomeStorage = &vcenter.DiskStorage{
			Datastore: s.ContentLibConfig.Datastore,
		}
	}

	return vm.ImportToContentLibrary(template)
}

func (s *StepImportToContentLibrary) Cleanup(multistep.StateBag) {
}
