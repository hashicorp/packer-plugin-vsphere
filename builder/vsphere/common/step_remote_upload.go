// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

const DefaultRemoteCachePath = "packer_cache"

type StepRemoteUpload struct {
	Datastore                  string
	Host                       string
	SetHostForDatastoreUploads bool
	RemoteCacheCleanup         bool
	RemoteCacheOverwrite       bool
	RemoteCacheDatastore       string
	RemoteCachePath            string
	UploadedCustomCD           bool
}

func (s *StepRemoteUpload) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	d := state.Get("driver").(driver.Driver)

	if path, ok := state.GetOk("iso_path"); ok {
		// user-supplied boot iso
		fullRemotePath, err := s.uploadFile(path.(string), d, ui, state)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Put("iso_remote_path", fullRemotePath)
	}
	if cdPath, ok := state.GetOk("cd_path"); ok {
		// Packer-created cd_files disk
		fullRemotePath, err := s.uploadFile(cdPath.(string), d, ui, state)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		s.UploadedCustomCD = true
		state.Put("cd_path", fullRemotePath)
	}

	if s.RemoteCacheCleanup {
		state.Put("remote_cache_cleanup", s.RemoteCacheCleanup)
	}

	return multistep.ActionContinue
}

func GetRemoteDirectoryAndPath(path string, ds driver.Datastore, remoteCachePath string) (string, string, string, string) {
	filename := filepath.Base(path)
	remotePath := fmt.Sprintf("%s/%s", remoteCachePath, filename)
	remoteDirectory := fmt.Sprintf("[%s] %s", ds.Name(), remoteCachePath)
	fullRemotePath := fmt.Sprintf("%s/%s", remoteDirectory, filename)

	return filename, remotePath, remoteDirectory, fullRemotePath
}

func (s *StepRemoteUpload) uploadFile(path string, d driver.Driver, ui packersdk.Ui, state multistep.StateBag) (string, error) {

	// Set the remote cache datastore. If not set, use the default datastore for the build.
	remoteCacheDatastore := s.Datastore
	if s.RemoteCacheDatastore != "" {
		remoteCacheDatastore = s.RemoteCacheDatastore
	}

	var ds driver.Datastore
	var err error

	// If a datastore was resolved (from datastore or datastore_cluster), use it.
	if resolvedDs, ok := state.GetOk("datastore"); ok && remoteCacheDatastore == s.Datastore {
		ds = resolvedDs.(driver.Datastore)
	} else {
		// Find the datastore to use for the remote cache.
		ds, err = d.FindDatastore(remoteCacheDatastore, s.Host)
		if err != nil {
			return "", fmt.Errorf("error finding the remote cache datastore: %v", err)
		}
	}

	// Set the remote cache path. If not set, use the default cache path.
	remoteCachePath := s.RemoteCachePath
	if remoteCachePath == "" {
		remoteCachePath = DefaultRemoteCachePath
	}

	filename, remotePath, remoteDirectory, fullRemotePath := GetRemoteDirectoryAndPath(path, ds, remoteCachePath)

	if exists := ds.FileExists(remotePath); exists {
		// If the remote cache overwrite flag is set to true, delete the file and download the
		// ISO file again.
		if s.RemoteCacheOverwrite {
			ui.Sayf("Overwriting %s in remote cache %s...", filename, remoteDirectory)
			// Delete the file from the remote cache datastore.
			if err := ds.Delete(remotePath); err != nil {
				return "", fmt.Errorf("error overwriting file in remote cache: %w", err)
			}
		} else {
			// Skip the download step if the remote cache overwrite flag is not set.
			ui.Sayf("Skipping upload, %s already exists in remote cache...", fullRemotePath)
			return fullRemotePath, nil
		}
	}

	ui.Sayf("Uploading %s to %s...", filename, remoteDirectory)

	// Check if the remote cache directory exists. If not, create it.
	if !ds.DirExists(remotePath) {
		ui.Sayf("Remote cache directory does not exist; creating %s...", remoteDirectory)
		if err := ds.MakeDirectory(remoteDirectory); err != nil {
			return "", err
		}
	}

	// Upload the file to the remote cache datastore.
	if err := ds.UploadFile(path, remotePath, s.Host, s.SetHostForDatastoreUploads); err != nil {
		return "", err
	}
	return fullRemotePath, nil
}

func (s *StepRemoteUpload) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	_, remoteCacheCleanup := state.GetOk("remote_cache_cleanup")

	if !cancelled && !halted && !remoteCacheCleanup {
		return
	}

	if !s.UploadedCustomCD {
		return
	}

	UploadedCDPath, ok := state.GetOk("cd_path")
	if !ok {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	d := state.Get("driver").(*driver.VCenterDriver)
	ui.Sayf("Removing %s...", UploadedCDPath)

	var ds driver.Datastore
	var err error

	// If a datastore was resolved (from datastore or datastore_cluster), use it.
	if resolvedDs, ok := state.GetOk("datastore"); ok {
		ds = resolvedDs.(driver.Datastore)
	} else {
		ds, err = d.FindDatastore(s.Datastore, s.Host)
		if err != nil {
			ui.Sayf("Unable to find the remote cache datastore. Please remove the item manually: %s", err)
			return
		}
	}

	err = ds.Delete(UploadedCDPath.(string))
	if err != nil {
		ui.Sayf("Unable to remove item from the remote cache. Please remove the item manually: %s", err)
		return
	}
}
