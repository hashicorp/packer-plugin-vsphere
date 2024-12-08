// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/rest"
)

type VCenterDriver struct {
	Ctx        context.Context
	Client     *govmomi.Client
	RestClient *rest.Client
	Finder     *find.Finder
	Datacenter *object.Datacenter
}

func NewDriver(config common.ConnectConfig) (*VCenterDriver, error) {
	ctx := context.Background()

	vcenterUrl, err := url.Parse(fmt.Sprintf("https://%v/sdk", config.VCenterServer))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	vcenterUrl.User = url.UserPassword(config.Username, config.Password)

	client, err := govmomi.NewClient(ctx, vcenterUrl, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create govmomi Client: %w", err)
	}

	restClient := rest.NewClient(client.Client)
	err = restClient.Login(ctx, vcenterUrl.User)
	if err != nil {
		return nil, fmt.Errorf("failed to login to REST API endpoint: %w", err)
	}

	finder := find.NewFinder(client.Client, true)
	datacenter, err := finder.DatacenterOrDefault(ctx, config.Datacenter)
	if err != nil {
		return nil, fmt.Errorf("failed to find datacenter: %w", err)
	}
	finder.SetDatacenter(datacenter)

	return &VCenterDriver{
		Ctx:        ctx,
		Client:     client,
		RestClient: restClient,
		Finder:     finder,
		Datacenter: datacenter,
	}, nil
}
