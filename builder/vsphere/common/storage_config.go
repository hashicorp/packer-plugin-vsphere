// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type StorageConfig,DiskConfig

package common

import (
	"fmt"
)

// The following example that will create a 15GB and a 20GB disk on the virtual
// machine. The second disk will be thin provisioned:
//
// HCL Example:
//
// ```hcl
//
//	storage {
//	    disk_size = 15000
//	}
//	storage {
//	    disk_size = 20000
//	    disk_thin_provisioned = true
//	}
//
// ```
//
// JSON Example:
//
// ```json
//
//	"storage": [
//	  {
//	    "disk_size": 15000
//	  },
//	  {
//	    "disk_size": 20000,
//	    "disk_thin_provisioned": true
//	  }
//	],
//
// ```
//
// The following example will use two PVSCSI controllers and two disks on each
// controller.
//
// HCL Example:
//
// ```hcl
//
//	 disk_controller_type = ["pvscsi", "pvscsi"]
//		storage {
//		   disk_size = 15000
//		   disk_controller_index = 0
//		}
//		storage {
//		   disk_size = 15000
//		   disk_controller_index = 0
//		}
//		storage {
//		   disk_size = 15000
//		   disk_controller_index = 1
//		}
//		storage {
//		   disk_size = 15000
//		   disk_controller_index = 1
//		}
//
// ```
//
// JSON Example:
//
// ```json
//
//	"disk_controller_type": ["pvscsi", "pvscsi"],
//	"storage": [
//	  {
//	    "disk_size": 15000,
//	    "disk_controller_index": 0
//	  },
//	  {
//	    "disk_size": 15000,
//	    "disk_controller_index": 0
//	  },
//	  {
//	    "disk_size": 15000,
//	    "disk_controller_index": 1
//	  },
//	  {
//	    "disk_size": 15000,
//	    "disk_controller_index": 1
//	  }
//	],
//
// ```
type DiskConfig struct {
	// The size of the disk in MiB.
	DiskSize int64 `mapstructure:"disk_size" required:"true"`
	// Enable thin provisioning for the disk.
	// Defaults to `false`.
	DiskThinProvisioned bool `mapstructure:"disk_thin_provisioned"`
	// Enable eager scrubbing for the disk.
	// Defaults to `false`.
	DiskEagerlyScrub bool `mapstructure:"disk_eagerly_scrub"`
	// The assigned disk controller for the disk.
	// Defaults to the first controller, `(0)`.
	DiskControllerIndex int `mapstructure:"disk_controller_index"`
}

type StorageConfig struct {
	// The disk controller type. One of `lsilogic`, `lsilogic-sas`, `pvscsi`,
	// `nvme`, `scsi`, or `sata`. Defaults to `lsilogic`. Use a list to define
	// additional controllers. Refer to [SCSI, SATA, and NVMe Storage Controller
	// Conditions, Limitations, and Compatibility](https://docs.vmware.com/en/VMware-vSphere/8.0/vsphere-vm-administration/GUID-5872D173-A076-42FE-8D0B-9DB0EB0E7362.html)
	// for additional information.
	DiskControllerType []string `mapstructure:"disk_controller_type"`
	// A collection of one or more disks to be provisioned.
	// Refer to the [Storage Configuration](#storage-configuration) section for additional information.
	Storage []DiskConfig `mapstructure:"storage"`
}

func (c *StorageConfig) Prepare() []error {
	var errs []error

	if len(c.Storage) > 0 {
		for i, storage := range c.Storage {
			if storage.DiskSize == 0 {
				errs = append(errs, fmt.Errorf("storage[%d].'disk_size' is required", i))
			}
			if storage.DiskControllerIndex >= len(c.DiskControllerType) {
				errs = append(errs, fmt.Errorf("storage[%d].'disk_controller_index' references an unknown disk controller", i))
			}
		}
	}

	return errs
}
