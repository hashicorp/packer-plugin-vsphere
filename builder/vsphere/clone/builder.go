// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clone

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
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

	var steps []multistep.Step

	steps = append(steps,
		&common.StepConnect{
			Config: &b.config.ConnectConfig,
		},
		&commonsteps.StepCreateCD{
			Files:   b.config.CDFiles,
			Content: b.config.CDContent,
			Label:   b.config.CDLabel,
		},
		&common.StepRemoteUpload{
			Datastore:                  b.config.Datastore,
			Host:                       b.config.Host,
			SetHostForDatastoreUploads: b.config.SetHostForDatastoreUploads,
		},
		&StepCloneVM{
			Config:   &b.config.CloneConfig,
			Location: &b.config.LocationConfig,
			Force:    b.config.PackerForce,
		},
		&common.StepConfigureHardware{
			Config: &b.config.HardwareConfig,
		},
		&common.StepAddFlag{
			FlagConfig: b.config.FlagConfig,
		},
		&common.StepAddCDRom{
			Config: &b.config.CDRomConfig,
		},
		&common.StepConfigParams{
			Config: &b.config.ConfigParamsConfig,
		},
	)

	if b.config.CustomizeConfig != nil {
		steps = append(steps, &StepCustomize{
			Config: b.config.CustomizeConfig,
		})
	}

	if b.config.Comm.Type != "none" {
		steps = append(steps,
			&commonsteps.StepCreateFloppy{
				Files:       b.config.FloppyFiles,
				Directories: b.config.FloppyDirectories,
				Content:     b.config.FloppyContent,
				Label:       b.config.FloppyLabel,
			},
			&common.StepAddFloppy{
				Config:                     &b.config.FloppyConfig,
				Datastore:                  b.config.Datastore,
				Host:                       b.config.Host,
				SetHostForDatastoreUploads: b.config.SetHostForDatastoreUploads,
			},
		)

		// Set the address for the HTTP server based on the configuration
		// provided by the user.
		if addrs := b.config.HTTPAddress; addrs != "" && addrs != common.DefaultHttpBindAddress {
			// Use the specified HTTPAddress, if valid.
			err := common.ValidateHTTPAddress(addrs)
			if err != nil {
				ui.Errorf("error validating IP address for HTTP server: %s", err)
				return nil, err
			}
			state.Put("http_bind_address", addrs)
		} else if intf := b.config.HTTPInterface; intf != "" {
			// Use the specified HTTPInterface, if valid.
			state.Put("http_interface", intf)
		} else {
			// Use IP discovery if neither is specified.
			steps = append(steps, &common.StepHTTPIPDiscover{
				HTTPIP:  b.config.HTTPIP,
				Network: b.config.GetIPNet(),
			})
		}

		steps = append(steps,
			commonsteps.HTTPServerFromHTTPConfig(&b.config.HTTPConfig),
			&common.StepSshKeyPair{
				Debug:        b.config.PackerDebug,
				DebugKeyPath: fmt.Sprintf("%s.pem", b.config.PackerBuildName),
				Comm:         &b.config.Comm,
			},
			&common.StepRun{
				Config:   &b.config.RunConfig,
				SetOrder: false,
			},
			&common.StepBootCommand{
				Config: &b.config.BootConfig,
				Ctx:    b.config.ctx,
				VMName: b.config.VMName,
			},
			&common.StepWaitForIp{
				Config: &b.config.WaitIpConfig,
			},
			&communicator.StepConnect{
				Config:    &b.config.Comm,
				Host:      common.CommHost(b.config.Comm.Host()),
				SSHConfig: b.config.Comm.SSHConfigFunc(),
			},
			&commonsteps.StepProvision{},
			&common.StepShutdown{
				Config: &b.config.ShutdownConfig,
			},
			&common.StepRemoveFloppy{
				Datastore: b.config.Datastore,
				Host:      b.config.Host,
			},
		)
	}

	steps = append(steps,
		&common.StepRemoveCDRom{
			Config: &b.config.RemoveCDRomConfig,
		},
		&common.StepReattachCDRom{
			Config:      &b.config.ReattachCDRomConfig,
			CDRomConfig: &b.config.CDRomConfig,
		},
		&common.StepCreateSnapshot{
			CreateSnapshot: b.config.CreateSnapshot,
			SnapshotName:   b.config.SnapshotName,
		},
		&common.StepRemoveNetworkAdapter{
			Config: &b.config.RemoveNetworkAdapterConfig,
		},
		&common.StepConvertToTemplate{
			ConvertToTemplate: b.config.ConvertToTemplate,
		},
	)

	if b.config.ContentLibraryDestinationConfig != nil {
		steps = append(steps, &common.StepImportToContentLibrary{
			ContentLibConfig: b.config.ContentLibraryDestinationConfig,
		})
	}

	if b.config.Export != nil {
		steps = append(steps, &common.StepExport{
			Name:       b.config.Export.Name,
			Force:      b.config.Export.Force,
			ImageFiles: b.config.Export.ImageFiles,
			Manifest:   b.config.Export.Manifest,
			OutputDir:  b.config.Export.OutputDir.OutputDir,
			Options:    b.config.Export.Options,
			Format:     b.config.Export.Format,
		})
	}

	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("vm"); !ok {
		return nil, nil
	}
	vm := state.Get("vm").(*driver.VirtualMachineDriver)
	artifact := &common.Artifact{
		Name:                 b.config.VMName,
		Datacenter:           vm.Datacenter(),
		Location:             b.config.LocationConfig,
		ContentLibraryConfig: b.config.ContentLibraryDestinationConfig,
		VM:                   vm,
		StateData: map[string]interface{}{
			"generated_data":  state.Get("generated_data"),
			"metadata":        state.Get("metadata"),
			"source_template": b.config.Template,
		},
	}
	if b.config.Export != nil {
		artifact.Outconfig = &b.config.Export.OutputDir
	}
	return artifact, nil
}
