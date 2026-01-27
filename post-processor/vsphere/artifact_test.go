// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestArtifact_ImplementsArtifact(t *testing.T) {
	var _ packersdk.Artifact = &Artifact{}
}

func TestArtifact_Id(t *testing.T) {
	artifact := NewArtifact("datastore", "vmfolder", "vmname", nil)
	if artifact.Id() != "datastore::vmfolder::vmname" {
		t.Fatalf("unexpected result: must return datastore, vmfolder, and vmname split by :: as id")
	}
}
