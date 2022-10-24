package supervisor

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
)

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("debug", b.config.PackerDebug)
	state.Put("hook", hook)
	state.Put("ui", ui)
	logger := &PackerLogger{UI: ui}
	state.Put("logger", logger)

	var steps []multistep.Step
	steps = append(steps,
		// Generate SSH key pairs for connecting to the source VM.
		&communicator.StepSSHKeyGen{
			CommConf: &b.config.CommunicatorConfig,
		},
		// Connect to the Supervisor cluster where the source VM created.
		&StepConnectSupervisor{
			Config: &b.config.ConnectSupervisorConfig,
		},
		// Create a source VM and other related resources in Supervisor cluster.
		&StepCreateSource{
			Config:             &b.config.CreateSourceConfig,
			CommunicatorConfig: &b.config.CommunicatorConfig,
		},
		// Watch for the source VM to be powered on and accessible.
		&StepWatchSource{
			Config: &b.config.WatchSourceConfig,
		},
		// Connect to the source VM via Packer provided SSH communicator.
		&communicator.StepConnect{
			Config:    &b.config.CommunicatorConfig,
			Host:      common.CommHost(b.config.CommunicatorConfig.Host()),
			SSHConfig: b.config.CommunicatorConfig.SSHConfigFunc(),
		},
		// Run provisioners defined in the Packer template.
		new(commonsteps.StepProvision),
	)

	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	logger.Info("Build 'vsphere-supervisor' finished without publishing the VM image (feature not available yet).")
	return nil, nil
}
