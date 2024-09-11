// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
)

const BuilderId = "packer.post-processor.vsphere"

type Artifact struct {
	files     []string
	datastore string
	vmfolder  string
	vmname    string
}

func NewArtifact(datastore, vmfolder, vmname string, files []string) *Artifact {
	return &Artifact{
		files:     files,
		datastore: datastore,
		vmfolder:  vmfolder,
		vmname:    vmname,
	}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.files
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s::%s::%s", a.datastore, a.vmfolder, a.vmname)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("VM: %s Folder: %s Datastore: %s", a.vmname, a.vmfolder, a.datastore)
}

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
    if len(a.files) == 0 {
        return fmt.Errorf("no files to delete")
    }

    for _, file := range a.files {
        if err := os.Remove(file); err != nil {
            return fmt.Errorf("error deleting file %s: %v", file, err)
        }
        fmt.Printf("Successfully deleted file: %s\n", file)
    }

    return nil
}
