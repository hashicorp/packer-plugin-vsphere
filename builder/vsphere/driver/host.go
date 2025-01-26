// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Host struct {
	driver *VCenterDriver
	host   *object.HostSystem
}

// NewHost creates and initializes a new Host object using a
// ManagedObjectReference and the VCenterDriver instance.
func (d *VCenterDriver) NewHost(ref *types.ManagedObjectReference) *Host {
	return &Host{
		host:   object.NewHostSystem(d.client.Client, *ref),
		driver: d,
	}
}

// FindHost locates a host within the vCenter environment by its name. Returns
// a Host object or an error if not found or if the retrieval process fails.
func (d *VCenterDriver) FindHost(name string) (*Host, error) {
	h, err := d.finder.HostSystem(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &Host{
		host:   h,
		driver: d,
	}, nil
}

// Info retrieves properties of the host object with optional filters specified
// as parameters. If no parameters are provided, all properties are returned.
func (h *Host) Info(params ...string) (*mo.HostSystem, error) {
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var info mo.HostSystem
	err := h.host.Properties(h.driver.ctx, h.host.Reference(), p, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
