// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package common

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatRemoveNetworkAdapterConfig is an auto-generated flat version of RemoveNetworkAdapterConfig.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatRemoveNetworkAdapterConfig struct {
	RemoveNetworkAdapter *bool `mapstructure:"remove_network_adapter" cty:"remove_network_adapter" hcl:"remove_network_adapter"`
}

// FlatMapstructure returns a new FlatRemoveNetworkAdapterConfig.
// FlatRemoveNetworkAdapterConfig is an auto-generated flat version of RemoveNetworkAdapterConfig.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*RemoveNetworkAdapterConfig) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatRemoveNetworkAdapterConfig)
}

// HCL2Spec returns the hcl spec of a RemoveNetworkAdapterConfig.
// This spec is used by HCL to read the fields of RemoveNetworkAdapterConfig.
// The decoded values from this spec will then be applied to a FlatRemoveNetworkAdapterConfig.
func (*FlatRemoveNetworkAdapterConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"remove_network_adapter": &hcldec.AttrSpec{Name: "remove_network_adapter", Type: cty.Bool, Required: false},
	}
	return s
}
