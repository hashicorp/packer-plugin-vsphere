// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere_template

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

type stepChooseDatacenter struct {
	Datacenter string
}

func (s *stepChooseDatacenter) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	cli := state.Get("client").(*govmomi.Client)
	finder := find.NewFinder(cli.Client, false)

	ui.Message("Choosing datacenter...")

	// Find the datacenter or use the default one if not specified.
	dc, err := finder.DatacenterOrDefault(context.Background(), s.Datacenter)
	if err != nil {
		err = fmt.Errorf("error finding datacenter %s: %s", s.Datacenter, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("dcPath", dc.InventoryPath)

	return multistep.ActionContinue
}

func (s *stepChooseDatacenter) Cleanup(multistep.StateBag) {}
