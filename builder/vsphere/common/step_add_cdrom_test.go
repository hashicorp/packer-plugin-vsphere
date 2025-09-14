// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/object"
)

func TestCDRomConfig_Prepare(t *testing.T) {
	tests := []struct {
		name           string
		config         *CDRomConfig
		keepConfig     *ReattachCDRomConfig
		fail           bool
		expectedErrMsg string
	}{
		{
			name:           "Empty config",
			config:         new(CDRomConfig),
			keepConfig:     new(ReattachCDRomConfig),
			fail:           false,
			expectedErrMsg: "",
		},
		{
			name:           "Valid cdrom type ide",
			config:         &CDRomConfig{CdromType: "ide"},
			keepConfig:     new(ReattachCDRomConfig),
			fail:           false,
			expectedErrMsg: "",
		},
		{
			name:           "Valid cdrom type sata",
			config:         &CDRomConfig{CdromType: "sata"},
			keepConfig:     new(ReattachCDRomConfig),
			fail:           false,
			expectedErrMsg: "",
		},
		{
			name:           "Invalid cdrom type",
			config:         &CDRomConfig{CdromType: "invalid"},
			keepConfig:     new(ReattachCDRomConfig),
			fail:           true,
			expectedErrMsg: "'cdrom_type' must be 'ide' or 'sata'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Prepare(tt.keepConfig, nil)
			if tt.fail {
				if len(errs) == 0 {
					t.Fatal("Expected error but got none")
				}
				if errs[0].Error() != tt.expectedErrMsg {
					t.Fatalf("Expected error '%s', got '%s'", tt.expectedErrMsg, errs[0])
				}
			} else {
				if len(errs) != 0 {
					t.Fatalf("Expected no error but got: %s", errs[0])
				}
			}
		})
	}
}

func TestStepAddCDRom_Run(t *testing.T) {
	tests := []struct {
		name           string
		state          *multistep.BasicStateBag
		step           *StepAddCDRom
		vmMock         *driver.VirtualMachineMock
		expectedAction multistep.StepAction
		expectedVmMock *driver.VirtualMachineMock
		fail           bool
		errMessage     string
	}{
		{
			name:  "CDRom SATA type with all iso paths set",
			state: cdAndIsoRemotePathStateBag(),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "sata",
					ISOPaths:  []string{"iso/path"},
				},
			},
			vmMock:         new(driver.VirtualMachineMock),
			expectedAction: multistep.ActionContinue,
			expectedVmMock: &driver.VirtualMachineMock{
				FindSATAControllerCalled: true,
				AddCdromCalledTimes:      3,
				AddCdromTypes:            []string{"sata", "sata", "sata"},
				AddCdromPaths:            []string{"remote/path", "iso/path", "cd/path"},
				CdromDevicesList:         object.VirtualDeviceList{nil, nil, nil},
			},
			fail:       false,
			errMessage: "",
		},
		{
			name:  "Add SATA Controller",
			state: basicStateBag(nil),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "sata",
				},
			},
			vmMock: &driver.VirtualMachineMock{
				FindSATAControllerErr: driver.ErrNoSataController,
			},
			expectedAction: multistep.ActionContinue,
			expectedVmMock: &driver.VirtualMachineMock{
				FindSATAControllerCalled: true,
				FindSATAControllerErr:    driver.ErrNoSataController,
				AddSATAControllerCalled:  true,
			},
			fail:       false,
			errMessage: "",
		},
		{
			name:  "Fail to add SATA Controller",
			state: basicStateBag(nil),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "sata",
				},
			},
			vmMock: &driver.VirtualMachineMock{
				FindSATAControllerErr: driver.ErrNoSataController,
				AddSATAControllerErr:  fmt.Errorf("AddSATAController error"),
			},
			expectedAction: multistep.ActionHalt,
			expectedVmMock: &driver.VirtualMachineMock{
				FindSATAControllerCalled: true,
				AddSATAControllerCalled:  true,
			},
			fail:       true,
			errMessage: fmt.Sprintf("error adding SATA controller: %v", fmt.Errorf("AddSATAController error")),
		},
		{
			name:  "IDE CDRom Type and Iso Path set",
			state: basicStateBag(nil),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "ide",
					ISOPaths:  []string{"iso/path"},
				},
			},
			vmMock:         new(driver.VirtualMachineMock),
			expectedAction: multistep.ActionContinue,
			expectedVmMock: &driver.VirtualMachineMock{
				AddCdromCalledTimes: 1,
				AddCdromTypes:       []string{"ide"},
				AddCdromPaths:       []string{"iso/path"},
				CdromDevicesList:    object.VirtualDeviceList{nil},
			},
			fail:       false,
			errMessage: "",
		},
		{
			name:  "Fail to add cdrom from ISOPaths",
			state: basicStateBag(nil),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					ISOPaths: []string{"iso/path"},
				},
			},
			vmMock: &driver.VirtualMachineMock{
				AddCdromErr: fmt.Errorf("AddCdrom error"),
			},
			expectedAction: multistep.ActionHalt,
			expectedVmMock: &driver.VirtualMachineMock{
				AddCdromCalledTimes: 1,
				AddCdromTypes:       []string{""},
				AddCdromPaths:       []string{"iso/path"},
				CdromDevicesList:    object.VirtualDeviceList{nil},
			},
			fail:       true,
			errMessage: fmt.Sprintf("error mounting an image 'iso/path': %v", fmt.Errorf("AddCdrom error")),
		},
		{
			name:  "Fail to add cdrom from state iso_remote_path",
			state: isoRemotePathStateBag(),
			step: &StepAddCDRom{
				Config: new(CDRomConfig),
			},
			vmMock: &driver.VirtualMachineMock{
				AddCdromErr: fmt.Errorf("AddCdrom error"),
			},
			expectedAction: multistep.ActionHalt,
			expectedVmMock: &driver.VirtualMachineMock{
				AddCdromCalledTimes: 1,
				AddCdromTypes:       []string{""},
				AddCdromPaths:       []string{"remote/path"},
				CdromDevicesList:    object.VirtualDeviceList{nil},
			},
			fail:       true,
			errMessage: fmt.Sprintf("error mounting an image 'remote/path': %v", fmt.Errorf("AddCdrom error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.Put("vm", tt.vmMock)
			if action := tt.step.Run(context.TODO(), tt.state); action != tt.expectedAction {
				t.Fatalf("Expected action '%#v', got '%#v'", tt.expectedAction, action)
			}
			err, ok := tt.state.Get("error").(error)
			if ok {
				if err.Error() != tt.errMessage {
					t.Fatalf("Expected error '%s', got '%s'", tt.errMessage, err)
				}
			} else {
				if tt.fail {
					t.Fatal("Expected error but got none")
				}
			}

			if diff := cmp.Diff(tt.vmMock, tt.expectedVmMock,
				cmpopts.IgnoreInterfaces(struct{ error }{})); diff != "" {
				t.Fatalf("Unexpected VM mock state: %s", diff)
			}
		})
	}
}

func cdAndIsoRemotePathStateBag() *multistep.BasicStateBag {
	state := basicStateBag(nil)
	state.Put("iso_remote_path", "remote/path")
	state.Put("cd_path", "cd/path")
	return state
}

func isoRemotePathStateBag() *multistep.BasicStateBag {
	state := basicStateBag(nil)
	state.Put("iso_remote_path", "remote/path")
	return state
}

func TestCDRomConfig_ValidateISOPaths(t *testing.T) {
	tests := []struct {
		name           string
		config         *CDRomConfig
		driverMock     *driver.DriverMock
		expectedErrors []string
	}{
		{
			name: "Empty ISO paths",
			config: &CDRomConfig{
				ISOPaths: []string{},
			},
			driverMock:     driver.NewDriverMock(),
			expectedErrors: []string{},
		},
		{
			name: "Valid datastore path - file exists",
			config: &CDRomConfig{
				ISOPaths: []string{"[datastore1] iso/ubuntu.iso"},
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			expectedErrors: []string{},
		},
		{
			name: "Valid datastore path - file does not exist",
			config: &CDRomConfig{
				ISOPaths: []string{"[datastore1] iso/ubuntu.iso"},
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErrors: []string{"ISO file not found: '[datastore1] iso/ubuntu.iso'"},
		},
		{
			name: "Invalid datastore path format",
			config: &CDRomConfig{
				ISOPaths: []string{"invalid-path-format"},
			},
			driverMock:     driver.NewDriverMock(),
			expectedErrors: []string{"unable to parse datastore path: 'invalid-path-format'"},
		},
		{
			name: "Datastore not found",
			config: &CDRomConfig{
				ISOPaths: []string{"[nonexistent] iso/ubuntu.iso"},
			},
			driverMock: &driver.DriverMock{
				FindDatastoreErr: fmt.Errorf("datastore not found"),
			},
			expectedErrors: []string{"unable to access datastore 'nonexistent' for ISO validation: datastore not found"},
		},
		{
			name: "Valid content library path - resolves successfully",
			config: &CDRomConfig{
				ISOPaths: []string{"Library/Item/file.iso"},
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			expectedErrors: []string{},
		},
		{
			name: "Content library path - library not found",
			config: &CDRomConfig{
				ISOPaths: []string{"NonexistentLibrary/Item/file.iso"},
			},
			driverMock:     &driver.DriverMock{},
			expectedErrors: []string{"content library not found: 'NonexistentLibrary'"},
		},
		{
			name: "Empty or whitespace-only path",
			config: &CDRomConfig{
				ISOPaths: []string{"", "   ", "\t\n"},
			},
			driverMock: driver.NewDriverMock(),
			expectedErrors: []string{
				"ISO path cannot be empty or whitespace-only",
				"ISO path cannot be empty or whitespace-only",
				"ISO path cannot be empty or whitespace-only",
			},
		},
		{
			name: "Multiple paths with mixed results",
			config: &CDRomConfig{
				ISOPaths: []string{
					"[datastore1] iso/valid.iso",
					"[datastore1] iso/missing.iso",
					"",
					"invalid-format",
				},
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false, // Will be used for both paths
				},
			},
			expectedErrors: []string{
				"ISO file not found: '[datastore1] iso/valid.iso'",
				"ISO file not found: '[datastore1] iso/missing.iso'",
				"ISO path cannot be empty or whitespace-only",
				"unable to parse datastore path: 'invalid-format'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.driverMock != nil {
				for _, isoPath := range tt.config.ISOPaths {
					parts := strings.Split(strings.TrimLeft(isoPath, "/"), "/")
					if len(parts) == 3 {
						if parts[0] == "NonexistentLibrary" {
							tt.driverMock.FindContentLibraryFileDatastorePathErr = fmt.Errorf("library not found")
						} else {
							tt.driverMock.FindContentLibraryFileDatastorePathReturn = "[datastore1] resolved/path.iso"
						}
					}
				}
			}

			errors := tt.config.validateISOPaths(tt.driverMock)

			if len(errors) != len(tt.expectedErrors) {
				t.Fatalf("Expected %d errors, got %d: %v", len(tt.expectedErrors), len(errors), errors)
			}

			for i, expectedErr := range tt.expectedErrors {
				if i >= len(errors) {
					t.Fatalf("Expected error %d: '%s', but got no error", i, expectedErr)
				}
				if !strings.Contains(errors[i].Error(), expectedErr) {
					t.Errorf("Expected error %d to contain '%s', got '%s'", i, expectedErr, errors[i].Error())
				}
			}
		})
	}
}

func TestCDRomConfig_ValidateDatastorePath(t *testing.T) {
	tests := []struct {
		name        string
		config      *CDRomConfig
		isoPath     string
		driverMock  *driver.DriverMock
		expectedErr string
		shouldFail  bool
	}{
		{
			name:    "Valid datastore path with existing file",
			config:  &CDRomConfig{},
			isoPath: "[datastore1] iso/ubuntu.iso",
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			shouldFail: false,
		},
		{
			name:    "Valid datastore path with missing file",
			config:  &CDRomConfig{},
			isoPath: "[datastore1] iso/missing.iso",
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErr: "ISO file not found: '[datastore1] iso/missing.iso'",
			shouldFail:  true,
		},
		{
			name:        "Invalid datastore path format",
			config:      &CDRomConfig{},
			isoPath:     "invalid-format",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "unable to parse datastore path: 'invalid-format'",
			shouldFail:  true,
		},
		{
			name:    "Datastore not found",
			config:  &CDRomConfig{},
			isoPath: "[nonexistent] iso/file.iso",
			driverMock: &driver.DriverMock{
				FindDatastoreErr: fmt.Errorf("datastore 'nonexistent' not found"),
			},
			expectedErr: "unable to access datastore 'nonexistent' for ISO validation",
			shouldFail:  true,
		},
		{
			name:        "Datastore path without brackets",
			config:      &CDRomConfig{},
			isoPath:     "iso/ubuntu.iso",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "unable to parse datastore path: 'iso/ubuntu.iso'",
			shouldFail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateDatastorePath(tt.isoPath, tt.driverMock)

			if tt.shouldFail {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestCDRomConfig_ValidateContentLibraryPath(t *testing.T) {
	tests := []struct {
		name                string
		config              *CDRomConfig
		isoPath             string
		driverMock          *driver.DriverMock
		mockResolvedPath    string
		mockResolutionError error
		mockFileExists      bool
		expectedErr         string
		shouldFail          bool
	}{
		{
			name:             "Valid content library path - successful resolution and file exists",
			config:           &CDRomConfig{},
			isoPath:          "MyLibrary/UbuntuItem/ubuntu.iso",
			mockResolvedPath: "[datastore1] resolved/ubuntu.iso",
			mockFileExists:   true,
			shouldFail:       false,
		},
		{
			name:                "Content library not found",
			config:              &CDRomConfig{},
			isoPath:             "NonexistentLibrary/Item/file.iso",
			mockResolutionError: fmt.Errorf("library not found"),
			expectedErr:         "content library not found: 'NonexistentLibrary'",
			shouldFail:          true,
		},
		{
			name:                "Content library item not found",
			config:              &CDRomConfig{},
			isoPath:             "MyLibrary/NonexistentItem/file.iso",
			mockResolutionError: fmt.Errorf("item not found"),
			expectedErr:         "content library item not found: 'NonexistentItem' in library 'MyLibrary'",
			shouldFail:          true,
		},
		{
			name:        "Invalid content library path format - too few parts",
			config:      &CDRomConfig{},
			isoPath:     "Library/Item",
			expectedErr: "not a content library path format",
			shouldFail:  true,
		},
		{
			name:        "Invalid content library path format - too many parts",
			config:      &CDRomConfig{},
			isoPath:     "Library/Item/File/Extra",
			expectedErr: "not a content library path format",
			shouldFail:  true,
		},
		{
			name:             "Content library path resolves but file doesn't exist",
			config:           &CDRomConfig{},
			isoPath:          "MyLibrary/Item/missing.iso",
			mockResolvedPath: "[datastore1] resolved/missing.iso",
			mockFileExists:   false,
			expectedErr:      "content library file validation failed",
			shouldFail:       true,
		},
		{
			name:             "Content library path resolution returns same path",
			config:           &CDRomConfig{},
			isoPath:          "MyLibrary/Item/file.iso",
			mockResolvedPath: "MyLibrary/Item/file.iso", // Same as input
			expectedErr:      "content library path resolution did not return a datastore path",
			shouldFail:       true,
		},
		{
			name:                "Generic content library resolution error",
			config:              &CDRomConfig{},
			isoPath:             "MyLibrary/Item/file.iso",
			mockResolutionError: fmt.Errorf("generic resolution error"),
			expectedErr:         "content library path resolution failed",
			shouldFail:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driverMock := driver.NewDriverMock()
			if tt.driverMock != nil {
				driverMock = tt.driverMock
			}

			if driverMock.DatastoreMock == nil {
				driverMock.DatastoreMock = &driver.DatastoreMock{}
			}
			driverMock.DatastoreMock.FileExistsReturn = tt.mockFileExists
			if tt.mockResolutionError != nil {
				driverMock.FindContentLibraryFileDatastorePathErr = tt.mockResolutionError
			} else if tt.mockResolvedPath != "" {
				driverMock.FindContentLibraryFileDatastorePathReturn = tt.mockResolvedPath
			} else {
				driverMock.FindContentLibraryFileDatastorePathErr = fmt.Errorf("not identified as a Content Library path")
			}

			err := tt.config.validateContentLibraryPath(tt.isoPath, driverMock)

			if tt.shouldFail {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestCDRomConfig_ValidateContentLibraryPath_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      *CDRomConfig
		isoPath     string
		driverMock  *driver.DriverMock
		expectedErr string
		shouldFail  bool
	}{
		{
			name:    "Connection timeout",
			config:  &CDRomConfig{},
			isoPath: "Library/Item/file.iso",
			driverMock: &driver.DriverMock{
				FindContentLibraryFileDatastorePathErr: fmt.Errorf("connection timeout"),
			},
			expectedErr: "unable to connect to vCenter for content library validation: connection timeout",
			shouldFail:  true,
		},
		{
			name:    "Network error",
			config:  &CDRomConfig{},
			isoPath: "Library/Item/file.iso",
			driverMock: &driver.DriverMock{
				FindContentLibraryFileDatastorePathErr: fmt.Errorf("network unreachable"),
			},
			expectedErr: "unable to connect to vCenter for content library validation: network unreachable",
			shouldFail:  true,
		},
		{
			name:    "Returns empty string",
			config:  &CDRomConfig{},
			isoPath: "Library/Item/file.iso",
			driverMock: &driver.DriverMock{
				FindContentLibraryFileDatastorePathReturn: "",
			},
			expectedErr: "content library path resolution returned empty datastore path for 'Library/Item/file.iso'",
			shouldFail:  true,
		},
		{
			name:    "Returns whitespace only string",
			config:  &CDRomConfig{},
			isoPath: "Library/Item/file.iso",
			driverMock: &driver.DriverMock{
				FindContentLibraryFileDatastorePathReturn: "   \t\n   ",
			},
			expectedErr: "content library path resolution returned empty datastore path for 'Library/Item/file.iso'",
			shouldFail:  true,
		},
		{
			name:        "Empty item name",
			config:      &CDRomConfig{},
			isoPath:     "Library//file.iso",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "content library item name cannot be empty in path: 'Library//file.iso'",
			shouldFail:  true,
		},
		{
			name:        "Empty file name",
			config:      &CDRomConfig{},
			isoPath:     "Library/Item/",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "content library file name cannot be empty in path: 'Library/Item/'",
			shouldFail:  true,
		},
		{
			name:        "Whitespace only library name",
			config:      &CDRomConfig{},
			isoPath:     "   /Item/file.iso",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "content library name cannot be empty in path: '   /Item/file.iso'",
			shouldFail:  true,
		},
		{
			name:        "Whitespace only item name",
			config:      &CDRomConfig{},
			isoPath:     "Library/   /file.iso",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "content library item name cannot be empty in path: 'Library/   /file.iso'",
			shouldFail:  true,
		},
		{
			name:        "Whitespace only file name",
			config:      &CDRomConfig{},
			isoPath:     "Library/Item/   ",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "content library file name cannot be empty in path: 'Library/Item/   '",
			shouldFail:  true,
		},
		{
			name:        "Not content library format",
			config:      &CDRomConfig{},
			isoPath:     "/Item/file.iso",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "not a content library path format",
			shouldFail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateContentLibraryPath(tt.isoPath, tt.driverMock)

			if tt.shouldFail {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestCDRomConfig_ErrorMessageFormatting(t *testing.T) {
	tests := []struct {
		name           string
		config         *CDRomConfig
		driverMock     *driver.DriverMock
		expectedFormat string
	}{
		{
			name: "Datastore file not found error format",
			config: &CDRomConfig{
				ISOPaths: []string{"[datastore1] iso/ubuntu.iso"},
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedFormat: "ISO file not found: '[datastore1] iso/ubuntu.iso'",
		},
		{
			name: "Datastore access error format",
			config: &CDRomConfig{
				ISOPaths: []string{"[nonexistent] iso/ubuntu.iso"},
			},
			driverMock: &driver.DriverMock{
				FindDatastoreErr: fmt.Errorf("datastore not found"),
			},
			expectedFormat: "unable to access datastore 'nonexistent' for ISO validation: datastore not found",
		},
		{
			name: "Invalid path format error",
			config: &CDRomConfig{
				ISOPaths: []string{"invalid-format"},
			},
			driverMock:     driver.NewDriverMock(),
			expectedFormat: "unable to parse datastore path: 'invalid-format'",
		},
		{
			name: "Empty path error format",
			config: &CDRomConfig{
				ISOPaths: []string{""},
			},
			driverMock:     driver.NewDriverMock(),
			expectedFormat: "ISO path cannot be empty or whitespace-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.validateISOPaths(tt.driverMock)

			if len(errors) == 0 {
				t.Fatal("Expected at least one error")
			}

			if errors[0].Error() != tt.expectedFormat {
				t.Errorf("Expected error format '%s', got '%s'", tt.expectedFormat, errors[0].Error())
			}
		})
	}
}

func TestCDRomConfig_ErrorAggregation(t *testing.T) {
	config := &CDRomConfig{
		ISOPaths: []string{
			"",
			"   ",
			"invalid-format",
			"[datastore1] iso/missing1.iso",
			"[datastore1] iso/missing2.iso",
		},
	}

	driverMock := &driver.DriverMock{
		DatastoreMock: &driver.DatastoreMock{
			FileExistsReturn: false,
		},
	}

	errors := config.validateISOPaths(driverMock)

	expectedErrorCount := 5
	if len(errors) != expectedErrorCount {
		t.Fatalf("Expected %d errors, got %d: %v", expectedErrorCount, len(errors), errors)
	}

	expectedSubstrings := []string{
		"ISO path cannot be empty or whitespace-only",
		"ISO path cannot be empty or whitespace-only",
		"unable to parse datastore path: 'invalid-format'",
		"ISO file not found: '[datastore1] iso/missing1.iso'",
		"ISO file not found: '[datastore1] iso/missing2.iso'",
	}

	for i, expectedSubstring := range expectedSubstrings {
		if !strings.Contains(errors[i].Error(), expectedSubstring) {
			t.Errorf("Error %d should contain '%s', got '%s'", i, expectedSubstring, errors[i].Error())
		}
	}
}

func TestCDRomConfig_ValidateISOPaths_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		config         *CDRomConfig
		driverMock     *driver.DriverMock
		expectedErrors []string
	}{
		{
			name: "Nil driver - should handle gracefully",
			config: &CDRomConfig{
				ISOPaths: []string{"[datastore1] iso/test.iso"},
			},
			driverMock:     nil,
			expectedErrors: []string{"driver is not available for ISO validation"},
		},
		{
			name: "Empty and whitespace-only paths",
			config: &CDRomConfig{
				ISOPaths: []string{"", "   ", "\t", "\n", "  \t\n  "},
			},
			driverMock: driver.NewDriverMock(),
			expectedErrors: []string{
				"ISO path cannot be empty or whitespace-only",
				"ISO path cannot be empty or whitespace-only",
				"ISO path cannot be empty or whitespace-only",
				"ISO path cannot be empty or whitespace-only",
				"ISO path cannot be empty or whitespace-only",
			},
		},
		{
			name: "Malformed datastore paths",
			config: &CDRomConfig{
				ISOPaths: []string{
					"[incomplete",
					"incomplete]",
					"[datastore1",
					"datastore1] file.iso",
					"[datastore][] /file.iso",
					"[data/store] /file.iso",
				},
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErrors: []string{
				"invalid datastore path format: '[incomplete'",
				"invalid datastore path format: 'incomplete]'",
				"invalid datastore path format: '[datastore1'",
				"invalid datastore path format: 'datastore1] file.iso'",
				"invalid datastore path format: '[datastore][] /file.iso'",
				"content library path resolution returned empty datastore path for '[data/store] /file.iso'",
			},
		},
		{
			name: "Datastore access error",
			config: &CDRomConfig{
				ISOPaths: []string{"[datastore1] iso/test.iso"},
			},
			driverMock: &driver.DriverMock{
				FindDatastoreErr: fmt.Errorf("datastore not found"),
			},
			expectedErrors: []string{"unable to access datastore 'datastore1' for ISO validation: datastore not found"},
		},

		{
			name: "Mixed edge cases",
			config: &CDRomConfig{
				ISOPaths: []string{
					"",
					"[datastore][] /file.iso",
					"Library//file.iso",
					"[datastore1] valid.iso",
				},
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			expectedErrors: []string{
				"ISO path cannot be empty or whitespace-only",
				"invalid datastore path format: '[datastore][] /file.iso'",
				"content library item name cannot be empty in path: 'Library//file.iso'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errors []error
			if tt.driverMock == nil {
				errors = tt.config.validateISOPaths(nil)
			} else {
				errors = tt.config.validateISOPaths(tt.driverMock)
			}

			if len(errors) != len(tt.expectedErrors) {
				t.Fatalf("Expected %d errors, got %d: %v", len(tt.expectedErrors), len(errors), errors)
			}

			for i, expectedErr := range tt.expectedErrors {
				if i >= len(errors) {
					t.Fatalf("Expected error %d: '%s', but got no error", i, expectedErr)
				}
				if !strings.Contains(errors[i].Error(), expectedErr) {
					t.Errorf("Expected error %d to contain '%s', got '%s'", i, expectedErr, errors[i].Error())
				}
			}
		})
	}
}

func TestCDRomConfig_ValidateDatastorePath_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      *CDRomConfig
		isoPath     string
		driverMock  *driver.DriverMock
		expectedErr string
		shouldFail  bool
	}{
		{
			name:    "Datastore access error",
			config:  &CDRomConfig{},
			isoPath: "[datastore1] iso/test.iso",
			driverMock: &driver.DriverMock{
				FindDatastoreErr: fmt.Errorf("datastore not accessible"),
			},
			expectedErr: "unable to access datastore 'datastore1' for ISO validation: datastore not accessible",
			shouldFail:  true,
		},
		{
			name:    "Datastore connection timeout",
			config:  &CDRomConfig{},
			isoPath: "[datastore1] iso/test.iso",
			driverMock: &driver.DriverMock{
				FindDatastoreErr: fmt.Errorf("connection timeout to datastore"),
			},
			expectedErr: "unable to access datastore 'datastore1' for ISO validation: connection timeout to datastore",
			shouldFail:  true,
		},
		{
			name:    "Datastore network error",
			config:  &CDRomConfig{},
			isoPath: "[datastore1] iso/test.iso",
			driverMock: &driver.DriverMock{
				FindDatastoreErr: fmt.Errorf("network unreachable"),
			},
			expectedErr: "unable to access datastore 'datastore1' for ISO validation: network unreachable",
			shouldFail:  true,
		},
		{
			name:        "Completely malformed path - no brackets",
			config:      &CDRomConfig{},
			isoPath:     "just-a-filename.iso",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "unable to parse datastore path: 'just-a-filename.iso'",
			shouldFail:  true,
		},
		{
			name:        "Path with only opening bracket",
			config:      &CDRomConfig{},
			isoPath:     "[datastore1 file.iso",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "invalid datastore path format: '[datastore1 file.iso'",
			shouldFail:  true,
		},
		{
			name:        "Path with only closing bracket",
			config:      &CDRomConfig{},
			isoPath:     "datastore1] file.iso",
			driverMock:  driver.NewDriverMock(),
			expectedErr: "invalid datastore path format: 'datastore1] file.iso'",
			shouldFail:  true,
		},
		{
			name:    "Empty datastore name - valid format but file doesn't exist",
			config:  &CDRomConfig{},
			isoPath: "[] file.iso",
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErr: "ISO file not found: '[] file.iso'",
			shouldFail:  true,
		},
		{
			name:    "Whitespace-only datastore name - valid format but file doesn't exist",
			config:  &CDRomConfig{},
			isoPath: "[   ] file.iso",
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErr: "ISO file not found: '[   ] file.iso'",
			shouldFail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateDatastorePath(tt.isoPath, tt.driverMock)

			if tt.shouldFail {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestCDRomConfig_PrepareWithMockDriver(t *testing.T) {
	tests := []struct {
		name                    string
		config                  *CDRomConfig
		keepConfig              *ReattachCDRomConfig
		driverMock              *driver.DriverMock
		expectedErrors          []string
		expectedDriverCalls     map[string]bool
		expectedDatastoreCalls  map[string]interface{}
		expectedContentLibCalls map[string]interface{}
	}{
		{
			name: "No ISO paths configured",
			config: &CDRomConfig{
				CdromType: "ide",
				ISOPaths:  []string{},
			},
			keepConfig:     &ReattachCDRomConfig{ReattachCDRom: 2},
			driverMock:     driver.NewDriverMock(),
			expectedErrors: []string{},
			expectedDriverCalls: map[string]bool{
				"FindDatastore": false,
			},
			expectedDatastoreCalls: map[string]interface{}{
				"FileExistsCalled": false,
			},
		},
		{
			name: "Valid datastore ISO path",
			config: &CDRomConfig{
				CdromType: "ide",
				ISOPaths:  []string{"[datastore1] iso/ubuntu.iso"},
			},
			keepConfig: &ReattachCDRomConfig{ReattachCDRom: 2},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			expectedErrors: []string{},
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
			expectedDatastoreCalls: map[string]interface{}{
				"FileExistsCalled": true,
			},
		},
		{
			name: "Missing datastore ISO file",
			config: &CDRomConfig{
				CdromType: "ide",
				ISOPaths:  []string{"[datastore1] iso/missing.iso"},
			},
			keepConfig: &ReattachCDRomConfig{ReattachCDRom: 2},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErrors: []string{"ISO file not found: '[datastore1] iso/missing.iso'"},
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
			expectedDatastoreCalls: map[string]interface{}{
				"FileExistsCalled": true,
			},
		},
		{
			name: "Datastore not found",
			config: &CDRomConfig{
				CdromType: "ide",
				ISOPaths:  []string{"[nonexistent] iso/file.iso"},
			},
			keepConfig: &ReattachCDRomConfig{ReattachCDRom: 2},
			driverMock: &driver.DriverMock{
				FindDatastoreErr: fmt.Errorf("datastore 'nonexistent' not found"),
			},
			expectedErrors: []string{"unable to access datastore 'nonexistent' for ISO validation: datastore 'nonexistent' not found"},
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
			expectedDatastoreCalls: map[string]interface{}{
				"FileExistsCalled": false,
			},
		},
		{
			name: "Valid content library path",
			config: &CDRomConfig{
				CdromType: "sata",
				ISOPaths:  []string{"MyLibrary/UbuntuItem/ubuntu.iso"},
			},
			keepConfig: &ReattachCDRomConfig{ReattachCDRom: 1},
			driverMock: &driver.DriverMock{
				FindContentLibraryFileDatastorePathReturn: "[datastore1] resolved/ubuntu.iso",
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			expectedErrors: []string{},
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
			expectedDatastoreCalls: map[string]interface{}{
				"FileExistsCalled": true,
			},
			expectedContentLibCalls: map[string]interface{}{
				"FindContentLibraryFileDatastorePathCalled": true,
			},
		},
		{
			name: "Content library not found",
			config: &CDRomConfig{
				CdromType: "ide",
				ISOPaths:  []string{"NonexistentLibrary/Item/file.iso"},
			},
			keepConfig: &ReattachCDRomConfig{ReattachCDRom: 3},
			driverMock: &driver.DriverMock{
				FindContentLibraryFileDatastorePathErr: fmt.Errorf("library not found"),
			},
			expectedErrors: []string{"content library not found: 'NonexistentLibrary'"},
			expectedDriverCalls: map[string]bool{
				"FindDatastore": false,
			},
			expectedDatastoreCalls: map[string]interface{}{
				"FileExistsCalled": false,
			},
			expectedContentLibCalls: map[string]interface{}{
				"FindContentLibraryFileDatastorePathCalled": true,
			},
		},
		{
			name: "Multiple ISO validation errors with cdrom_type error",
			config: &CDRomConfig{
				CdromType: "invalid",
				ISOPaths: []string{
					"",
					"[datastore1] iso/missing.iso",
					"invalid-format",
				},
			},
			keepConfig: &ReattachCDRomConfig{ReattachCDRom: 5},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErrors: []string{
				"ISO path cannot be empty or whitespace-only",
				"ISO file not found: '[datastore1] iso/missing.iso'",
				"unable to parse datastore path: 'invalid-format'",
				"'cdrom_type' must be 'ide' or 'sata'",
				"'reattach_cdroms' should be between 1 and 4",
			},
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
			expectedDatastoreCalls: map[string]interface{}{
				"FileExistsCalled": true,
			},
		},
		{
			name: "Mixed datastore and content library paths with some missing",
			config: &CDRomConfig{
				CdromType: "sata",
				ISOPaths: []string{
					"[datastore1] iso/missing.iso",
				},
			},
			keepConfig: &ReattachCDRomConfig{ReattachCDRom: 2},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErrors: []string{"ISO file not found: '[datastore1] iso/missing.iso'"},
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
			expectedDatastoreCalls: map[string]interface{}{
				"FileExistsCalled": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Prepare(tt.keepConfig, tt.driverMock)

			if len(errors) != len(tt.expectedErrors) {
				t.Fatalf("Expected %d errors, got %d: %v", len(tt.expectedErrors), len(errors), errors)
			}

			for i, expectedErr := range tt.expectedErrors {
				if i >= len(errors) {
					t.Fatalf("Expected error %d: '%s', but got no error", i, expectedErr)
				}
				if !strings.Contains(errors[i].Error(), expectedErr) {
					t.Errorf("Expected error %d to contain '%s', got '%s'", i, expectedErr, errors[i].Error())
				}
			}

			for method, expectedCalled := range tt.expectedDriverCalls {
				switch method {
				case "FindDatastore":
					if tt.driverMock.FindDatastoreCalled != expectedCalled {
						t.Errorf("Expected FindDatastore called: %v, got: %v", expectedCalled, tt.driverMock.FindDatastoreCalled)
					}
				}
			}

			if tt.driverMock.DatastoreMock != nil {
				for method, expectedValue := range tt.expectedDatastoreCalls {
					switch method {
					case "FileExistsCalled":
						if tt.driverMock.DatastoreMock.FileExistsCalled != expectedValue.(bool) {
							t.Errorf("Expected DatastoreMock.FileExistsCalled: %v, got: %v", expectedValue, tt.driverMock.DatastoreMock.FileExistsCalled)
						}
					}
				}
			}

			for method, expectedValue := range tt.expectedContentLibCalls {
				switch method {
				case "FindContentLibraryFileDatastorePathCalled":
					if tt.driverMock.FindContentLibraryFileDatastorePathCalled != expectedValue.(bool) {
						t.Errorf("Expected FindContentLibraryFileDatastorePathCalled: %v, got: %v", expectedValue, tt.driverMock.FindContentLibraryFileDatastorePathCalled)
					}
				}
			}

		})
	}
}

func TestStepAddCDRom_RunWithISOValidation(t *testing.T) {
	tests := []struct {
		name                string
		step                *StepAddCDRom
		state               *multistep.BasicStateBag
		driverMock          *driver.DriverMock
		vmMock              *driver.VirtualMachineMock
		expectedAction      multistep.StepAction
		expectedError       string
		shouldHaveError     bool
		expectedDriverCalls map[string]bool
	}{
		{
			name: "ISO validation passes, step continues normally",
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "ide",
					ISOPaths:  []string{"[datastore1] iso/ubuntu.iso"},
				},
			},
			state: func() *multistep.BasicStateBag {
				state := basicStateBag(nil)
				return state
			}(),
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			vmMock:          new(driver.VirtualMachineMock),
			expectedAction:  multistep.ActionContinue,
			shouldHaveError: false,
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
		},
		{
			name: "ISO validation fails, step halts",
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "ide",
					ISOPaths:  []string{"[datastore1] iso/missing.iso"},
				},
			},
			state: func() *multistep.BasicStateBag {
				state := basicStateBag(nil)
				return state
			}(),
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			vmMock:          new(driver.VirtualMachineMock),
			expectedAction:  multistep.ActionHalt,
			expectedError:   "ISO validation failed: ISO file not found: '[datastore1] iso/missing.iso'",
			shouldHaveError: true,
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
		},
		{
			name: "Multiple ISO validation errors aggregated",
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "sata",
					ISOPaths: []string{
						"",
						"[datastore1] iso/missing1.iso",
						"[datastore1] iso/missing2.iso",
					},
				},
			},
			state: func() *multistep.BasicStateBag {
				state := basicStateBag(nil)
				return state
			}(),
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			vmMock:          new(driver.VirtualMachineMock),
			expectedAction:  multistep.ActionHalt,
			expectedError:   "ISO validation failed:",
			shouldHaveError: true,
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
		},
		{
			name: "No driver in state, validation skipped",
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "ide",
					ISOPaths:  []string{"[datastore1] iso/any.iso"},
				},
			},
			state: func() *multistep.BasicStateBag {
				state := basicStateBag(nil)
				return state
			}(),
			driverMock:      nil, // No driver
			vmMock:          new(driver.VirtualMachineMock),
			expectedAction:  multistep.ActionContinue,
			shouldHaveError: false,
			expectedDriverCalls: map[string]bool{
				"FindDatastore": false,
			},
		},
		{
			name: "Content library path validation in runtime",
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "sata",
					ISOPaths:  []string{"MyLibrary/Item/file.iso"},
				},
			},
			state: func() *multistep.BasicStateBag {
				state := basicStateBag(nil)
				return state
			}(),
			driverMock: &driver.DriverMock{
				FindContentLibraryFileDatastorePathReturn: "[datastore1] resolved/file.iso",
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			vmMock:          &driver.VirtualMachineMock{},
			expectedAction:  multistep.ActionContinue,
			shouldHaveError: false,
			expectedDriverCalls: map[string]bool{
				"FindDatastore": true,
			},
		},
		{
			name: "Content library path fails validation in runtime",
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "ide",
					ISOPaths:  []string{"NonexistentLibrary/Item/file.iso"},
				},
			},
			state: func() *multistep.BasicStateBag {
				state := basicStateBag(nil)
				return state
			}(),
			driverMock: &driver.DriverMock{
				FindContentLibraryFileDatastorePathErr: fmt.Errorf("library not found"),
			},
			vmMock:          new(driver.VirtualMachineMock),
			expectedAction:  multistep.ActionHalt,
			expectedError:   "ISO validation failed: content library not found: 'NonexistentLibrary'",
			shouldHaveError: true,
			expectedDriverCalls: map[string]bool{
				"FindDatastore": false, // Won't be called due to content library error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.Put("vm", tt.vmMock)
			if tt.driverMock != nil {
				tt.state.Put("driver", tt.driverMock)
			}

			action := tt.step.Run(context.TODO(), tt.state)

			if action != tt.expectedAction {
				t.Fatalf("Expected action %v, got %v", tt.expectedAction, action)
			}
			err, hasError := tt.state.Get("error").(error)
			if tt.shouldHaveError {
				if !hasError {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if hasError {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}

			if tt.driverMock != nil {
				for method, expectedCalled := range tt.expectedDriverCalls {
					switch method {
					case "FindDatastore":
						if tt.driverMock.FindDatastoreCalled != expectedCalled {
							t.Errorf("Expected FindDatastore called: %v, got: %v", expectedCalled, tt.driverMock.FindDatastoreCalled)
						}
					}
				}
			}
		})
	}
}

func TestCDRomConfig_PrepareErrorCollection(t *testing.T) {
	tests := []struct {
		name           string
		config         *CDRomConfig
		keepConfig     *ReattachCDRomConfig
		driverMock     *driver.DriverMock
		expectedErrors []string
		description    string
	}{
		{
			name: "ISO validation errors combined with existing validation errors",
			config: &CDRomConfig{
				CdromType: "invalid", // This will cause a cdrom_type validation error
				ISOPaths: []string{
					"[datastore1] iso/missing.iso", // This will cause an ISO validation error
				},
			},
			keepConfig: &ReattachCDRomConfig{
				ReattachCDRom: 5, // This will cause a reattach_cdroms validation error
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErrors: []string{
				"ISO file not found: '[datastore1] iso/missing.iso'",
				"'cdrom_type' must be 'ide' or 'sata'",
				"'reattach_cdroms' should be between 1 and 4",
			},
			description: "Verifies that ISO validation errors are properly integrated with existing validation",
		},
		{
			name: "ISO validation passes, other validations fail",
			config: &CDRomConfig{
				CdromType: "invalid",
				ISOPaths:  []string{"[datastore1] iso/valid.iso"},
			},
			keepConfig: &ReattachCDRomConfig{
				ReattachCDRom: -1,
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			expectedErrors: []string{
				"'cdrom_type' must be 'ide' or 'sata'",
				"'reattach_cdroms' should be between 1 and 4",
			},
			description: "Verifies that when ISO validation passes, other validation errors are still reported",
		},
		{
			name: "All validations pass",
			config: &CDRomConfig{
				CdromType: "ide",
				ISOPaths:  []string{"[datastore1] iso/valid.iso"},
			},
			keepConfig: &ReattachCDRomConfig{
				ReattachCDRom: 2,
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: true,
				},
			},
			expectedErrors: []string{},
			description:    "Verifies that when all validations pass, no errors are returned",
		},
		{
			name: "Multiple ISO validation errors with valid other config",
			config: &CDRomConfig{
				CdromType: "sata",
				ISOPaths: []string{
					"",
					"[datastore1] iso/missing1.iso",
					"invalid-format",
					"[datastore1] iso/missing2.iso",
				},
			},
			keepConfig: &ReattachCDRomConfig{
				ReattachCDRom: 3,
			},
			driverMock: &driver.DriverMock{
				DatastoreMock: &driver.DatastoreMock{
					FileExistsReturn: false,
				},
			},
			expectedErrors: []string{
				"ISO path cannot be empty or whitespace-only",
				"ISO file not found: '[datastore1] iso/missing1.iso'",
				"unable to parse datastore path: 'invalid-format'",
				"ISO file not found: '[datastore1] iso/missing2.iso'",
			},
			description: "Verifies that multiple ISO validation errors are all collected and reported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Prepare(tt.keepConfig, tt.driverMock)

			if len(errors) != len(tt.expectedErrors) {
				t.Fatalf("%s: Expected %d errors, got %d: %v", tt.description, len(tt.expectedErrors), len(errors), errors)
			}
			for i, expectedErr := range tt.expectedErrors {
				if i >= len(errors) {
					t.Fatalf("%s: Expected error %d: '%s', but got no error", tt.description, i, expectedErr)
				}
				if !strings.Contains(errors[i].Error(), expectedErr) {
					t.Errorf("%s: Expected error %d to contain '%s', got '%s'", tt.description, i, expectedErr, errors[i].Error())
				}
			}
		})
	}
}
