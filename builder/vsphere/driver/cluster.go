package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type Cluster struct {
	driver  *VCenterDriver
	cluster *object.ClusterComputeResource
}

func (d *VCenterDriver) NewCluster(ref *types.ManagedObjectReference) *Cluster {
	return &Cluster{
		cluster: object.NewClusterComputeResource(d.client.Client, *ref),
		driver:  d,
	}
}

func (d *VCenterDriver) FindCluster(name string) (*Cluster, error) {
	c, err := d.finder.ClusterComputeResource(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &Cluster{
		cluster: c,
		driver:  d,
	}, nil
}
