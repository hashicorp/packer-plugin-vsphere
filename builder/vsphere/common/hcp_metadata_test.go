// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/types"
)

func TestGetVMMetadata(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	defer sim.Close()

	state := new(multistep.BasicStateBag)
	state.Put("content_library_datastore", []string{"tmpl-datastore-mock"})

	vm, vmSim := sim.ChooseSimulatorPreCreatedVM()
	confSpec := types.VirtualMachineConfigSpec{Annotation: "simple vm description"}
	if err := vm.Reconfigure(confSpec); err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	datastore := sim.model.Service.Context.Map.Any("Datastore").(*simulator.Datastore)

	metadata := GetVMMetadata(vm.(*driver.VirtualMachineDriver), state)
	// Validate Labels
	expectedLabels := map[string]string{
		"annotation":         vmSim.Config.Annotation,
		"num_cpu":            fmt.Sprintf("%d", vmSim.Config.Hardware.NumCPU),
		"memory_mb":          fmt.Sprintf("%d", vmSim.Config.Hardware.MemoryMB),
		"datastore":          datastore.Name,
		"network":            "DC0_DVPG0",
		"vsphere_uuid":       vmSim.Config.Uuid,
		"template_datastore": "tmpl-datastore-mock",
	}

	if diff := cmp.Diff(expectedLabels, metadata); diff != "" {
		t.Fatalf("unexpected result: '%s'", diff)
	}
}
