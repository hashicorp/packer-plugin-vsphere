// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"testing"

	"github.com/vmware/govmomi/simulator"
)

func TestVCenterDriver_FindResourcePool(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	defer sim.Close()

	res, err := sim.driver.FindResourcePool("", "DC0_H0", "")
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if res == nil {
		t.Fatalf("unexpected result: expected '%v', but returned 'nil'", res)
	}
	expectedResourcePool := "Resources"
	if res.pool.Name() != expectedResourcePool {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", expectedResourcePool, res.pool.Name())
	}
}

func TestVCenterDriver_FindResourcePoolStandaloneESX(t *testing.T) {
	// Standalone ESX host without a vCenter instance
	model := simulator.ESX()
	defer model.Remove()

	opts := simulator.VPX()
	model.Datastore = opts.Datastore
	model.Machine = opts.Machine
	model.Autostart = opts.Autostart
	model.DelayConfig.Delay = opts.DelayConfig.Delay
	model.DelayConfig.MethodDelay = opts.DelayConfig.MethodDelay
	model.DelayConfig.DelayJitter = opts.DelayConfig.DelayJitter

	sim, err := NewCustomVCenterSimulator(model)
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	defer sim.Close()

	res, err := sim.driver.FindResourcePool("", "localhost.localdomain", "")
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if res == nil {
		t.Fatalf("unexpected result: expected '%v', but returned 'nil'", res)
	}
	expectedResourcePool := "Resources"
	if res.pool.Name() != expectedResourcePool {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", expectedResourcePool, res.pool.Name())
	}

	// Invalid resource name should look for default resource pool
	res, err = sim.driver.FindResourcePool("", "localhost.localdomain", "invalid")
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if res == nil {
		t.Fatalf("unexpected result: expected '%v', but returned 'nil'", res)
	}
	if res.pool.Name() != expectedResourcePool {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", expectedResourcePool, res.pool.Name())
	}
}
