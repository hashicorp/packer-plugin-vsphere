// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vmware/govmomi/vim25/types"
)

// VSphereVersion represents a version number.
type VSphereVersion struct {
	Product string
	Major   int
	Minor   int
	Patch   int
	Build   int
}

// parseVersion creates a new VSphereVersion from a parsed version string and build number.
func parseVersion(name, version, build string) (VSphereVersion, error) {
	v := VSphereVersion{
		Product: name,
	}
	s := strings.Split(version, ".")
	if len(s) < 3 {
		return v, fmt.Errorf("version string %q has less than 3 components", version)
	}
	var err error
	v.Major, err = strconv.Atoi(s[0])
	if err != nil {
		return v, fmt.Errorf("could not parse major version %q from version string %q", s[0], version)
	}
	v.Minor, err = strconv.Atoi(s[1])
	if err != nil {
		return v, fmt.Errorf("could not parse minor version %q from version string %q", s[1], version)
	}
	v.Patch, err = strconv.Atoi(s[2])
	if err != nil {
		return v, fmt.Errorf("could not parse patch version %q from version string %q", s[2], version)
	}
	v.Build, err = strconv.Atoi(build)
	if err != nil {
		return v, fmt.Errorf("could not parse build version string %q", build)
	}

	return v, nil
}

// parseVersionFromAboutInfo returns a populated VSphereVersion from an AboutInfo data object.
func parseVersionFromAboutInfo(info types.AboutInfo) VSphereVersion {
	v, err := parseVersion(info.Name, info.Version, info.Build)
	if err != nil {
		// Return a default version if parsing fails.
		return VSphereVersion{
			Product: info.Name,
			Major:   7,
			Minor:   0,
			Patch:   0,
			Build:   0, // Build number is not important.
		}
	}
	return v
}

// String implements stringer for VSphereVersion.
func (v VSphereVersion) String() string {
	return fmt.Sprintf("%s %d.%d.%d build-%d", v.Product, v.Major, v.Minor, v.Patch, v.Build)
}

// ProductEqual returns true if this version's product name is the same as the supplied version's name.
func (v VSphereVersion) ProductEqual(other VSphereVersion) bool {
	return v.Product == other.Product
}

// AtLeast returns true if this version's product is equal or greater than the one required
func (v VSphereVersion) AtLeast(other VSphereVersion) bool {
	if !v.ProductEqual(other) {
		return false
	}

	vc := v.Major<<16 + v.Minor<<8 + v.Patch
	vo := other.Major<<16 + other.Minor<<8 + other.Patch
	return vc >= vo
}
