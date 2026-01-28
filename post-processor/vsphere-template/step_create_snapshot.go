// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere_template

import (
	"context"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/vmware/govmomi"
	"github.com/vmware/packer-plugin-vsphere/post-processor/vsphere"
)

type StepCreateSnapshot struct {
	VMName              string
	RemoteFolder        string
	SnapshotName        string
	SnapshotDescription string
	SnapshotEnable      bool
}

func NewStepCreateSnapshot(artifact packersdk.Artifact, p *PostProcessor) *StepCreateSnapshot {
	// Set the default folder.
	remoteFolder := "Discovered virtual machine"
	vmname := artifact.Id()

	if artifact.BuilderId() == vsphere.BuilderId {
		id := strings.Split(artifact.Id(), "::")
		remoteFolder = id[1]
		vmname = id[2]
	}

	return &StepCreateSnapshot{
		VMName:              vmname,
		RemoteFolder:        remoteFolder,
		SnapshotEnable:      p.config.SnapshotEnable,
		SnapshotName:        p.config.SnapshotName,
		SnapshotDescription: p.config.SnapshotDescription,
	}
}

func (s *StepCreateSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	cli := state.Get("client").(*govmomi.Client)
	dcPath := state.Get("dcPath").(string)

	if !s.SnapshotEnable {
		return multistep.ActionContinue
	}

	ui.Say("Creating virtual machine snapshot...")

	vm, err := findVirtualMachine(cli, dcPath, s.VMName, s.RemoteFolder)
	if err != nil {
		state.Put("error", err)
		ui.Errorf("%s", err)
		return multistep.ActionHalt
	}

	task, err := vm.CreateSnapshot(context.Background(), s.SnapshotName, s.SnapshotDescription, false, false)
	if err != nil {
		state.Put("error", err)
		ui.Errorf("%s", err)
		return multistep.ActionHalt
	}

	if err = task.Wait(context.Background()); err != nil {
		state.Put("error", err)
		ui.Errorf("%s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCreateSnapshot) Cleanup(multistep.StateBag) {}
