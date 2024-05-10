// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CloneConfig,vAppConfig

package clone

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type vAppConfig struct {
	// The values for the available vApp properties. These are used to supply
	// configuration parameters to a virtual machine. This machine is cloned
	// from a template that originated from an imported OVF or OVA file.
	//
	// -> **Note:** The only supported usage path for vApp properties is for
	// existing user-configurable keys. These generally come from an existing
	// template that was created from an imported OVF or OVA file.
	//
	// You cannot set values for vApp properties on virtual machines created
	// from scratch, on virtual machines that lack a vApp configuration, or on
	// property keys that do not exist.
	Properties map[string]string `mapstructure:"properties"`
}

type CloneConfig struct {
	// The name of the source virtual machine to clone.
	Template string `mapstructure:"template"`
	// The size of the primary disk in MiB. Cannot be used with `linked_clone`.
	// -> **Note:** Only the primary disk size can be specified. Additional
	// disks are not supported.
	DiskSize int64 `mapstructure:"disk_size"`
	// Create the virtual machine as a linked clone from the latest snapshot.
	// Defaults to `false`. Cannot be used with `disk_size`.`
	LinkedClone bool `mapstructure:"linked_clone"`
	// The network to which the virtual machine will connect.
	//
	// For example:
	//
	// - Name: `<NetworkName>`
	// - Inventory Path: `/<DatacenterName>/<FolderName>/<NetworkName>`
	// - Managed Object ID (Port Group): `Network:network-<xxxxx>`
	// - Managed Object ID (Distributed Port Group): `DistributedVirtualPortgroup::dvportgroup-<xxxxx>`
	// - Logical Switch UUID: `<uuid>`
	// - Segment ID: `/infra/segments/<SegmentID>`
	//
	// ~> **Note:** If more than one network resolves to the same name, either
	// the inventory path to network or an ID must be provided.
	//
	// ~> **Note:** If no network is specified, provide `host` to allow the
	// plugin to search for an available network.
	Network string `mapstructure:"network"`
	// The network card MAC address. For example `00:50:56:00:00:00`.
	// If set, the `network` must be also specified.
	MacAddress string `mapstructure:"mac_address"`
	// The annotations for the virtual machine.
	Notes string `mapstructure:"notes"`
	// Destroy the virtual machine after the build is complete.
	// Defaults to `false`.
	Destroy bool `mapstructure:"destroy"`
	// The vApp Options for the virtual machine. For more information, refer to
	// the [vApp Options Configuration](/packer/plugins/builders/vmware/vsphere-clone#vapp-options-configuration)
	// section.
	VAppConfig    vAppConfig           `mapstructure:"vapp"`
	StorageConfig common.StorageConfig `mapstructure:",squash"`
}

func (c *CloneConfig) Prepare() []error {
	var errs []error
	errs = append(errs, c.StorageConfig.Prepare()...)

	if c.Template == "" {
		errs = append(errs, fmt.Errorf("'template' is required"))
	}

	if c.LinkedClone == true && c.DiskSize != 0 {
		errs = append(errs, fmt.Errorf("'linked_clone' and 'disk_size' cannot be used together"))
	}

	if c.MacAddress != "" && c.Network == "" {
		errs = append(errs, fmt.Errorf("'network' is required when 'mac_address' is specified"))
	}

	return errs
}

type StepCloneVM struct {
	Config        *CloneConfig
	Location      *common.LocationConfig
	Force         bool
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCloneVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	d := state.Get("driver").(driver.Driver)
	vmPath := path.Join(s.Location.Folder, s.Location.VMName)

	err := d.PreCleanVM(ui, vmPath, s.Force, s.Location.Cluster, s.Location.Host, s.Location.ResourcePool)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Cloning virtual machine...")
	template, err := d.FindVM(s.Config.Template)
	if err != nil {
		state.Put("error", fmt.Errorf("error finding virtual machine to clone: %s", err))
		return multistep.ActionHalt
	}

	var disks []driver.Disk
	for _, disk := range s.Config.StorageConfig.Storage {
		disks = append(disks, driver.Disk{
			DiskSize:            disk.DiskSize,
			DiskEagerlyScrub:    disk.DiskEagerlyScrub,
			DiskThinProvisioned: disk.DiskThinProvisioned,
			ControllerIndex:     disk.DiskControllerIndex,
		})
	}

	vm, err := template.Clone(ctx, &driver.CloneConfig{
		Name:            s.Location.VMName,
		Folder:          s.Location.Folder,
		Cluster:         s.Location.Cluster,
		Host:            s.Location.Host,
		ResourcePool:    s.Location.ResourcePool,
		Datastore:       s.Location.Datastore,
		LinkedClone:     s.Config.LinkedClone,
		Network:         s.Config.Network,
		MacAddress:      strings.ToLower(s.Config.MacAddress),
		Annotation:      s.Config.Notes,
		VAppProperties:  s.Config.VAppConfig.Properties,
		PrimaryDiskSize: s.Config.DiskSize,
		StorageConfig: driver.StorageConfig{
			DiskControllerType: s.Config.StorageConfig.DiskControllerType,
			Storage:            disks,
		},
	})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if vm == nil {
		return multistep.ActionHalt
	}
	if s.Config.Destroy {
		state.Put("destroy_vm", s.Config.Destroy)
	}
	state.Put("vm", vm)
	return multistep.ActionContinue
}

func (s *StepCloneVM) Cleanup(state multistep.StateBag) {
	common.CleanupVM(state)
}
