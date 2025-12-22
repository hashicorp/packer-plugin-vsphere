// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"testing"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common/utils"
)

func TestHostAcc(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	d := newTestDriver(t)
	host, err := d.FindHost(utils.DefaultVsphereHost)
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}

	info, err := host.Info("name")
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if info.Name != utils.DefaultVsphereHost {
		t.Errorf("unexpected result: expected '%s', but returned '%s'", utils.DefaultVsphereHost, info.Name)
	}
}
