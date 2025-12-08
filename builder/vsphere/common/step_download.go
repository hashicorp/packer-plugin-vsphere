// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

// DownloadStep is an interface representing a step in the download process, providing execution, cleanup, and cache
// handling.
type DownloadStep interface {
	Run(context.Context, multistep.StateBag) multistep.StepAction
	Cleanup(multistep.StateBag)
	UseSourceToFindCacheTarget(source string) (*url.URL, string, error)
}

// Before downloading and uploading an ISO file to the remote cache, check if the ISO file
// already exists on the datastore.
//
// If it exists, we check if the overwrite flag is set to true. If it is, we delete the file
// and download it again. If it is not set or false, the download is skipped.
//
// This wrapping-step still uses the common StepDownload, but only if the ISO file does not
// already exist on the datastore.
type StepDownload struct {
	DownloadStep DownloadStep
	// These keys are vSphere-specific and are used to check the remote datastore.
	Url                  []string
	ResultKey            string
	Datastore            string
	Host                 string
	LocalCacheOverwrite  bool
	RemoteCacheOverwrite bool
	RemoteCacheDatastore string
	RemoteCachePath      string
}

func (s *StepDownload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	d := state.Get("driver").(driver.Driver)
	ui := state.Get("ui").(packersdk.Ui)

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
			state.Put("error", fmt.Errorf("error finding the datastore: %v", err))
			return multistep.ActionHalt
		}
	}

	// Set the remote cache path. If not set, use the default cache path.
	remoteCachePath := s.RemoteCachePath
	if remoteCachePath == "" {
		remoteCachePath = DefaultRemoteCachePath
	}

	// Loop over the URLs to see if any are already present.
	// If they are, store in state and continue.
	for _, source := range s.Url {
		_, targetPath, err := s.DownloadStep.UseSourceToFindCacheTarget(source)
		if err != nil {
			state.Put("error", fmt.Errorf("error returning target path: %s", err))
			return multistep.ActionHalt
		}

		filename := filepath.Base(targetPath)

		// Check if the ISO file is already present in the local cache.
		if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
			// If the local cache overwrite flag is set to true, delete the file and download the
			// ISO file again.
			if s.LocalCacheOverwrite {
				ui.Sayf("Overwriting %s in local cache...", filename)
				// Delete the file from the local cache.
				if err := os.Remove(targetPath); err != nil {
					state.Put("error", fmt.Errorf("error overwriting file in local cache: %w", err))
					return multistep.ActionHalt
				}
			}
		}

		_, remotePath, remoteDirectory, _ := GetRemoteDirectoryAndPath(filename, ds, remoteCachePath)

		if exists := ds.FileExists(remotePath); exists {
			// If the remote cache overwrite flag is set to true, delete the file and download the
			// ISO file again.
			if s.RemoteCacheOverwrite {
				if s.LocalCacheOverwrite {
					log.Println("The local cache overwrite flag is set to true. Files will also be overwritten in the remote cache datastore to ensure consistency.")
				}
				ui.Sayf("Overwriting %s in remote cache %s...", filename, remoteDirectory)
				// Delete the file from the remote cache datastore.
				if err := ds.Delete(remotePath); err != nil {
					state.Put("error", fmt.Errorf("error overwriting file in remote cache: %w", err))
					return multistep.ActionHalt
				}
			} else {
				// Skip the download step if the file exists in the local cache.
				ui.Sayf("Skipping download, %s already exists in the local cache...", filename)
				state.Put(s.ResultKey, targetPath)
				state.Put("SourceImageURL", source)
				return multistep.ActionContinue
			}
		}
	}

	// The ISO file is not present on the remote cache datastore. Continue with the download step.
	return s.DownloadStep.Run(ctx, state)
}

func (s *StepDownload) Cleanup(state multistep.StateBag) {
}
