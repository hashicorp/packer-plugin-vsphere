// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vmware/govmomi/vim25/types"
)

func TestVirtualMachineDriver_FindAndAddSATAController(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	defer sim.Close()

	vm, _ := sim.ChooseSimulatorPreCreatedVM()

	_, err = vm.FindSATAController()
	if err != nil && !strings.Contains(err.Error(), "no available SATA controller") {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if err == nil {
		t.Fatalf("unexpected error: '%s'", err)
	}

	if err := vm.AddSATAController(); err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}

	sc, err := vm.FindSATAController()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if sc == nil {
		t.Fatalf("unexpected result: expected '%s', but returned '%v'", "sata controller", sc)
	}
}

func TestVirtualMachineDriver_CreateAndRemoveCdrom(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	defer sim.Close()

	vm, _ := sim.ChooseSimulatorPreCreatedVM()

	// Add SATA Controller
	if err := vm.AddSATAController(); err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}

	// Verify if controller was created
	sc, err := vm.FindSATAController()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if sc == nil {
		t.Fatalf("unexpected result: expected '%s', but returned '%v'", "sata controller", sc)
	}

	// Create CDROM
	controller := sc.GetVirtualController()
	cdrom, err := vm.CreateCdrom(controller)
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if cdrom == nil {
		t.Fatalf("unexpected result: expected '%s', but returned '%v'", "cd-rom", cdrom)
	}

	// Verify if CDROM was created
	cdroms, err := vm.CdromDevices()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if len(cdroms) != 1 {
		t.Fatalf("unexpected result: expected '1', but returned '%d'", len(cdroms))
	}

	// Remove CDROM
	err = vm.RemoveCdroms()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	// Verify if CDROM was removed
	cdroms, err = vm.CdromDevices()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if len(cdroms) != 0 {
		t.Fatalf("unexpected result: expected '0', but returned '%d'", len(cdroms))
	}
}

func TestVirtualMachineDriver_EjectCdrom(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	defer sim.Close()

	vm, _ := sim.ChooseSimulatorPreCreatedVM()

	// Add SATA Controller
	if err := vm.AddSATAController(); err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}

	// Verify if controller was created
	sc, err := vm.FindSATAController()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if sc == nil {
		t.Fatalf("unexpected result: expected '%s', but returned '%v'", "sata controller", sc)
	}

	// Create CDROM
	controller := sc.GetVirtualController()
	cdrom, err := vm.CreateCdrom(controller)
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if cdrom == nil {
		t.Fatalf("unexpected result: expected '%s', but returned '%v'", "cd-rom", cdrom)
	}

	// Verify if CDROM was created
	cdroms, err := vm.CdromDevices()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if len(cdroms) != 1 {
		t.Fatalf("unexpected result: expected '1', but returned '%d'", len(cdroms))
	}

	// Remove CDROM
	err = vm.EjectCdroms()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	// Verify if CDROM was removed
	cdroms, err = vm.CdromDevices()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if len(cdroms) != 1 {
		t.Fatalf("unexpected result: expected '1', but returned '%d'", len(cdroms))
	}
	cd, ok := cdroms[0].(*types.VirtualCdrom)
	if !ok {
		t.Fatalf("unexpected result: expected '%s', but returned '%v'", "cdrom", cd)
	}
	if diff := cmp.Diff(cd.Backing, &types.VirtualCdromRemotePassthroughBackingInfo{}); diff != "" {
		t.Fatalf("unexpected result: '%s'", diff)
	}
	if diff := cmp.Diff(cd.Connectable, &types.VirtualDeviceConnectInfo{}); diff != "" {
		t.Fatalf("unexpected result: '%s'", diff)
	}
}
