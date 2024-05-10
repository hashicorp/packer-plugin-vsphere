// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type NIC,CreateConfig

package iso

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

// Defines a Network Adapter
// If no adapter is defined, network tasks (communicators, most provisioners) won't work, so it's advised to define one.
//
// Example that creates two network adapters:
//
// In JSON:
// ```json
//
//	"network_adapters": [
//	  {
//	    "network": "VM Network",
//	    "network_card": "vmxnet3"
//	  },
//	  {
//	    "network": "OtherNetwork",
//	    "network_card": "vmxnet3"
//	  }
//	],
//
// ```
// In HCL2:
// ```hcl
//
//	network_adapters {
//	    network = "VM Network"
//	    network_card = "vmxnet3"
//	}
//	network_adapters {
//	    network = "OtherNetwork"
//	    network_card = "vmxnet3"
//	}
//
// ```
type NIC struct {
	// Specifies the network to which the virtual machine will connect.
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
	// ~> **Note:** If more than one network resolves to the same name, either the inventory path to
	// network or an ID must be provided.
	//
	// ~> **Note:** If no network is specified, provide `host` to allow the plugin to search for an
	// available network.
	Network string `mapstructure:"network"`
	// Specifies the virtual machine network card type. For example `vmxnet3`.
	NetworkCard string `mapstructure:"network_card" required:"true"`
	// Specifies the network card MAC address. For example `00:50:56:00:00:00`.
	MacAddress string `mapstructure:"mac_address"`
	// Specifies whether to enable DirectPath I/O passthrough for the network device.
	Passthrough *bool `mapstructure:"passthrough"`
}

type CreateConfig struct {
	// Specifies the virtual machine hardware version. Defaults to the most current virtual machine
	// hardware version supported by the ESXi host.
	// Refer to [KB 315655](https://knowledge.broadcom.com/external/article?articleNumber=315655)
	// for more information on supported virtual hardware versions.
	Version uint `mapstructure:"vm_version"`
	// Specifies the guest operating system identifier for the virtual machine.
	// If not specified, the setting defaults to `otherGuest`.
	//
	// To get a list of supported guest operating system identifiers for your ESXi host,
	// run the following PowerShell command using `VMware.PowerCLI`:
	//
	// ```powershell
	// Connect-VIServer -Server "vc.example.com" -User "administrator@vsphere" -Password "password"
	// $esxiHost = Get-VMHost -Name "esxi.example.com"
	// $environmentBrowser = Get-View -Id $esxiHost.ExtensionData.Parent.ExtensionData.ConfigManager.EnvironmentBrowser
	// $vmxVersion = ($environmentBrowser.QueryConfigOptionDescriptor() | Where-Object DefaultConfigOption).Key
	// $osDescriptor = $environmentBrowser.QueryConfigOption($vmxVersion, $null).GuestOSDescriptor
	// $osDescriptor | Select-Object Id, Fullname
	// ```
	GuestOSType   string               `mapstructure:"guest_os_type"`
	StorageConfig common.StorageConfig `mapstructure:",squash"`
	// Specifies the network adapters for the virtual machine.
	// If no network adapter is defined, all network-related operations will be skipped.
	NICs []NIC `mapstructure:"network_adapters"`
	// Specifies the USB controllers for the virtual machine. Use `usb` for a USB 2.0 controller and
	// `xhci`` for a USB 3.0 controller.
	// -> **Note:** Maximum of one controller of each type.
	USBController []string `mapstructure:"usb_controller"`
	// Specifies the annotations for the virtual machine.
	Notes string `mapstructure:"notes"`
	// Specifies whether to destroy the virtual machine after the build is complete.
	Destroy bool `mapstructure:"destroy"`
}

func (c *CreateConfig) Prepare() []error {
	var errs []error

	if len(c.StorageConfig.DiskControllerType) == 0 {
		c.StorageConfig.DiskControllerType = append(c.StorageConfig.DiskControllerType, "")
	}

	// there should be at least one
	if len(c.StorageConfig.Storage) == 0 {
		errs = append(errs, fmt.Errorf("no storage devices have been defined"))
	}
	errs = append(errs, c.StorageConfig.Prepare()...)

	if c.GuestOSType == "" {
		c.GuestOSType = "otherGuest"
	}

	usbCount := 0
	xhciCount := 0

	for i, s := range c.USBController {
		switch s {
		// 1 and true for backwards compatibility
		case "usb", "1", "true":
			usbCount++
		case "xhci":
			xhciCount++
		// 0 and false for backwards compatibility
		case "false", "0":
			continue
		default:
			errs = append(errs, fmt.Errorf("usb_controller[%d] references an unknown usb controller", i))
		}
	}
	if usbCount > 1 || xhciCount > 1 {
		errs = append(errs, fmt.Errorf("there can only be one usb controller and one xhci controller"))
	}

	return errs
}

type StepCreateVM struct {
	Config        *CreateConfig
	Location      *common.LocationConfig
	Force         bool
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCreateVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	d := state.Get("driver").(driver.Driver)
	vmPath := path.Join(s.Location.Folder, s.Location.VMName)

	err := d.PreCleanVM(ui, vmPath, s.Force, s.Location.Cluster, s.Location.Host, s.Location.ResourcePool)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Creating virtual machine...")

	// add network/network card an the first nic for backwards compatibility in the type is defined
	var networkCards []driver.NIC
	for _, nic := range s.Config.NICs {
		networkCards = append(networkCards, driver.NIC{
			Network:     nic.Network,
			NetworkCard: nic.NetworkCard,
			MacAddress:  strings.ToLower(nic.MacAddress),
			Passthrough: nic.Passthrough,
		})
	}

	// add disk as the first drive for backwards compatibility if the type is defined
	var disks []driver.Disk
	for _, disk := range s.Config.StorageConfig.Storage {
		disks = append(disks, driver.Disk{
			DiskSize:            disk.DiskSize,
			DiskEagerlyScrub:    disk.DiskEagerlyScrub,
			DiskThinProvisioned: disk.DiskThinProvisioned,
			ControllerIndex:     disk.DiskControllerIndex,
		})
	}

	vm, err := d.CreateVM(&driver.CreateConfig{
		StorageConfig: driver.StorageConfig{
			DiskControllerType: s.Config.StorageConfig.DiskControllerType,
			Storage:            disks,
		},
		Annotation:    s.Config.Notes,
		Name:          s.Location.VMName,
		Folder:        s.Location.Folder,
		Cluster:       s.Location.Cluster,
		Host:          s.Location.Host,
		ResourcePool:  s.Location.ResourcePool,
		Datastore:     s.Location.Datastore,
		GuestOS:       s.Config.GuestOSType,
		NICs:          networkCards,
		USBController: s.Config.USBController,
		Version:       s.Config.Version,
	})
	if err != nil {
		state.Put("error", fmt.Errorf("error creating virtual machine: %v", err))
		return multistep.ActionHalt
	}
	if s.Config.Destroy {
		state.Put("destroy_vm", s.Config.Destroy)
	}
	state.Put("vm", vm)

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	common.CleanupVM(state)
}
