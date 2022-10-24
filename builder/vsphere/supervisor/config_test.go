package supervisor_test

import (
	"testing"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func getCompleteConfig(t *testing.T) map[string]interface{} {
	// Use a valid kubeconfig file as we check the content in config.Prepare() function.
	validPath := getTestKubeconfigFile(t, "").Name()

	return map[string]interface{}{
		"image_name":               "test-image",
		"class_name":               "test-class",
		"storage_class":            "test-storage",
		"supervisor_namespace":     "test-namespace",
		"source_name":              "test-source",
		"network_type":             "test-networkType",
		"network_name":             "test-networkName",
		"watch_source_timeout_sec": 60,
		"keep_input_artifact":      true,
		"kubeconfig_path":          validPath,
	}
}

func getMinimalConfig() map[string]interface{} {
	return map[string]interface{}{
		"image_name":    "test-image",
		"class_name":    "test-class",
		"storage_class": "test-storage",
	}
}

func TestConfig_Minimal(t *testing.T) {
	// Using a minimal config requires that a valid kubeconfig is loaded automatially
	validPath := getTestKubeconfigFile(t, "").Name()
	t.Setenv("KUBECONFIG", validPath)

	c := new(supervisor.Config)
	minConfigs := getMinimalConfig()
	// The 'supervisor_namespace' is an optional config but it
	// requires a valid kubeconfig file to get the default value.
	minConfigs["supervisor_namespace"] = "test-ns"
	warns, err := c.Prepare(minConfigs)
	if len(warns) != 0 {
		t.Errorf("expected no warnings, got: %#v", warns)
	}
	if err != nil {
		t.Errorf("expected no errors, got: %s", err)
	}
}

func TestConfig_Required(t *testing.T) {
	c := new(supervisor.Config)
	minConfigs := getMinimalConfig()
	for key, val := range minConfigs {
		minConfigs[key] = ""
		_, err := c.Prepare(minConfigs)
		if err == nil {
			t.Errorf("expected an error for the required config: %s", key)
		}
		minConfigs[key] = val
	}
}

func TestConfig_Complete(t *testing.T) {
	c := new(supervisor.Config)
	allConfigs := getCompleteConfig(t)
	warns, err := c.Prepare(allConfigs)
	if len(warns) != 0 {
		t.Errorf("expected no warnings, got: %#v", warns)
	}
	if err != nil {
		t.Errorf("expected no errors, got: %s", err)
	}
}

func TestConfig_Values(t *testing.T) {
	c := new(supervisor.Config)
	providedConfigs := getCompleteConfig(t)
	warns, err := c.Prepare(providedConfigs)
	if len(warns) != 0 {
		t.Fatalf("expected no warnings, got: %#v", warns)
	}
	if err != nil {
		t.Fatalf("expected no errors, got: %s", err)
	}

	if c.ImageName != providedConfigs["image_name"] {
		t.Errorf("expected image_name to be: %s, got: %s",
			providedConfigs["image_name"], c.ImageName)
	}
	if c.ClassName != providedConfigs["class_name"] {
		t.Errorf("expected class_name to be: %s, got: %s",
			providedConfigs["class_name"], c.ClassName)
	}
	if c.StorageClass != providedConfigs["storage_class"] {
		t.Errorf("expected storage_class to be: %s, got: %s",
			providedConfigs["storage_class"], c.StorageClass)
	}
	if c.KubeconfigPath != providedConfigs["kubeconfig_path"] {
		t.Errorf("expected kubeconfig_path to be: %s, got: %s",
			providedConfigs["kubeconfig_path"], c.KubeconfigPath)
	}
	if c.SupervisorNamespace != providedConfigs["supervisor_namespace"] {
		t.Errorf("expected supervisor_namespace to be: %s, got: %s",
			providedConfigs["supervisor_namespace"], c.SupervisorNamespace)
	}
	if c.SourceName != providedConfigs["source_name"] {
		t.Errorf("expected source_name to be: %s, got: %s",
			providedConfigs["source_name"], c.SourceName)
	}
	if c.NetworkType != providedConfigs["network_type"] {
		t.Errorf("expected network_type to be: %s, got: %s",
			providedConfigs["network_type"], c.NetworkType)
	}
	if c.NetworkName != providedConfigs["network_name"] {
		t.Errorf("expected network_name to be: %s, got: %s",
			providedConfigs["network_name"], c.NetworkName)
	}
	if c.WatchSourceTimeoutSec != providedConfigs["watch_source_timeout_sec"] {
		t.Errorf("expected watch_source_timeout_sec to be: %d, got: %d",
			providedConfigs["watch_source_timeout_sec"], c.WatchSourceTimeoutSec)
	}
	if c.KeepInputArtifact != providedConfigs["keep_input_artifact"] {
		t.Errorf("expected keep_input_artifact to be: true, got: false")
	}
}

func TestConfig_SSH(t *testing.T) {
	c := new(supervisor.Config)
	configs := getCompleteConfig(t)

	// Test when non-supported communicator is provided (should fail).
	configs["communicator"] = "none"
	_, err := c.Prepare(configs)
	if err == nil {
		t.Errorf("expected an error for non-supported ('none') communicator type")
	}
	configs["communicator"] = "winrm"
	_, err = c.Prepare(configs)
	if err == nil {
		t.Errorf("expected an error for non-supported ('winrm') communicator type")
	}

	// Test if default values are set for SSH communicator.
	configs["communicator"] = ""
	configs["ssh_username"] = ""
	_, err = c.Prepare(configs)
	if err != nil {
		t.Errorf("expected no errors, got: %s", err)
	}
	if c.CommunicatorConfig.Type != "ssh" {
		t.Errorf("expected default communicator to be: 'ssh', got: %q",
			c.CommunicatorConfig.Type)
	}
	if c.CommunicatorConfig.SSHUsername != supervisor.DefaultSSHUsername {
		t.Errorf("expected default ssh_username to be: %q, got: %q",
			supervisor.DefaultSSHUsername, c.CommunicatorConfig.SSHUsername)
	}
}
