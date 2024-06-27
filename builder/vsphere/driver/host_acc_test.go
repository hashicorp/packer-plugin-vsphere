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
		t.Fatalf("unexpected error: '%s'", err)
	}

	info, err := host.Info("name")
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if info.Name != TestHostName {
		t.Errorf("unexpected result: expected '%s', but returned '%s'", TestHostName, info.Name)
	}
}
