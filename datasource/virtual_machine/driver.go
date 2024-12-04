package virtual_machine

import (
	"context"
	"fmt"

	"net/url"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/rest"
)

type VCenterDriver struct {
	ctx        context.Context
	client     *govmomi.Client
	restClient *rest.Client
	finder     *find.Finder
	datacenter *object.Datacenter
}

func newDriver(config Config) (*VCenterDriver, error) {
	ctx := context.Background()

	vcenterUrl, err := url.Parse(fmt.Sprintf("https://%v/sdk", config.VCenterServer))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse URL")
	}
	vcenterUrl.User = url.UserPassword(config.Username, config.Password)

	client, err := govmomi.NewClient(ctx, vcenterUrl, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create govmomi client")
	}

	var restClient *rest.Client
	if config.VmTags != nil {
		// REST client is only needed when the plugin has to retrieve tags from VMs.
		// Skip initialization if not needed (there is additional risk of fail on old vCenter versions).
		restClient = rest.NewClient(client.Client)
		err = restClient.Login(ctx, vcenterUrl.User)
		if err != nil {
			return nil, errors.Wrap(err, "failed to login to REST API endpoint")
		}
	}

	finder := find.NewFinder(client.Client, true)
	datacenter, err := finder.DatacenterOrDefault(ctx, config.Datacenter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find datacenter")
	}
	finder.SetDatacenter(datacenter)

	return &VCenterDriver{
		ctx:        ctx,
		client:     client,
		restClient: restClient,
		finder:     finder,
		datacenter: datacenter,
	}, nil
}
