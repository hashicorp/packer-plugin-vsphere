// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"testing"

	"github.com/vmware/govmomi/vim25/types"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		product string
		version string
		build   string
		expect  VSphereVersion
		isError bool
	}{
		{
			name:    "6.5",
			product: "VMware vCenter Server",
			version: "6.5.0",
			build:   "12345678",
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   6,
				Minor:   5,
				Patch:   0,
				Build:   12345678,
			},
			isError: false,
		},
		{
			name:    "6.7",
			product: "VMware vCenter Server",
			version: "6.7.0",
			build:   "12345678",
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   6,
				Minor:   7,
				Patch:   0,
				Build:   12345678,
			},
			isError: false,
		},
		{
			name:    "7.0",
			product: "VMware vCenter Server",
			version: "7.0.0",
			build:   "12345678",
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   0,
				Build:   12345678,
			},
			isError: false,
		},
		{
			name:    "7.0 Update 1",
			product: "VMware vCenter Server",
			version: "7.0.1",
			build:   "12345678",
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			isError: false,
		},
		{
			name:    "7.0 Update 2",
			product: "VMware vCenter Server",
			version: "7.0.2",
			build:   "12345678",
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   2,
				Build:   12345678,
			},
			isError: false,
		},
		{
			name:    "7.0 Update 3",
			product: "VMware vCenter Server",
			version: "7.0.3",
			build:   "12345678",
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   3,
				Build:   12345678,
			},
			isError: false,
		},
		{
			name:    "8.0",
			product: "VMware vCenter Server",
			version: "8.0.0",
			build:   "12345678",
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   8,
				Minor:   0,
				Patch:   0,
				Build:   12345678,
			},
			isError: false,
		},
		{
			name:    "8.0 Update 1",
			product: "VMware vCenter Server",
			version: "8.0.1",
			build:   "12345678",
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   8,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			isError: false,
		},
		{
			name:    "8.0 Update 2",
			product: "VMware vCenter Server",
			version: "8.0.2",
			build:   "12345678",
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   8,
				Minor:   0,
				Patch:   2,
				Build:   12345678,
			},
			isError: false,
		},
		{
			name:    "Invalid version string",
			product: "VMware vCenter Server",
			version: "7.0",
			build:   "12345678",
			expect:  VSphereVersion{},
			isError: true,
		},
		{
			name:    "Invalid major version",
			product: "VMware vCenter Server",
			version: "a.0.0",
			build:   "12345678",
			expect:  VSphereVersion{},
			isError: true,
		},
		{
			name:    "Invalid minor version",
			product: "VMware vCenter Server",
			version: "7.b.0",
			build:   "12345678",
			expect:  VSphereVersion{},
			isError: true,
		},
		{
			name:    "Invalid patch version",
			product: "VMware vCenter Server",
			version: "7.0.c",
			build:   "12345678",
			expect:  VSphereVersion{},
			isError: true,
		},
		{
			name:    "Invalid build number",
			product: "VMware vCenter Server",
			version: "7.0.0",
			build:   "abcdefgh",
			expect:  VSphereVersion{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVersion(tt.product, tt.version, tt.build)
			if (err != nil) != tt.isError {
				t.Errorf("parseVersion() error = %v, isError = %v", err, tt.isError)
				return
			}
			if !tt.isError {
				if got.Product != tt.expect.Product || got.Major != tt.expect.Major ||
					got.Minor != tt.expect.Minor || got.Patch != tt.expect.Patch ||
					got.Build != tt.expect.Build {
					t.Errorf("parseVersion() = %v, expect %v", got, tt.expect)
				}
			}
		})
	}
}

func TestParseVersionFromAboutInfo(t *testing.T) {
	tests := []struct {
		name   string
		about  types.AboutInfo
		expect VSphereVersion
	}{
		{
			name: "Valid ",
			about: types.AboutInfo{
				Name:    "VMware vCenter Server",
				Version: "7.0.1",
				Build:   "12345678",
			},
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
		},
		{
			name: "Invalid About Info",
			about: types.AboutInfo{
				Name:    "VMware vCenter Server",
				Version: "invalid",
				Build:   "12345678",
			},
			expect: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   0,
				Build:   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVersionFromAboutInfo(tt.about)
			if got.Product != tt.expect.Product || got.Major != tt.expect.Major ||
				got.Minor != tt.expect.Minor || got.Patch != tt.expect.Patch ||
				got.Build != tt.expect.Build {
				t.Errorf("parseVersionFromAboutInfo() = %v, expect %v", got, tt.expect)
			}
		})
	}
}

func TestVSphereVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version VSphereVersion
		expect  string
	}{
		{
			name: "Basic Version String",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expect: "VMware vCenter Server 7.0.1 build-12345678",
		},
		{
			name: "vCenter Server Version String",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expect: "VMware vCenter Server 7.0.1 build-12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()
			if got != tt.expect {
				t.Errorf("VSphereVersion.String() = %v, expect %v", got, tt.expect)
			}
		})
	}
}

func TestVSphereVersionProductEqual(t *testing.T) {
	tests := []struct {
		name     string
		version  VSphereVersion
		compare  VSphereVersion
		expected bool
	}{
		{
			name: "Same Product",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			compare: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   8,
				Minor:   0,
				Patch:   0,
				Build:   12345678,
			},
			expected: true,
		},
		{
			name: "Different Product",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			compare: VSphereVersion{
				Product: "VMware ESXi",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.ProductEqual(tt.compare)
			if got != tt.expected {
				t.Errorf("VSphereVersion.ProductEqual() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestVSphereVersionAtLeast(t *testing.T) {
	tests := []struct {
		name            string
		version         VSphereVersion
		requiredVersion VSphereVersion
		expected        bool
	}{
		{
			name: "Same Version",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			requiredVersion: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expected: true,
		},
		{
			name: "Greater Than Required (Major)",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   8,
				Minor:   0,
				Patch:   0,
				Build:   12345678,
			},
			requiredVersion: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expected: true,
		},
		{
			name: "Greater Than Required (Minor)",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   1,
				Patch:   0,
				Build:   12345678,
			},
			requiredVersion: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expected: true,
		},
		{
			name: "Greater Than required (Patch)",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   2,
				Build:   12345678,
			},
			requiredVersion: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expected: true,
		},
		{
			name: "Less Than Required (Major)",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   6,
				Minor:   5,
				Patch:   0,
				Build:   12345678,
			},
			requiredVersion: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expected: false,
		},
		{
			name: "Less than Required (Minor)",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   0,
				Build:   12345678,
			},
			requiredVersion: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expected: false,
		},
		{
			name: "Different Product",
			version: VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			requiredVersion: VSphereVersion{
				Product: "VMware ESXi",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   12345678,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.AtLeast(tt.requiredVersion)
			if got != tt.expected {
				t.Errorf("VSphereVersion.AtLeast() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestVSphereVersionBuildComparison(t *testing.T) {

	version1 := VSphereVersion{
		Product: "VMware vCenter Server",
		Major:   7,
		Minor:   0,
		Patch:   1,
		Build:   12345678,
	}

	version2 := VSphereVersion{
		Product: "VMware vCenter Server",
		Major:   7,
		Minor:   0,
		Patch:   1,
		Build:   87654321,
	}

	// Versions should be equal regardless of build number
	if !version1.AtLeast(version2) {
		t.Errorf("AtLeast should return true for same versions with different builds: %s vs %s",
			version1.String(), version2.String())
	}

	if !version2.AtLeast(version1) {
		t.Errorf("AtLeast should return true for same versions with different builds: %s vs %s",
			version2.String(), version1.String())
	}

	// Higher version with lower build number should still be greater
	higherVersion := VSphereVersion{
		Product: "VMware vCenter Server",
		Major:   7,
		Minor:   0,
		Patch:   2,
		Build:   12345678, // Lower build number
	}

	lowerVersion := VSphereVersion{
		Product: "VMware vCenter Server",
		Major:   7,
		Minor:   0,
		Patch:   1,
		Build:   87654321, // Higher build number
	}

	if !higherVersion.AtLeast(lowerVersion) {
		t.Errorf("higher version with lower build should be greater: %s should be >= %s",
			higherVersion.String(), lowerVersion.String())
	}

	if lowerVersion.AtLeast(higherVersion) {
		t.Errorf("lower version with higher build should not be greater: %s should not be >= %s",
			lowerVersion.String(), higherVersion.String())
	}
}

func TestVSphereVersionCheck701(t *testing.T) {
	// Test version comparison for vSphere 7.0.1 or later
	tests := []struct {
		name     string
		version  string
		build    string
		expected bool
	}{
		{"vCenter 6.5", "6.5.0", "12345678", false},
		{"vCenter 6.7", "6.7.0", "12345678", false},
		{"vCenter 7.0.0", "7.0.0", "12345678", false},
		{"vCenter 7.0.1", "7.0.1", "12345678", true},
		{"vCenter 7.0.2", "7.0.2", "12345678", true},
		{"vCenter 7.0.3", "7.0.3", "12345678", true},
		{"vCenter 8.0.0", "8.0.0", "12345678", true},
		{"vCenter 8.0.1", "8.0.1", "12345678", true},
		{"vCenter 8.0.2", "8.0.2", "12345678", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the test version
			version, err := parseVersion("VMware vCenter Server", tt.version, tt.build)
			if err != nil {
				t.Fatalf("error parsing version: %v", err)
			}

			// Create the required version (7.0.1)
			requiredVersion := VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
				Build:   0, // Build number doesn't matter for AtLeast comparison
			}

			// Check if the version is at least 7.0.1
			result := version.AtLeast(requiredVersion)

			if result != tt.expected {
				t.Errorf("Version %s >= 7.0.1: expected %v, got %v", tt.version, tt.expected, result)
			}
			// Check the version check against step_config_params.go example
			result = version.AtLeast(VSphereVersion{Product: version.Product, Major: 7, Minor: 0, Patch: 1})
			if result != tt.expected {
				t.Errorf("error during version check %s: expected %v, got %v", tt.version, tt.expected, result)
			}
		})
	}
}

func TestVCenterServerVersion701OrLater(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		build    string
		expected bool
	}{
		{"vCenter 6.5.0", "6.5.0", "12345678", false},
		{"vCenter 6.7.0", "6.7.0", "12345678", false},
		{"vCenter 7.0.0", "7.0.0", "12345678", false},
		{"vCenter 7.0.1", "7.0.1", "12345678", true},
		{"vCenter 7.0.2", "7.0.2", "12345678", true},
		{"vCenter 7.0.3", "7.0.3", "12345678", true},
		{"vCenter 8.0.0", "8.0.0", "12345678", true},
		{"vCenter 8.0.1", "8.0.1", "12345678", true},
		{"vCenter 8.0.2", "8.0.2", "12345678", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseVersion("VMware vCenter Server", tt.version, tt.build)
			if err != nil {
				t.Fatalf("error parsing version: %v", err)
			}

			isAtLeast701 := version.AtLeast(VSphereVersion{
				Product: "VMware vCenter Server",
				Major:   7,
				Minor:   0,
				Patch:   1,
			})

			if isAtLeast701 != tt.expected {
				t.Errorf("error during version check %s: expected %v, got %v", tt.version, tt.expected, isAtLeast701)
			}

			result := version.AtLeast(VSphereVersion{Product: version.Product, Major: 7, Minor: 0, Patch: 1})
			if result != tt.expected {
				t.Errorf("error during version check %s: expected %v, got %v", tt.version, tt.expected, result)
			}
		})
	}
}
