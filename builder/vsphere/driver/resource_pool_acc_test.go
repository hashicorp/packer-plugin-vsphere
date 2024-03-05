// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import "testing"

func TestResourcePoolAcc(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	d := newTestDriver(t)
	p, err := d.FindResourcePool("", "esxi-01.example.com", "pool1/pool2")
	if err != nil {
		t.Fatalf("Cannot find the default resource pool '%v': %v", "pool1/pool2", err)
	}

	path, err := p.Path()
	if err != nil {
		t.Fatalf("Cannot read resource pool name: %v", err)
	}
	if path != "pool1/pool2" {
		t.Errorf("Wrong folder. expected: 'pool1/pool2', got: '%v'", path)
	}
}
