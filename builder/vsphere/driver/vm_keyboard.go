// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/mobile/event/key"
)

type KeyInput struct {
	Scancode key.Code
	Alt      bool
	Ctrl     bool
	Shift    bool
}

// TypeOnKeyboard sends a sequence of USB scan code events to simulate keyboard
// typing on a virtual machine. The input parameter specifies the USB HID
// scancode and key modifiers like Ctrl, Alt, and Shift.
func (vm *VirtualMachineDriver) TypeOnKeyboard(input KeyInput) (int32, error) {
	var spec types.UsbScanCodeSpec

	spec.KeyEvents = append(spec.KeyEvents, types.UsbScanCodeSpecKeyEvent{
		UsbHidCode: int32(input.Scancode)<<16 | 7,
		Modifiers: &types.UsbScanCodeSpecModifierType{
			LeftControl: &input.Ctrl,
			LeftAlt:     &input.Alt,
			LeftShift:   &input.Shift,
		},
	})

	req := &types.PutUsbScanCodes{
		This: vm.vm.Reference(),
		Spec: spec,
	}

	resp, err := methods.PutUsbScanCodes(vm.driver.Ctx, vm.driver.Client.RoundTripper, req)
	if err != nil {
		return 0, err
	}

	return resp.Returnval, nil
}
