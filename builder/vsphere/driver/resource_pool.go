// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"fmt"
	"log"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ResourcePool struct {
	pool   *object.ResourcePool
	driver *VCenterDriver
}

// NewResourcePool creates and returns a new ResourcePool object using the
// provided ManagedObjectReference.
func (d *VCenterDriver) NewResourcePool(ref *types.ManagedObjectReference) *ResourcePool {
	return &ResourcePool{
		pool:   object.NewResourcePool(d.client.Client, *ref),
		driver: d,
	}
}

// FindResourcePool locates a resource pool by its name within a specified
// cluster or host context in vCenter. It falls back to the default resource
// pool or a vApp if the specified pool is not found. Returns a ResourcePool
// object or an error if neither the specified nor default pool is accessible.
func (d *VCenterDriver) FindResourcePool(cluster string, host string, name string) (*ResourcePool, error) {
	var res string
	if cluster != "" {
		res = cluster
	} else {
		res = host
	}

	resourcePath := fmt.Sprintf("%v/Resources/%v", res, name)
	p, err := d.finder.ResourcePool(d.ctx, resourcePath)
	if err != nil {
		log.Printf("[WARN] %s not found. Looking for default resource pool.", resourcePath)
		dp, dperr := d.finder.DefaultResourcePool(d.ctx)
		if _, ok := dperr.(*find.NotFoundError); ok {
			vapp, verr := d.finder.VirtualApp(d.ctx, name)
			if verr != nil {
				return nil, err
			}
			dp = vapp.ResourcePool
		} else if dperr != nil {
			return nil, err
		}
		p = dp
	}

	return &ResourcePool{
		pool:   p,
		driver: d,
	}, nil
}

// Info retrieves the properties of the ResourcePool object with optional
// filters specified as parameters. If no parameters are provided, all
// properties are returned.
func (p *ResourcePool) Info(params ...string) (*mo.ResourcePool, error) {
	var params2 []string
	if len(params) == 0 {
		params2 = []string{"*"}
	} else {
		params2 = params
	}
	var info mo.ResourcePool
	err := p.pool.Properties(p.driver.ctx, p.pool.Reference(), params2, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// Path returns the full hierarchical path of the ResourcePool or an empty
// string if it's a top-level entity. It recursively resolves the parent's
// path until reaching the root or a top-level parent.
func (p *ResourcePool) Path() (string, error) {
	poolInfo, err := p.Info("name", "parent")
	if err != nil {
		return "", err
	}
	if poolInfo.Parent.Type == "ComputeResource" || poolInfo.Parent.Type == "ClusterComputeResource" {
		return "", nil
	} else {
		parent := p.driver.NewResourcePool(poolInfo.Parent)
		parentPath, err := parent.Path()
		if err != nil {
			return "", err
		}
		if parentPath == "" {
			return poolInfo.Name, nil
		} else {
			return fmt.Sprintf("%v/%v", parentPath, poolInfo.Name), nil
		}
	}
}
