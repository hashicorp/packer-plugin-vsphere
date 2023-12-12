// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package common

import (
	"io/fs"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatExportConfig is an auto-generated flat version of ExportConfig.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatExportConfig struct {
	Name       *string      `mapstructure:"name" cty:"name" hcl:"name"`
	Force      *bool        `mapstructure:"force" cty:"force" hcl:"force"`
	ImageFiles *bool        `mapstructure:"image_files" cty:"image_files" hcl:"image_files"`
	Manifest   *string      `mapstructure:"manifest" cty:"manifest" hcl:"manifest"`
	OutputDir  *string      `mapstructure:"output_directory" required:"false" cty:"output_directory" hcl:"output_directory"`
	DirPerm    *fs.FileMode `mapstructure:"directory_permission" required:"false" cty:"directory_permission" hcl:"directory_permission"`
	Options    []string     `mapstructure:"options" cty:"options" hcl:"options"`
}

// FlatMapstructure returns a new FlatExportConfig.
// FlatExportConfig is an auto-generated flat version of ExportConfig.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*ExportConfig) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatExportConfig)
}

// HCL2Spec returns the hcl spec of a ExportConfig.
// This spec is used by HCL to read the fields of ExportConfig.
// The decoded values from this spec will then be applied to a FlatExportConfig.
func (*FlatExportConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"name":                 &hcldec.AttrSpec{Name: "name", Type: cty.String, Required: false},
		"force":                &hcldec.AttrSpec{Name: "force", Type: cty.Bool, Required: false},
		"image_files":          &hcldec.AttrSpec{Name: "image_files", Type: cty.Bool, Required: false},
		"manifest":             &hcldec.AttrSpec{Name: "manifest", Type: cty.String, Required: false},
		"output_directory":     &hcldec.AttrSpec{Name: "output_directory", Type: cty.String, Required: false},
		"directory_permission": &hcldec.AttrSpec{Name: "directory_permission", Type: cty.Number, Required: false},
		"options":              &hcldec.AttrSpec{Name: "options", Type: cty.List(cty.String), Required: false},
	}
	return s
}
