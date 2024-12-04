// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package virtual_machine

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatConfig is an auto-generated flat version of Config.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConfig struct {
	PackerBuildName     *string           `mapstructure:"packer_build_name" cty:"packer_build_name" hcl:"packer_build_name"`
	PackerBuilderType   *string           `mapstructure:"packer_builder_type" cty:"packer_builder_type" hcl:"packer_builder_type"`
	PackerCoreVersion   *string           `mapstructure:"packer_core_version" cty:"packer_core_version" hcl:"packer_core_version"`
	PackerDebug         *bool             `mapstructure:"packer_debug" cty:"packer_debug" hcl:"packer_debug"`
	PackerForce         *bool             `mapstructure:"packer_force" cty:"packer_force" hcl:"packer_force"`
	PackerOnError       *string           `mapstructure:"packer_on_error" cty:"packer_on_error" hcl:"packer_on_error"`
	PackerUserVars      map[string]string `mapstructure:"packer_user_variables" cty:"packer_user_variables" hcl:"packer_user_variables"`
	PackerSensitiveVars []string          `mapstructure:"packer_sensitive_variables" cty:"packer_sensitive_variables" hcl:"packer_sensitive_variables"`
	VCenterServer       *string           `mapstructure:"vcenter_server" cty:"vcenter_server" hcl:"vcenter_server"`
	Username            *string           `mapstructure:"username" cty:"username" hcl:"username"`
	Password            *string           `mapstructure:"password" cty:"password" hcl:"password"`
	InsecureConnection  *bool             `mapstructure:"insecure_connection" cty:"insecure_connection" hcl:"insecure_connection"`
	Datacenter          *string           `mapstructure:"datacenter" cty:"datacenter" hcl:"datacenter"`
	Name                *string           `mapstructure:"name" cty:"name" hcl:"name"`
	NameRegex           *string           `mapstructure:"name_regex" cty:"name_regex" hcl:"name_regex"`
	Template            *bool             `mapstructure:"template" cty:"template" hcl:"template"`
	Node                *string           `mapstructure:"node" cty:"node" hcl:"node"`
	VmTags              []FlatTag         `mapstructure:"vm_tags" cty:"vm_tags" hcl:"vm_tags"`
	Latest              *bool             `mapstructure:"latest" cty:"latest" hcl:"latest"`
}

// FlatMapstructure returns a new FlatConfig.
// FlatConfig is an auto-generated flat version of Config.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Config) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfig)
}

// HCL2Spec returns the hcl spec of a Config.
// This spec is used by HCL to read the fields of Config.
// The decoded values from this spec will then be applied to a FlatConfig.
func (*FlatConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"packer_build_name":          &hcldec.AttrSpec{Name: "packer_build_name", Type: cty.String, Required: false},
		"packer_builder_type":        &hcldec.AttrSpec{Name: "packer_builder_type", Type: cty.String, Required: false},
		"packer_core_version":        &hcldec.AttrSpec{Name: "packer_core_version", Type: cty.String, Required: false},
		"packer_debug":               &hcldec.AttrSpec{Name: "packer_debug", Type: cty.Bool, Required: false},
		"packer_force":               &hcldec.AttrSpec{Name: "packer_force", Type: cty.Bool, Required: false},
		"packer_on_error":            &hcldec.AttrSpec{Name: "packer_on_error", Type: cty.String, Required: false},
		"packer_user_variables":      &hcldec.AttrSpec{Name: "packer_user_variables", Type: cty.Map(cty.String), Required: false},
		"packer_sensitive_variables": &hcldec.AttrSpec{Name: "packer_sensitive_variables", Type: cty.List(cty.String), Required: false},
		"vcenter_server":             &hcldec.AttrSpec{Name: "vcenter_server", Type: cty.String, Required: false},
		"username":                   &hcldec.AttrSpec{Name: "username", Type: cty.String, Required: false},
		"password":                   &hcldec.AttrSpec{Name: "password", Type: cty.String, Required: false},
		"insecure_connection":        &hcldec.AttrSpec{Name: "insecure_connection", Type: cty.Bool, Required: false},
		"datacenter":                 &hcldec.AttrSpec{Name: "datacenter", Type: cty.String, Required: false},
		"name":                       &hcldec.AttrSpec{Name: "name", Type: cty.String, Required: false},
		"name_regex":                 &hcldec.AttrSpec{Name: "name_regex", Type: cty.String, Required: false},
		"template":                   &hcldec.AttrSpec{Name: "template", Type: cty.Bool, Required: false},
		"node":                       &hcldec.AttrSpec{Name: "node", Type: cty.String, Required: false},
		"vm_tags":                    &hcldec.BlockListSpec{TypeName: "vm_tags", Nested: hcldec.ObjectSpec((*FlatTag)(nil).HCL2Spec())},
		"latest":                     &hcldec.AttrSpec{Name: "latest", Type: cty.Bool, Required: false},
	}
	return s
}

// FlatDatasourceOutput is an auto-generated flat version of DatasourceOutput.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatDatasourceOutput struct {
	VmName *string `mapstructure:"vm_name" cty:"vm_name" hcl:"vm_name"`
}

// FlatMapstructure returns a new FlatDatasourceOutput.
// FlatDatasourceOutput is an auto-generated flat version of DatasourceOutput.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*DatasourceOutput) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatDatasourceOutput)
}

// HCL2Spec returns the hcl spec of a DatasourceOutput.
// This spec is used by HCL to read the fields of DatasourceOutput.
// The decoded values from this spec will then be applied to a FlatDatasourceOutput.
func (*FlatDatasourceOutput) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"vm_name": &hcldec.AttrSpec{Name: "vm_name", Type: cty.String, Required: false},
	}
	return s
}

// FlatTag is an auto-generated flat version of Tag.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatTag struct {
	Name     *string `mapstructure:"name" required:"true" cty:"name" hcl:"name"`
	Category *string `mapstructure:"category" required:"true" cty:"category" hcl:"category"`
}

// FlatMapstructure returns a new FlatTag.
// FlatTag is an auto-generated flat version of Tag.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Tag) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatTag)
}

// HCL2Spec returns the hcl spec of a Tag.
// This spec is used by HCL to read the fields of Tag.
// The decoded values from this spec will then be applied to a FlatTag.
func (*FlatTag) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"name":     &hcldec.AttrSpec{Name: "name", Type: cty.String, Required: false},
		"category": &hcldec.AttrSpec{Name: "category", Type: cty.String, Required: false},
	}
	return s
}
