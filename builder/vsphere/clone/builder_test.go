// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package clone

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestCloneBuilder_ImplementsBuilder(t *testing.T) {
	var _ packersdk.Builder = &Builder{}
}
