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

func (vm *VirtualMachineDriver) AddSATAController() error {
	sata := &types.VirtualAHCIController{}
	return vm.addDevice(sata)
}

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
// An error will occur If n is larger then the attached CD-ROM count.
// n == 0 results in no CD-ROMs being removed.
func (vm *VirtualMachineDriver) RemoveNCdroms(n int) error {
	if n == 0 {
		return nil
	}
	devices, err := vm.Devices()
	if err != nil {
		return err
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	if (n < 0) || (n > len(cdroms)) {
		return fmt.Errorf("invalid number: n must be between 0 and %d", len(cdroms))
	}

	cdroms = cdroms[:n]
	if err = vm.RemoveDevice(true, cdroms...); err != nil {
		return err
	}

	return nil
}

func (vm *VirtualMachineDriver) EjectCdroms() error {
	devices, err := vm.Devices()
	if err != nil {
		return err
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	for _, cd := range cdroms {
		c := cd.(*types.VirtualCdrom)
		c.Backing = &types.VirtualCdromRemotePassthroughBackingInfo{}
		c.Connectable = &types.VirtualDeviceConnectInfo{}
		err := vm.vm.EditDevice(vm.driver.ctx, c)
		if err != nil {
			return err
		}
	}

	return nil
}
