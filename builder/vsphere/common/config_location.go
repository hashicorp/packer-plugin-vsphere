// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type LocationConfig

package common

import (
	"fmt"
	"path"
	"strings"
)

type LocationConfig struct {
	// The name of the virtual machine.
	VMName string `mapstructure:"vm_name"`
	// The virtual machine folder where the virtual machine is created.
	Folder string `mapstructure:"folder"`
	// The cluster where the virtual machine is created.
	// Refer to the [Working With Clusters And Hosts](#working-with-clusters-and-hosts)
	// section for more details.
	Cluster string `mapstructure:"cluster"`
	// The ESXi host where the virtual machine is created. A full path must be
	// specified if the ESXi host is in a folder. For example `folder/host`.
	// Refer to the [Working With Clusters And Hosts](#working-with-clusters-and-hosts)
	// section for more details.
	Host string `mapstructure:"host"`
	// The resource pool where the virtual machine is created.
	// If this is not specified, the root resource pool associated with the
	// `host` or `cluster` is used.
	//
	// ~> **Note:**  The full path to the resource pool must be provided.
	// For example, a simple resource pool path might resemble `rp-packer` and
	// a nested path might resemble 'rp-packer/rp-linux-images'.
	ResourcePool string `mapstructure:"resource_pool"`
	// The datastore where the virtual machine is created.
	// Required if `host` is a cluster, or if `host` has multiple datastores.
	Datastore string `mapstructure:"datastore"`
	// The ESXI host used for uploading files to the datastore.
	// Defaults to `false`.
	SetHostForDatastoreUploads bool `mapstructure:"set_host_for_datastore_uploads"`
}

func (c *LocationConfig) Prepare() []error {
	var errs []error

	if c.VMName == "" {
		errs = append(errs, fmt.Errorf("'vm_name' is required"))
	}
	if c.Cluster == "" && c.Host == "" {
		errs = append(errs, fmt.Errorf("'host' or 'cluster' is required"))
	}

	// clean Folder path and remove leading slash as folders are relative within vsphere
	c.Folder = path.Clean(c.Folder)
	c.Folder = strings.TrimLeft(c.Folder, "/")

	return errs
}
