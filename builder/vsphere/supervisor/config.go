//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package supervisor

import (
	packercommon "github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
)

type Config struct {
	packercommon.PackerConfig `mapstructure:",squash"`
	CommunicatorConfig        communicator.Config `mapstructure:",squash"`
	ConnectK8sConfig          `mapstructure:",squash"`
	CreateSourceConfig        `mapstructure:",squash"`
	WatchSourceConfig         `mapstructure:",squash"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         common.BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
			},
		},
	}, raws...)

	if err != nil {
		return nil, err
	}

	errs := new(packersdk.MultiError)
	errs = packersdk.MultiErrorAppend(errs, c.CommunicatorConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.ConnectK8sConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.CreateSourceConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.WatchSourceConfig.Prepare()...)

	if len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}
