// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clone

import (
	"testing"

	"github.com/vmware/govmomi/vim25/types"
)

// TestSysprepFieldsMutuallyExclusive validates an error message is thrown
// when both windows_sysprep_text and windows_sysprep_file are included in the config
func TestSysprepFieldsMutuallyExclusive(t *testing.T) {
	// Customize config with windows sysprep file and windows sysprep text set.
	config := &CustomizeConfig{
		WindowsSysPrepFile: "path-to-file",
		WindowsSysPrepText: "text",
		NetworkInterfaces: []NetworkInterface{
			NetworkInterface{},
		},
	}

	// Expected error message
	expectedError := errCustomizeOptionMutalExclusive
	_, errors := config.Prepare()

	// Make sure we only received on error
	expectedErrorLength := 1
	if len(errors) != expectedErrorLength {
		t.Fatalf("expected errors of length %d but got %d", expectedErrorLength, len(errors))
	}

	// Validate the error messages are what we expect.
	if errors[0].Error() != expectedError.Error() {
		t.Fatalf("expected error message %s, but got error message %s", expectedError, errors[0].Error())
	}
}

// TestWindowsSysprepFilePrintsWarning tests a warning about
// the field being deprecated is returned when windows_sysprep_file field is set.
func TestWindowsSysprepFilePrintsWarning(t *testing.T) {
	// Customize config with windows sysprep file and windows sysprep text set.
	config := &CustomizeConfig{
		WindowsSysPrepFile: "path-to-file",
		NetworkInterfaces: []NetworkInterface{
			NetworkInterface{},
		},
	}

	// Expected warning message
	expectedWarning := windowsSysprepFileDeprecatedMessage
	warnings, errors := config.Prepare()

	// Fail if there were errors
	if len(errors) > 0 {
		t.Fatalf("there were errors when running prepare")
	}

	// Search warnings array for the warning message
	found := false
	for _, warning := range warnings {
		if warning == expectedWarning {
			found = true
			break
		}
	}

	// If we didn't find the expect warning message fail.
	if found == false {
		t.Fatalf("didn't find %s in warnings array", expectedWarning)
	}

}

// TestWindowsSysprepTextSetsContent validates that when WindowSyrepText is set
// that it sets the value for the vSphere customization spec.
func TestWindowsSysprepTextSetsContent(t *testing.T) {
	// Expected text
	text := "xml customization spec"
	config := &CustomizeConfig{
		WindowsSysPrepText: text,
		NetworkInterfaces: []NetworkInterface{
			NetworkInterface{},
		},
	}

	// Create a step customize object with the given config
	stepCustomize := &StepCustomize{Config: config}

	// Get the identity settings for the customize step
	baseCustomizationSettings, err := stepCustomize.identitySettings()
	if err != nil {
		t.Fatalf("identity settings had unexpected errors. %v", err)
	}

	// Cast the result to the vsphere govmomi type
	sysprepText, ok := baseCustomizationSettings.(*types.CustomizationSysprepText)
	if !ok {
		t.Fatalf("identity settings did not return CustomizationSysprepText type")
	}

	// Make sure the sysprep text value is equal to what was passed in via the config.
	if sysprepText.Value != text {
		t.Fatalf("expected the customization spec to contain %s but was %s", text, sysprepText.Value)
	}
}
