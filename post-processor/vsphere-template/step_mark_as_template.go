// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere_template

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-vsphere/post-processor/vsphere"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type StepMarkAsTemplate struct {
	VMName       string
	TemplateName string
	RemoteFolder string
	ReregisterVM config.Trilean
}

func NewStepMarkAsTemplate(artifact packersdk.Artifact, p *PostProcessor) *StepMarkAsTemplate {
	// Set the default folder.
	remoteFolder := "Discovered virtual machine"

	// If the post-processor configuration's folder is defined, use it as the `remoteFolder`.
	if p.config.Folder != "" {
		remoteFolder = p.config.Folder
	}

	vmname := artifact.Id()

	if artifact.BuilderId() == vsphere.BuilderId {
		id := strings.Split(artifact.Id(), "::")
		remoteFolder = id[1]
		vmname = id[2]
	}

	return &StepMarkAsTemplate{
		VMName:       vmname,
		TemplateName: p.config.TemplateName,
		RemoteFolder: remoteFolder,
		ReregisterVM: p.config.ReregisterVM,
	}
}

func (s *StepMarkAsTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	cli := state.Get("client").(*govmomi.Client)
	folder := state.Get("folder").(*object.Folder)
	dcPath := state.Get("dcPath").(string)

	vm, err := findRuntimeVM(cli, dcPath, s.VMName, s.RemoteFolder)
	if err != nil {
		state.Put("error", err)
		ui.Errorf("%s", err)
		return multistep.ActionHalt
	}

	// Use the MarkAsTemplate method unless the `reregister_vm` is set to `true`.
	if s.ReregisterVM.False() {
		ui.Say("Marking as a template...")

		if err := vm.MarkAsTemplate(context.Background()); err != nil {
			state.Put("error", err)
			ui.Errorf("vm.MarkAsTemplate: %s", err)
			return multistep.ActionHalt
		}
		return multistep.ActionContinue
	}

	dsPath, err := datastorePath(vm)
	if err != nil {
		state.Put("error", err)
		ui.Errorf("datastorePath: %s", err)
		return multistep.ActionHalt
	}

	host, err := vm.HostSystem(context.Background())
	if err != nil {
		state.Put("error", err)
		ui.Errorf("vm.HostSystem: %s", err)
		return multistep.ActionHalt
	}

	if err := vm.Unregister(context.Background()); err != nil {
		state.Put("error", err)
		ui.Errorf("vm.Unregister: %s", err)
		return multistep.ActionHalt
	}

	if err := unregisterPreviousVM(cli, folder, s.VMName); err != nil {
		state.Put("error", err)
		ui.Errorf("unregisterPreviousVM: %s", err)
		return multistep.ActionHalt
	}

	artifactName := s.VMName
	if s.TemplateName != "" {
		artifactName = s.TemplateName
	}

	ui.Say("Registering virtual machine as a template: " + artifactName)

	task, err := folder.RegisterVM(context.Background(), dsPath.String(), artifactName, true, nil, host)
	if err != nil {
		state.Put("error", err)
		ui.Errorf("RegisterVM: %s", err)
		return multistep.ActionHalt
	}

	if err = task.Wait(context.Background()); err != nil {
		state.Put("error", err)
		ui.Errorf("task.Wait: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func datastorePath(vm *object.VirtualMachine) (*object.DatastorePath, error) {
	devices, err := vm.Device(context.Background())
	if err != nil {
		return nil, err
	}

	disk := ""
	for _, device := range devices {
		if d, ok := device.(*types.VirtualDisk); ok {
			if b, ok := d.Backing.(types.BaseVirtualDeviceFileBackingInfo); ok {
				disk = b.GetVirtualDeviceFileBackingInfo().FileName
			}
			break
		}
	}

	if disk == "" {
		return nil, fmt.Errorf("error finding disk in '%v'", vm.Name())
	}

	re := regexp.MustCompile(`\[(.*?)\]`)
	datastore := re.FindStringSubmatch(disk)[1]
	vmxPath := path.Join("/", path.Dir(strings.Split(disk, " ")[1]), vm.Name()+".vmx")

	return &object.DatastorePath{
		Datastore: datastore,
		Path:      vmxPath,
	}, nil
}

func findRuntimeVM(cli *govmomi.Client, dcPath, name, remoteFolder string) (*object.VirtualMachine, error) {
	si := object.NewSearchIndex(cli.Client)
	fullPath := path.Join(dcPath, "vm", remoteFolder, name)

	ref, err := si.FindByInventoryPath(context.Background(), fullPath)
	if err != nil {
		return nil, err
	}

	if ref == nil {
		return nil, fmt.Errorf("error finding virtual machine at path %s", fullPath)
	}

	vm := ref.(*object.VirtualMachine)
	if vm.InventoryPath == "" {
		vm.SetInventoryPath(fullPath)
	}
	return vm, nil
}

func unregisterPreviousVM(cli *govmomi.Client, folder *object.Folder, name string) error {
	si := object.NewSearchIndex(cli.Client)
	fullPath := path.Join(folder.InventoryPath, name)

	ref, err := si.FindByInventoryPath(context.Background(), fullPath)
	if err != nil {
		return err
	}

	if ref != nil {
		if vm, ok := ref.(*object.VirtualMachine); ok {
			return vm.Unregister(context.Background())
		} else {
			return fmt.Errorf("object name '%v' already exists", name)
		}
	}
	return nil
}

func (s *StepMarkAsTemplate) Cleanup(multistep.StateBag) {}
