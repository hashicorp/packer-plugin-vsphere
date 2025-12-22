// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package version

import "github.com/hashicorp/packer-plugin-sdk/version"

var (
	Version           = "2.1.0"
	VersionPrerelease = "dev"
	VersionMetadata   = ""
	PluginVersion     = version.NewPluginVersion(Version, VersionPrerelease, VersionMetadata)
)
