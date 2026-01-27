// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
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
	DatastoreRefs      []*types.ManagedObjectReference
}

// AddStorageDevices adds virtual storage devices to an existing device list.
// It creates a controller for each type specified in DiskControllerType and attaches
// virtual disks to the controllers. If DatastoreRefs is provided, each disk is placed
// on the corresponding datastore; otherwise, disks inherit the VM's datastore.
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

	for i, dc := range c.Storage {
		backing := &types.VirtualDiskFlatVer2BackingInfo{
			DiskMode:        string(types.VirtualDiskModePersistent),
			ThinProvisioned: types.NewBool(dc.DiskThinProvisioned),
			EagerlyScrub:    types.NewBool(dc.DiskEagerlyScrub),
		}

		if i < len(c.DatastoreRefs) && c.DatastoreRefs[i] != nil {
			backing.Datastore = c.DatastoreRefs[i]
		}

		disk := &types.VirtualDisk{
			VirtualDevice: types.VirtualDevice{
				Key:     existingDevices.NewKey(),
				Backing: backing,
			},
			CapacityInKB: dc.DiskSize * 1024,
		}

		existingDevices.AssignController(disk, controllers[dc.ControllerIndex])
		existingDevices = append(existingDevices, disk)
		newDevices = append(newDevices, disk)
	}

	return newDevices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
}

// findDisk scans a list of virtual devices and retrieves a single virtual disk.
// Returns an error if no disk or multiple disks are found.
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
		return nil, errors.New("error finding virtual disk")
	case 1:
		return disks[0], nil
	}
	return nil, errors.New("more than one virtual disk found, only a single disk is allowed")
}
