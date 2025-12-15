// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"fmt"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Network struct {
	driver  *VCenterDriver
	network object.NetworkReference
}

// NewNetwork creates and initializes a new Network object using the provided
// ManagedObjectReference.
func (d *VCenterDriver) NewNetwork(ref *types.ManagedObjectReference) *Network {
	return &Network{
		network: object.NewNetwork(d.Client.Client, *ref),
		driver:  d,
	}
}

// FindNetwork locates a network by its name within the vCenter context.
// Returns a Network object or an error if the network is not found.
func (d *VCenterDriver) FindNetwork(name string) (*Network, error) {
	n, err := d.Finder.Network(d.Ctx, name)
	if err != nil {
		return nil, err
	}
	return &Network{
		network: n,
		driver:  d,
	}, nil
}

// FindNetworks retrieves a list of networks in the vCenter matching the
// provided name and returns them as Network objects.
func (d *VCenterDriver) FindNetworks(name string) ([]*Network, error) {
	ns, err := d.Finder.NetworkList(d.Ctx, name)
	if err != nil {
		return nil, err
	}
	var networks []*Network
	for _, n := range ns {
		networks = append(networks, &Network{
			network: n,
			driver:  d,
		})
	}
	return networks, nil
}

// Info retrieves the properties of the network object with optional filters
// specified as parameters. If no parameters are provided, all properties are
// returned.
func (n *Network) Info(params ...string) (*mo.Network, error) {
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var info mo.Network

	network, ok := n.network.(*object.Network)
	if !ok {
		return nil, fmt.Errorf("unexpected %t network object type", n.network)
	}

	err := network.Properties(n.driver.Ctx, network.Reference(), p, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

type MultipleNetworkFoundError struct {
	path   string
	append string
}

// Error returns a formatted error message for the MultipleNetworkFoundError.
func (e *MultipleNetworkFoundError) Error() string {
	return fmt.Sprintf("'%s' resolves to more than one network name; %s", e.path, e.append)
}
