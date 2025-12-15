// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"errors"
	"fmt"

	"github.com/vmware/govmomi/vim25/types"
)

var (
	ErrNoSataController = errors.New("no available SATA controller")
)

// AddSATAController adds a new SATA controller to the virtual machine configuration.
// Returns an error if the operation fails.
func (vm *VirtualMachineDriver) AddSATAController() error {
	sata := &types.VirtualAHCIController{}
	return vm.addDevice(sata)
}

// FindSATAController searches and returns the first available SATA controller
// for the virtual machine. Returns an error if no SATA controller is found or
// if there is an issue obtaining the devices.
func (vm *VirtualMachineDriver) FindSATAController() (*types.VirtualAHCIController, error) {
	l, err := vm.Devices()
	if err != nil {
		return nil, err
	}

	c := l.PickController((*types.VirtualAHCIController)(nil))
	if c == nil {
		return nil, ErrNoSataController
	}

	return c.(*types.VirtualAHCIController), nil
}

// CreateCdrom creates a new virtual CD-ROM device and attaches it to the
// specified virtual controller. It initializes the CD-ROM with default
// connectable settings, allowing guest control and automatic connection.
// Returns the created VirtualCdrom object or an error if the devices cannot
// be retrieved or assigned.
func (vm *VirtualMachineDriver) CreateCdrom(c *types.VirtualController) (*types.VirtualCdrom, error) {
	l, err := vm.Devices()
	if err != nil {
		return nil, err
	}

	device := &types.VirtualCdrom{}

	l.AssignController(device, c)

	device.Backing = &types.VirtualCdromAtapiBackingInfo{
		VirtualDeviceDeviceBackingInfo: types.VirtualDeviceDeviceBackingInfo{},
	}

	device.Connectable = &types.VirtualDeviceConnectInfo{
		AllowGuestControl: true,
		Connected:         true,
		StartConnected:    true,
	}

	return device, nil
}

// RemoveCdroms removes all virtual CD-ROM drives and associated SATA
// controllers from the virtual machine configuration.
func (vm *VirtualMachineDriver) RemoveCdroms() error {
	devices, err := vm.Devices()
	if err != nil {
		return err
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	if err = vm.RemoveDevice(true, cdroms...); err != nil {
		return err
	}

	sata := devices.SelectByType((*types.VirtualAHCIController)(nil))
	if err = vm.RemoveDevice(true, sata...); err != nil {
		return err
	}
	return nil
}

// RemoveNCdroms removes up to n CD-ROMs from the image.
// An error will occur if n is greater than the attached CD-ROM count.
// n == 0 results in no CD-ROMs being removed.
func (vm *VirtualMachineDriver) RemoveNCdroms(n int) error {
	if n == 0 {
		return nil
	}
	cdroms, err := vm.CdromDevices()
	if err != nil {
		return err
	}
	if (n < 0) || (n > len(cdroms)) {
		return fmt.Errorf("invalid number: n must be between 0 and %d", len(cdroms))
	}

	// Remove up to n CD-ROMs from the end of the list.
	// For example, removing from the end preserves lower-numbered device slots, which
	// prevents IDE controller errors where a slave device exists without a primary
	// (master) device.
	cdroms = cdroms[len(cdroms)-n:]
	if err = vm.RemoveDevice(true, cdroms...); err != nil {
		return err
	}

	return nil
}

// EjectCdroms removes all attached CD-ROM devices from the virtual machine by
// resetting their backing and connection information.
func (vm *VirtualMachineDriver) EjectCdroms() error {
	cdroms, err := vm.CdromDevices()
	if err != nil {
		return err
	}
	for _, cd := range cdroms {
		c := cd.(*types.VirtualCdrom)
		c.Backing = &types.VirtualCdromRemotePassthroughBackingInfo{}
		c.Connectable = &types.VirtualDeviceConnectInfo{}
		err := vm.vm.EditDevice(vm.driver.Ctx, c)
		if err != nil {
			return err
		}
	}

	return nil
}
