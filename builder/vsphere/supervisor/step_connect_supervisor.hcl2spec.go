// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package supervisor

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatConnectSupervisorConfig is an auto-generated flat version of ConnectSupervisorConfig.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConnectSupervisorConfig struct {
	KubeconfigPath      *string `mapstructure:"kubeconfig_path" cty:"kubeconfig_path" hcl:"kubeconfig_path"`
	SupervisorNamespace *string `mapstructure:"supervisor_namespace" cty:"supervisor_namespace" hcl:"supervisor_namespace"`
}

// FlatMapstructure returns a new FlatConnectSupervisorConfig.
// FlatConnectSupervisorConfig is an auto-generated flat version of ConnectSupervisorConfig.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*ConnectSupervisorConfig) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConnectSupervisorConfig)
}

// HCL2Spec returns the hcl spec of a ConnectSupervisorConfig.
// This spec is used by HCL to read the fields of ConnectSupervisorConfig.
// The decoded values from this spec will then be applied to a FlatConnectSupervisorConfig.
func (*FlatConnectSupervisorConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"kubeconfig_path":      &hcldec.AttrSpec{Name: "kubeconfig_path", Type: cty.String, Required: false},
		"supervisor_namespace": &hcldec.AttrSpec{Name: "supervisor_namespace", Type: cty.String, Required: false},
	}
	return s
}
