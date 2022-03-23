//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ShutdownConfig

package common

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type ShutdownConfig struct {
	// Specify a VM guest shutdown command. This command will be executed using
	// the `communicator`. Otherwise, the VMware Tools are used to gracefully shutdown
	// the VM.
	Command string `mapstructure:"shutdown_command"`
	// Amount of time to wait for graceful VM shutdown.
	// Defaults to 5m or five minutes.
	// This will likely need to be modified if the `communicator` is 'none'.
	Timeout time.Duration `mapstructure:"shutdown_timeout"`
	// Packer normally halts the virtual machine after all provisioners have
	// run when no `shutdown_command` is defined. If this is set to `true`, Packer
	// *will not* halt the virtual machine but will assume that you will send the stop
	// signal yourself through a preseed.cfg, a script or the final provisioner.
	// Packer will wait for a default of five minutes until the virtual machine is shutdown.
	// The timeout can be changed using `shutdown_timeout` option.
	DisableShutdown bool `mapstructure:"disable_shutdown"`
	// Wait duration between polling if the VM is shutdown. Defaults to 10 seconds wait between each IsVMDown() call
	PollingInterval time.Duration `mapstructure:"shutdown_polling_interval"`
	// Time to wait before packer checks if the VM is off and send the shutdown command. Defaults to 0
	PauseBeforeShutdown time.Duration `mapstructure:"pause_before_shutdown"`
}

func (c *ShutdownConfig) Prepare(comm communicator.Config) (warnings []string, errs []error) {

	if c.Timeout == 0 {
		c.Timeout = 5 * time.Minute
	}

	if c.PollingInterval > c.Timeout {
		errs = append(errs, fmt.Errorf("The shutdown_interval=%s must be lesser than the shutdown_timeout=%s",
			c.PollingInterval, c.Timeout))
	} else if c.PollingInterval == 0 {
		c.PollingInterval = 10 * time.Second
		if c.PollingInterval > c.Timeout {
			c.PollingInterval = 1 * time.Second
		}
	}

	if comm.Type == "none" && c.Command != "" {
		warnings = append(warnings, "The parameter `shutdown_command` is ignored as it requires a `communicator`.")
	}

	return
}

type StepShutdown struct {
	Config *ShutdownConfig
}

func (s *StepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)

	if s.Config.PauseBeforeShutdown > 0 {
		ui.Say(fmt.Sprintf("Waiting %s before checking if the VM is down", s.Config.PauseBeforeShutdown))
		time.Sleep(s.Config.PauseBeforeShutdown)
	}
	if off, _ := vm.IsPoweredOff(); off {
		if s.Config.DisableShutdown {
			ui.Say("VM is already powered off")
		} else {
			ui.Say("VM is already powered off though disable_shutdown is not true. Moving on")
		}
		return multistep.ActionContinue
	}

	comm, _ := state.Get("communicator").(packersdk.Communicator)
	if s.Config.DisableShutdown {
		msg := fmt.Sprintf("Automatic shutdown from vSphere is disabled. Please shutdown virtual machine within %s.",
			s.Config.Timeout)
		ui.Say(msg)
	} else if comm == nil {
		var msg string
		if s.Config.Command != "" {
			ui.Message("The custom shutdown_command is ignored as the VM `communicator` is not available.")
		}
		msg = fmt.Sprintf("Automatic shutdown via vSphere is disabled. "+
			"Please shutdown virtual machine within %s.", s.Config.Timeout)
		ui.Message(msg)
	} else if s.Config.Command != "" {
		// Communicator is not needed unless shutdown_command is populated

		ui.Say("Executing shutdown command...")
		log.Printf("Shutdown command: %s", s.Config.Command)

		var stdout, stderr bytes.Buffer
		cmd := &packersdk.RemoteCmd{
			Command: s.Config.Command,
			Stdout:  &stdout,
			Stderr:  &stderr,
		}
		err := comm.Start(ctx, cmd)
		if err != nil {
			state.Put("error", fmt.Errorf("Failed to send shutdown command: %s", err))
			return multistep.ActionHalt
		}
	} else {
		ui.Say("Shutting down VM...")

		err := vm.StartShutdown()
		if err != nil {
			state.Put("error", fmt.Errorf("Cannot shut down VM: %v", err))
			return multistep.ActionHalt
		}
	}

	log.Printf("Waiting max %s for shutdown to complete", s.Config.Timeout)
	err := vm.WaitForShutdown(ctx, s.Config.Timeout, time.Duration(s.Config.PollingInterval))
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepShutdown) Cleanup(state multistep.StateBag) {}
