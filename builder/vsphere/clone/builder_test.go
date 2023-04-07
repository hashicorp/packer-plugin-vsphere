// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clone

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestCloneBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packersdk.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}
