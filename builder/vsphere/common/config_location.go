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
	// Name of the virtual machine.
	VMName string `mapstructure:"vm_name"`
	// VM folder where the virtual machine is created.
	Folder string `mapstructure:"folder"`
	// vSphere cluster where the virtual machine is created. See the
	// [Working With Clusters And Hosts](#working-with-clusters-and-hosts)
	// section above for more details.
	Cluster string `mapstructure:"cluster"`
	// ESXi host where the virtual machine is created. A full path must be
	// specified if the host is in a folder. For example `folder/host`. See the
	// [Working With Clusters And Hosts](#working-with-clusters-and-hosts)
	// section above for more details.
	Host string `mapstructure:"host"`
	// vSphere resource pool where the virtual machine is created.
	// If this is not specified, the root resource pool associated with the
	// `host` or `cluster` is used.
	// Note that the full path to the resource pool must be provided.
	// For example, a simple resource pool path might resemble `rp-packer` and
	// a nested path might resemble 'rp-packer/rp-linux-images'.
	ResourcePool string `mapstructure:"resource_pool"`
	// vSphere datastore where the virtual machine is created.
	// Required if `host` is a cluster, or if `host` has multiple datastores.
	Datastore string `mapstructure:"datastore"`
	// Specifies that the host is used for uploading files to the datastore.
	// Defaults to false.
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
