// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ValidatePublishConfig

package supervisor

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	imgregv1a1 "github.com/vmware-tanzu/image-registry-operator-api/api/v1alpha1"
	vmopv1alpha1 "github.com/vmware-tanzu/vm-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	StateKeyPublishLocationName = "publish_location_name"

	vmPubFeatureNotEnabledMsg = "WCP_VM_Image_Registry feature not enabled"
)

type ValidatePublishConfig struct {
	// Name of a writable ContentLibrary resource associated with namespace where the source VM will be published.
	PublishLocationName string `mapstructure:"publish_location_name"`
}

func (c *ValidatePublishConfig) Prepare() []error {
	return nil
}

type StepValidatePublish struct {
	Config *ValidatePublishConfig

	Namespace  string
	KubeClient client.Client
}

func (s *StepValidatePublish) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	logger := state.Get("logger").(*PackerLogger)
	logger.Info("Validating VM publish location...")

	var err error
	defer func() {
		if err != nil {
			state.Put("error", err)
		}
	}()

	state.Put(StateKeyPublishLocationName, s.Config.PublishLocationName)

	if s.Config.PublishLocationName == "" {
		logger.Info("VM publish step will be skipped as the `publish_location_name` config is not set")
		return multistep.ActionContinue
	}

	if err = s.initStep(state); err != nil {
		return multistep.ActionHalt
	}

	if err = s.isPublishFeatureEnabled(ctx, logger); err != nil {
		return multistep.ActionHalt
	}

	if err = s.isPublishLocationValid(ctx, logger); err != nil {
		return multistep.ActionHalt
	}

	logger.Info("VM publish location is valid")

	return multistep.ActionContinue
}

func (s *StepValidatePublish) Cleanup(state multistep.StateBag) {}

func (s *StepValidatePublish) initStep(state multistep.StateBag) error {
	if err := CheckRequiredStates(state,
		StateKeySupervisorNamespace,
		StateKeyKubeClient,
	); err != nil {
		return err
	}

	var (
		ok         bool
		namespace  string
		kubeClient client.Client
	)

	if namespace, ok = state.Get(StateKeySupervisorNamespace).(string); !ok {
		return fmt.Errorf("failed to cast %q from state bag as type string", StateKeySupervisorNamespace)
	}
	if kubeClient, ok = state.Get(StateKeyKubeClient).(client.Client); !ok {
		return fmt.Errorf("failed to cast %q from state bag as type 'client.Client'", StateKeyKubeClient)
	}

	s.Namespace = namespace
	s.KubeClient = kubeClient

	return nil
}

func (s *StepValidatePublish) isPublishFeatureEnabled(ctx context.Context, logger *PackerLogger) error {
	vmPublishReq := &vmopv1alpha1.VirtualMachinePublishRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    s.Namespace,
			GenerateName: "vmpub-",
		},
	}

	// Use dry run mode to send a VM publish creation request to API-Server without applying the resource.
	err := client.NewDryRunClient(s.KubeClient).Create(ctx, vmPublishReq)
	if err != nil && strings.Contains(err.Error(), vmPubFeatureNotEnabledMsg) {
		logger.Error("publish feature is not enabled in the version of vsphere supervisor cluster")
		return err
	}

	return nil
}

func (s *StepValidatePublish) isPublishLocationValid(ctx context.Context, logger *PackerLogger) error {
	cl := &imgregv1a1.ContentLibrary{}
	objKey := client.ObjectKey{Name: s.Config.PublishLocationName, Namespace: s.Namespace}
	if err := s.KubeClient.Get(ctx, objKey, cl); err != nil {
		return err
	}

	if !cl.Spec.Writable {
		return fmt.Errorf("publish location %q is not writable", s.Config.PublishLocationName)
	}

	return nil
}
