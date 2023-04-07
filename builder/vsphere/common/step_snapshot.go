// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type StepCreateSnapshot struct {
	CreateSnapshot bool
	SnapshotName   string
}

func (s *StepCreateSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)

	if s.CreateSnapshot {
		ui.Say("Creating snapshot...")

		snapshotName := "Created By Packer"

		if s.SnapshotName != "" {
			snapshotName = s.SnapshotName
		}

		err := vm.CreateSnapshot(snapshotName)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateSnapshot) Cleanup(state multistep.StateBag) {}
