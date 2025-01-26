// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"errors"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type Disk struct {
	DiskSize            int64
	DiskEagerlyScrub    bool
	DiskThinProvisioned bool
	ControllerIndex     int
}

type StorageConfig struct {
	DiskControllerType []string
	Storage            []Disk
}

// AddStorageDevices adds virtual storage devices to an existing device list
// based on the configuration. Adds a new controller for each controller type
// specified in the configuration and adds virtual disks to the controller.
func (c *StorageConfig) AddStorageDevices(existingDevices object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error) {
	newDevices := object.VirtualDeviceList{}

	var controllers []types.BaseVirtualController
	for _, controllerType := range c.DiskControllerType {
		var device types.BaseVirtualDevice
		var err error
		switch controllerType {
		case "nvme":
			device, err = existingDevices.CreateNVMEController()
		case "sata":
			device, err = existingDevices.CreateSATAController()
		default:
			device, err = existingDevices.CreateSCSIController(controllerType)
		}
		if err != nil {
			return nil, err
		}
		existingDevices = append(existingDevices, device)
		newDevices = append(newDevices, device)
		controller, err := existingDevices.FindDiskController(existingDevices.Name(device))
		if err != nil {
			return nil, err
		}
		controllers = append(controllers, controller)
	}

	for _, dc := range c.Storage {
		disk := &types.VirtualDisk{
			VirtualDevice: types.VirtualDevice{
				Key: existingDevices.NewKey(),
				Backing: &types.VirtualDiskFlatVer2BackingInfo{
					DiskMode:        string(types.VirtualDiskModePersistent),
					ThinProvisioned: types.NewBool(dc.DiskThinProvisioned),
					EagerlyScrub:    types.NewBool(dc.DiskEagerlyScrub),
				},
			},
			CapacityInKB: dc.DiskSize * 1024,
		}

		existingDevices.AssignController(disk, controllers[dc.ControllerIndex])
		existingDevices = append(existingDevices, disk)
		newDevices = append(newDevices, disk)
	}

	return newDevices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
}

// findDisk scans a list of virtual devices and retrieves a single virtual disk
// if exactly one is found.  Returns an error if no disk or multiple disks are found.
// TODO: Add support for multiple disks.
func findDisk(devices object.VirtualDeviceList) (*types.VirtualDisk, error) {
	var disks []*types.VirtualDisk
	for _, device := range devices {
		switch d := device.(type) {
		case *types.VirtualDisk:
			disks = append(disks, d)
		}
	}

	switch len(disks) {
	case 0:
		// No disks found.
		return nil, errors.New("error finding virtual disk")
	case 1:
		// Single disk found.
		return disks[0], nil
	}
	// Multiple disks found.
	return nil, errors.New("more than one virtual disk found, only a single disk is allowed")
}
