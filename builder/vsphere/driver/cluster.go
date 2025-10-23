// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import "github.com/vmware/govmomi/object"

type Cluster struct {
	driver  *VCenterDriver
	cluster *object.ClusterComputeResource
}

// FindCluster locates a cluster within the vCenter environment by its name.
// Returns a Cluster object or an error if not found or if the retrieval
// process fails.
func (d *VCenterDriver) FindCluster(name string) (*Cluster, error) {
	c, err := d.Finder.ClusterComputeResource(d.Ctx, name)
	if err != nil {
		return nil, err
	}
	return &Cluster{
		cluster: c,
		driver:  d,
	}, nil
}
