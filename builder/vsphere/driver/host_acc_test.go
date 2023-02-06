// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"testing"
)

func TestHostAcc(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	d := newTestDriver(t)
	host, err := d.FindHost(TestHostName)
	if err != nil {
		t.Fatalf("Cannot find the default host '%v': %v", "datastore1", err)
	}

	info, err := host.Info("name")
	if err != nil {
		t.Fatalf("Cannot read host properties: %v", err)
	}
	if info.Name != TestHostName {
		t.Errorf("Wrong host name: expected '%v', got: '%v'", TestHostName, info.Name)
	}
}
