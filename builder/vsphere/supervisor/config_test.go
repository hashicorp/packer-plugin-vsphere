package supervisor_test

import (
	"testing"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func getCompleteConfig() map[string]interface{} {
	return map[string]interface{}{
		"image_name":               "test-image",
		"class_name":               "test-class",
		"storage_class":            "test-storage",
		"kubeconfig_path":          "test-kubeconfig",
		"k8s_namespace":            "test-namespace",
		"source_name":              "test-source",
		"network_type":             "test-networkType",
		"network_name":             "test-networkName",
		"watch_source_timeout_sec": int64(60),
		"keep_source":              true,
	}
}

func getMinimaConfig() map[string]interface{} {
	return map[string]interface{}{
		"image_name":    "test-image",
		"class_name":    "test-class",
		"storage_class": "test-storage",
	}
}

func TestSupervisorConfig_Minimal(t *testing.T) {
	c := new(supervisor.Config)
	minConfigs := getMinimaConfig()
	warns, err := c.Prepare(minConfigs)
	if len(warns) != 0 {
		t.Errorf("expected no warnings, got: %#v", warns)
	}
	if err != nil {
		t.Errorf("expected no errors, got: %s", err)
	}
}

func TestSupervisorConfig_Required(t *testing.T) {
	c := new(supervisor.Config)
	minConfigs := getMinimaConfig()
	for key, val := range minConfigs {
		minConfigs[key] = ""
		_, err := c.Prepare(minConfigs)
		if err == nil {
			t.Errorf("expected error for required config: %s", key)
		}
		minConfigs[key] = val
	}
}

func TestSupervisorConfig_Complete(t *testing.T) {
	c := new(supervisor.Config)
	allConfigs := getCompleteConfig()
	warns, err := c.Prepare(allConfigs)
	if len(warns) != 0 {
		t.Errorf("expected no warnings, got: %#v", warns)
	}
	if err != nil {
		t.Errorf("expected no errors, got: %s", err)
	}
}

func TestSupervisorConfig_Values(t *testing.T) {
	c := new(supervisor.Config)
	providedConfigs := getCompleteConfig()
	warns, err := c.Prepare(providedConfigs)
	if len(warns) != 0 {
		t.Fatalf("expected no warnings, got: %#v", warns)
	}
	if err != nil {
		t.Fatalf("expected no errors, got: %s", err)
	}

	if c.ImageName != providedConfigs["image_name"] {
		t.Errorf("expected image_name to be: %s, got: %s", providedConfigs["image_name"], c.ImageName)
	}
	if c.ClassName != providedConfigs["class_name"] {
		t.Errorf("expected class_name to be: %s, got: %s", providedConfigs["class_name"], c.ClassName)
	}
	if c.StorageClass != providedConfigs["storage_class"] {
		t.Errorf("expected storage_class to be: %s, got: %s", providedConfigs["storage_class"], c.StorageClass)
	}
	if c.KubeconfigPath != providedConfigs["kubeconfig_path"] {
		t.Errorf("expected kubeconfig_path to be: %s, got: %s", providedConfigs["kubeconfig_path"], c.KubeconfigPath)
	}
	if c.K8sNamespace != providedConfigs["k8s_namespace"] {
		t.Errorf("expected k8s_namespace to be: %s, got: %s", providedConfigs["k8s_namespace"], c.K8sNamespace)
	}
	if c.SourceName != providedConfigs["source_name"] {
		t.Errorf("expected source_name to be: %s, got: %s", providedConfigs["source_name"], c.SourceName)
	}
	if c.NetworkType != providedConfigs["network_type"] {
		t.Errorf("expected network_type to be: %s, got: %s", providedConfigs["network_type"], c.NetworkType)
	}
	if c.NetworkName != providedConfigs["network_name"] {
		t.Errorf("expected network_name to be: %s, got: %s", providedConfigs["network_name"], c.NetworkName)
	}
	if c.TimeoutSecond != providedConfigs["watch_source_timeout_sec"].(int64) {
		t.Errorf("expected watch_source_timeout_sec to be: %d, got: %d", providedConfigs["watch_source_timeout_sec"].(int64), c.TimeoutSecond)
	}
	if c.KeepSource != providedConfigs["keep_source"].(bool) {
		t.Errorf("expected keep_source to be: true, got: false")
	}
}
