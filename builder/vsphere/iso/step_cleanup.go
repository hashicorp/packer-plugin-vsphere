//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CleanupConfig

package iso

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

type CleanupConfig struct {
	// If set to true, the VM will be destroyed after the builder completes
	Destroy bool `mapstructure:"destroy"`
}

type StepCleanupVM struct {
	Config *CleanupConfig
}

func (c *CleanupConfig) Prepare() []error {
	return nil
}

func (s *StepCleanupVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	if s.Config.Destroy {
		state.Put("destroy_vm", s.Config.Destroy)
	}

	return multistep.ActionContinue
}

func (s *StepCleanupVM) Cleanup(state multistep.StateBag) {

}
