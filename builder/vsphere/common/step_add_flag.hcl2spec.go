// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package common

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatFlagConfig is an auto-generated flat version of FlagConfig.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatFlagConfig struct {
	VbsEnabled  *bool `mapstructure:"vbs_enabled" cty:"vbs_enabled" hcl:"vbs_enabled"`
	VvtdEnabled *bool `mapstructure:"vvtd_enabled" cty:"vvtd_enabled" hcl:"vvtd_enabled"`
}

// FlatMapstructure returns a new FlatFlagConfig.
// FlatFlagConfig is an auto-generated flat version of FlagConfig.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*FlagConfig) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatFlagConfig)
}

// HCL2Spec returns the hcl spec of a FlagConfig.
// This spec is used by HCL to read the fields of FlagConfig.
// The decoded values from this spec will then be applied to a FlatFlagConfig.
func (*FlatFlagConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"vbs_enabled":  &hcldec.AttrSpec{Name: "vbs_enabled", Type: cty.Bool, Required: false},
		"vvtd_enabled": &hcldec.AttrSpec{Name: "vvtd_enabled", Type: cty.Bool, Required: false},
	}
	return s
}
