// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ImportImageConfig

package supervisor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	imgregv1 "github.com/vmware-tanzu/image-registry-operator-api/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ImportTargetKind       = "ContentLibrary"
	ImportTargetAPIVersion = "imageregistry.vmware.com/v1alpha1"

	DefaultWatchImageImportTimeoutSec = 600

	ImportRequestDefaultNamePrefix    = "packer-vsphere-supervisor-import-req-"
	StateKeyImageImportRequestCreated = "item_import_req_created"
	StateKeyCleanImportedImage        = "clean_imported_image"
	StateKeyImportedImageName         = "imported_image_name"

	importOVFFeatureNotEnabledMsg = "WCP_VMImageService_ImportOVF feature is not enabled"
)

var IsWatchingImageImport bool

type ImportImageConfig struct {
	// The remote URL where the to-be-imported image is hosted.
	ImportSourceURL string `mapstructure:"import_source_url"`
	// The SSL certificate of the remote HTTP server that hosts the to-be-imported image.
	ImportSourceSSLCertificate string `mapstructure:"import_source_ssl_certificate"`

	// Name of a writable & import-allowed ContentLibrary resource in the namespace where the image will be imported.
	ImportTargetLocationName string `mapstructure:"import_target_location_name"`
	// The type of the imported image.
	// Defaults to `ovf`. Available options include `ovf`.
	ImportTargetImageType string `mapstructure:"import_target_image_type"`

	// The name of the image import request.
	// Defaults to `packer-vsphere-supervisor-import-req-<random-suffix>`.
	ImportRequestName string `mapstructure:"import_request_name"`
	// The timeout in seconds to wait for the image to be imported.
	// Defaults to `600`.
	WatchImportTimeoutSec int `mapstructure:"watch_import_timeout_sec"`

	// Whether to clean the image imported in this step. If it is set to true, the imported image will be deleted after
	// source VM is created and becomes ready.
	// Defaults to false.
	CleanImportedImage bool `mapstructure:"clean_imported_image"`
}

func (c *ImportImageConfig) Prepare() []error {
	if c.ImportSourceURL == "" || c.ImportTargetLocationName == "" {
		return nil
	}

	var errs []error

	switch c.ImportTargetImageType {
	case "":
		c.ImportTargetImageType = "ovf"
	case "ovf":
		// If it's already "ovf", do nothing.
	default:
		errs = append(errs, fmt.Errorf("unsupported ImportTargetImageType: %s", c.ImportTargetImageType))
	}

	if c.WatchImportTimeoutSec == 0 {
		c.WatchImportTimeoutSec = DefaultWatchImageImportTimeoutSec
	}
	if c.ImportRequestName == "" {
		c.ImportRequestName = ImportRequestDefaultNamePrefix + uuid.NewString()[:5]
	}

	return errs
}

type StepImportImage struct {
	ImportImageConfig  *ImportImageConfig
	CreateSourceConfig *CreateSourceConfig

	SourceURL, SourceSSLCertificate, TargetLocationName, TargetItemName, Namespace string
	TargetItemType                                                                 imgregv1.ContentLibraryItemType
	KeepInputArtifact                                                              bool
	KubeWatchClient                                                                client.WithWatch
}

func (s *StepImportImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	logger := state.Get("logger").(*PackerLogger)

	var err error
	defer func() {
		if err != nil {
			state.Put("error", err)
		}
	}()

	// Skip importing if the import source URL or target location is not specified.
	if s.ImportImageConfig.ImportSourceURL == "" || s.ImportImageConfig.ImportTargetLocationName == "" {
		logger.Info("Skipping image import step. Required configurations for the import are not set.")
		return multistep.ActionContinue
	}

	if err = s.initStep(state, logger); err != nil {
		logger.Error("failed to initialize image import: %s", err.Error())
		return multistep.ActionHalt
	}

	if err = s.validate(ctx, logger); err != nil {
		logger.Error("failed to validate import image configs: %s", err.Error())
		return multistep.ActionHalt
	}

	logger.Info("Importing the source image from %s to %s.", s.SourceURL, s.TargetLocationName)

	if err = s.createImageImportRequest(ctx, logger); err != nil {
		return multistep.ActionHalt
	}
	state.Put(StateKeyImageImportRequestCreated, true)

	if err = s.watchItemImport(ctx, state, logger); err != nil {
		return multistep.ActionHalt
	}

	logger.Info("Finished importing the image from %s to %s.", s.SourceURL, s.TargetLocationName)

	return multistep.ActionContinue
}

func (s *StepImportImage) validate(ctx context.Context, logger *PackerLogger) error {
	logger.Info("Validating image import request...")

	var err error
	if err = s.isImportFeatureEnabled(ctx, logger); err != nil {
		return err
	}

	if err = s.isImportSourceValid(); err != nil {
		return err
	}

	if err = s.isImportTargetValid(ctx, logger); err != nil {
		return err
	}

	logger.Info("Image import request source and target are valid.")
	return nil
}

func (s *StepImportImage) Cleanup(state multistep.StateBag) {
	if state.Get(StateKeyImageImportRequestCreated) == false {
		// Either the image import step was skipped or the object was not created successfully.
		// Skip deleting the ContentLibraryItemImportRequest object.
		return
	}

	logger := state.Get("logger").(*PackerLogger)
	if s.KeepInputArtifact {
		logger.Info("Skipping clean up of the ContentLibraryItemImportRequest object as specified in config.")
		return
	}

	logger.Info(fmt.Sprintf("Deleting the ContentLibraryItemImportRequest object %s in namespace %s.",
		s.ImportImageConfig.ImportRequestName, s.Namespace))
	ctx := context.Background()
	itemImportReqObj := &imgregv1.ContentLibraryItemImportRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ImportImageConfig.ImportRequestName,
			Namespace: s.Namespace,
		},
	}
	if err := s.KubeWatchClient.Delete(ctx, itemImportReqObj); err != nil {
		logger.Error("error deleting the ContentLibraryItemImportRequest object : %s", err)
	} else {
		logger.Info("Successfully deleted the ContentLibraryItemImportRequest object %s in namespace %s.",
			s.ImportImageConfig.ImportRequestName, s.Namespace)
	}
}

func (s *StepImportImage) initStep(state multistep.StateBag, logger *PackerLogger) error {
	if err := CheckRequiredStates(state,
		StateKeySupervisorNamespace,
		StateKeyKubeClient,
	); err != nil {
		logger.Error("error checking required states: %s", err)
		return err
	}

	var ok bool
	if s.Namespace, ok = state.Get(StateKeySupervisorNamespace).(string); !ok {
		return fmt.Errorf("failed to cast %s to type string", StateKeySupervisorNamespace)
	}
	if s.KubeWatchClient, ok = state.Get(StateKeyKubeClient).(client.WithWatch); !ok {
		return fmt.Errorf("failed to cast %s to type client.WithWatch", StateKeyKubeClient)
	}

	s.SourceURL = s.ImportImageConfig.ImportSourceURL
	s.SourceSSLCertificate = s.ImportImageConfig.ImportSourceSSLCertificate
	s.TargetLocationName = s.ImportImageConfig.ImportTargetLocationName
	if s.ImportImageConfig.ImportTargetImageType != "" {
		s.TargetItemType = imgregv1.ContentLibraryItemType(strings.ToUpper(s.ImportImageConfig.ImportTargetImageType))
	}

	// Use the ImageName from create source config as the target item name.
	s.TargetItemName = s.CreateSourceConfig.ImageName
	s.KeepInputArtifact = s.CreateSourceConfig.KeepInputArtifact

	state.Put(StateKeyCleanImportedImage, s.ImportImageConfig.CleanImportedImage)
	state.Put(StateKeyImportedImageName, s.TargetItemName)

	return nil
}

func (s *StepImportImage) isImportFeatureEnabled(ctx context.Context, logger *PackerLogger) error {
	importReq := &imgregv1.ContentLibraryItemImportRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    s.Namespace,
			GenerateName: "import-",
		},
	}

	// Use dry run mode to send an image import creation request to API-Server without applying the resource.
	err := client.NewDryRunClient(s.KubeWatchClient).Create(ctx, importReq)
	if err != nil && strings.Contains(err.Error(), importOVFFeatureNotEnabledMsg) {
		logger.Error("image import feature is not enabled")
		return err
	}

	return nil
}

func (s *StepImportImage) isImportSourceValid() error {
	if strings.HasPrefix(s.ImportImageConfig.ImportSourceURL, "https://") && s.ImportImageConfig.ImportSourceSSLCertificate == "" {
		return fmt.Errorf("import request source url certificate is empty")
	}

	return nil
}

func (s *StepImportImage) isImportTargetValid(ctx context.Context, logger *PackerLogger) error {
	cl := &imgregv1.ContentLibrary{}
	objKey := client.ObjectKey{Name: s.TargetLocationName, Namespace: s.Namespace}
	if err := s.KubeWatchClient.Get(ctx, objKey, cl); err != nil {
		logger.Error(fmt.Sprintf("failed to return the content library by name %s in namespace %s", s.TargetLocationName, s.Namespace))
		return err
	}

	if !cl.Spec.Writable || !cl.Spec.AllowImport {
		return fmt.Errorf("import target content library %q is not writable or does not allow import", s.TargetLocationName)
	}

	// Only supports OVF type for now, this check needs to be updated when supporting other types.
	if s.TargetItemType != imgregv1.ContentLibraryItemTypeOvf {
		return fmt.Errorf("image type %s is not supported", s.ImportImageConfig.ImportTargetImageType)
	}

	return nil
}

func (s *StepImportImage) createImageImportRequest(ctx context.Context, logger *PackerLogger) error {
	logger.Info("Creating ContentLibraryItemImportRequest object %s in namespace %s.", s.ImportImageConfig.ImportRequestName, s.Namespace)

	imageImportReq := &imgregv1.ContentLibraryItemImportRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ImportImageConfig.ImportRequestName,
			Namespace: s.Namespace,
		},
		Spec: imgregv1.ContentLibraryItemImportRequestSpec{
			Source: imgregv1.ContentLibraryItemImportRequestSource{
				URL:            s.SourceURL,
				SSLCertificate: s.SourceSSLCertificate,
			},
			Target: imgregv1.ContentLibraryItemImportRequestTarget{
				Library: imgregv1.LocalObjectRef{
					Kind:       ImportTargetKind,
					APIVersion: ImportTargetAPIVersion,
					Name:       s.TargetLocationName,
				},
				Item: imgregv1.ContentLibraryItemImportRequestTargetItem{
					Name: s.TargetItemName,
					Type: s.TargetItemType,
				},
			},
		},
	}

	if err := s.KubeWatchClient.Create(ctx, imageImportReq); err != nil {
		logger.Error("error creating the ContentLibraryItemImportRequest object %s.", imageImportReq.Name)
		return err
	}

	logger.Info("Successfully created the ContentLibraryItemImportRequest object %s.", imageImportReq.Name)
	return nil
}

func (s *StepImportImage) watchItemImport(ctx context.Context, state multistep.StateBag, logger *PackerLogger) error {
	itemImportReqWatch, err := s.KubeWatchClient.Watch(ctx, &imgregv1.ContentLibraryItemImportRequestList{}, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", s.ImportImageConfig.ImportRequestName),
		Namespace:     s.Namespace,
	})

	if err != nil {
		logger.Error("error watching the ContentLibraryItemImportRequest object in Supervisor cluster")
		return err
	}

	timedCtx, cancel := context.WithTimeout(ctx, time.Duration(s.ImportImageConfig.WatchImportTimeoutSec)*time.Second)

	defer func() {
		itemImportReqWatch.Stop()
		cancel()

		Mu.Lock()
		IsWatchingImageImport = false
		Mu.Unlock()
	}()

	Mu.Lock()
	IsWatchingImageImport = true
	Mu.Unlock()

	for {
		select {
		case event := <-itemImportReqWatch.ResultChan():
			if event.Object == nil {
				return fmt.Errorf("watch ContentLibraryItemImportRequest event object is nil")
			}

			itemImportReqObj, ok := event.Object.(*imgregv1.ContentLibraryItemImportRequest)
			if !ok {
				return fmt.Errorf("failed to convert the watch ContentLibraryItemImportRequest event object")
			}

			importSuccess := false
			for _, cond := range itemImportReqObj.Status.Conditions {
				if cond.Type == imgregv1.ContentLibraryItemImportRequestComplete {
					importSuccess = cond.Status == corev1.ConditionTrue
				}
			}
			if importSuccess {
				// Set item ref name if the import is successful.
				state.Put(StateKeyImportedImageName, itemImportReqObj.Status.ItemRef.Name)

				logger.Info("Successfully imported the image as a content library item %q.", itemImportReqObj.Status.ItemRef)
				return nil
			} else {
				logger.Info("Waiting for the image import request to complete...")
			}

		case <-timedCtx.Done():
			return fmt.Errorf("timed out watching for ContentLibraryItemImportRequest object to complete")
		}
	}
}
