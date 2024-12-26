// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package supervisor

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatCreateSourceConfig is an auto-generated flat version of CreateSourceConfig.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatCreateSourceConfig struct {
	ClassName         *string `mapstructure:"class_name" required:"true" cty:"class_name" hcl:"class_name"`
	StorageClass      *string `mapstructure:"storage_class" required:"true" cty:"storage_class" hcl:"storage_class"`
	ImageName         *string `mapstructure:"image_name" cty:"image_name" hcl:"image_name"`
	SourceName        *string `mapstructure:"source_name" cty:"source_name" hcl:"source_name"`
	KeepInputArtifact *bool   `mapstructure:"keep_input_artifact" cty:"keep_input_artifact" hcl:"keep_input_artifact"`
	BootstrapProvider *string `mapstructure:"bootstrap_provider" cty:"bootstrap_provider" hcl:"bootstrap_provider"`
	BootstrapDataFile *string `mapstructure:"bootstrap_data_file" cty:"bootstrap_data_file" hcl:"bootstrap_data_file"`
	GuestOSType       *string `mapstructure:"guest_os_type" cty:"guest_os_type" hcl:"guest_os_type"`
	IsoBootDiskSize   *string `mapstructure:"iso_boot_disk_size" cty:"iso_boot_disk_size" hcl:"iso_boot_disk_size"`
}

// FlatMapstructure returns a new FlatCreateSourceConfig.
// FlatCreateSourceConfig is an auto-generated flat version of CreateSourceConfig.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*CreateSourceConfig) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatCreateSourceConfig)
}

// HCL2Spec returns the hcl spec of a CreateSourceConfig.
// This spec is used by HCL to read the fields of CreateSourceConfig.
// The decoded values from this spec will then be applied to a FlatCreateSourceConfig.
func (*FlatCreateSourceConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"class_name":          &hcldec.AttrSpec{Name: "class_name", Type: cty.String, Required: false},
		"storage_class":       &hcldec.AttrSpec{Name: "storage_class", Type: cty.String, Required: false},
		"image_name":          &hcldec.AttrSpec{Name: "image_name", Type: cty.String, Required: false},
		"source_name":         &hcldec.AttrSpec{Name: "source_name", Type: cty.String, Required: false},
		"keep_input_artifact": &hcldec.AttrSpec{Name: "keep_input_artifact", Type: cty.Bool, Required: false},
		"bootstrap_provider":  &hcldec.AttrSpec{Name: "bootstrap_provider", Type: cty.String, Required: false},
		"bootstrap_data_file": &hcldec.AttrSpec{Name: "bootstrap_data_file", Type: cty.String, Required: false},
		"guest_os_type":       &hcldec.AttrSpec{Name: "guest_os_type", Type: cty.String, Required: false},
		"iso_boot_disk_size":  &hcldec.AttrSpec{Name: "iso_boot_disk_size", Type: cty.String, Required: false},
	}
	return s
}
