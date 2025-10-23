// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package virtualmachine

import (
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
	"github.com/vmware/govmomi/simulator"

	commonT "github.com/hashicorp/packer-plugin-vsphere/datasource/common/testing"
)

func TestExecute(t *testing.T) {
	machinesToPrepare := []commonT.SimulatedVMConfig{
		{
			Name: "first-vm",
			Tags: []commonT.Tag{
				{
					Category: "operating-system-class",
					Name:     "Linux",
				},
			},
		}, {
			Name: "second-vm",
			Tags: []commonT.Tag{
				{
					Category: "operating-system-class",
					Name:     "Linux",
				},
				{
					Category: "security-team",
					Name:     "red",
				},
				{
					Category: "security-team",
					Name:     "blue",
				},
			},
			Template: true,
		}, {
			Name: "machine-three",
			Tags: []commonT.Tag{
				{
					Category: "operating-system-class",
					Name:     "Linux",
				},
				{
					Category: "security-team",
					Name:     "blue",
				},
			},
			CreationTime: time.Now().AddDate(0, 0, 1),
		},
	}

	model := simulator.VPX()
	model.Datacenter = 2
	model.Machine = 8

	vcSim, err := commonT.NewVCenterSimulator(model)
	if err != nil {
		t.Fatalf("error creating vCenter simulator: %s", err)
	}
	defer vcSim.Stop()

	err = vcSim.CustomizeSimulator(machinesToPrepare)
	if err != nil {
		t.Fatalf("error customizing simulator: %s", err)
	}

	simulatorPassword, _ := vcSim.Server.URL.User.Password()
	connectConfig := common.ConnectConfig{
		VCenterServer:      vcSim.Server.URL.Host,
		Username:           vcSim.Server.URL.User.Username(),
		Password:           simulatorPassword,
		InsecureConnection: true,
		Datacenter:         vcSim.Datacenter.Name(),
	}

	dsTestConfigs := []struct {
		name          string
		expectFailure bool
		expectVmName  string
		config        Config
	}{
		{
			name:          "first-vm was found by name, no error",
			expectFailure: false,
			expectVmName:  "first-vm",
			config: Config{
				Name: "first-vm",
			},
		},
		{
			name:          "no machines match the filter, error",
			expectFailure: true,
			expectVmName:  "",
			config: Config{
				Name: "firstest-vm",
			},
		},
		{
			name:          "second-vm was found by the regex, no error",
			expectFailure: false,
			expectVmName:  "second-vm",
			config: Config{
				NameRegex: "^seco.*m$",
			},
		},
		{
			name:          "multiple machines match the regex, but latest not used, error",
			expectFailure: true,
			expectVmName:  "",
			config: Config{
				NameRegex: ".*-vm",
			},
		},
		{
			name:          "multiple guests match the regex and latest used, no error",
			expectFailure: false,
			expectVmName:  "machine-three",
			config: Config{
				NameRegex: "^[^_]+$",
				Latest:    true,
			},
		},
		{
			name:          "found machine that is a template, no error",
			expectFailure: false,
			expectVmName:  "second-vm",
			config: Config{
				Template: true,
			},
		},
		{
			name:          "found multiple machines at the host, error",
			expectFailure: true,
			expectVmName:  "",
			config: Config{
				Host: "DC0_H0",
			},
		},
		{
			name:          "cluster host not found, error",
			expectFailure: true,
			expectVmName:  "",
			config: Config{
				Host: "unexpected_host",
			},
		},
		{
			name:          "found machine with defined set of tags, no error",
			expectFailure: false,
			expectVmName:  "second-vm",
			config: Config{
				Tags: []Tag{
					{
						Category: "security-team",
						Name:     "blue",
					},
					{
						Category: "security-team",
						Name:     "red",
					},
				},
			},
		},
		{
			name:          "found multiple machines with defined set of tags, error",
			expectFailure: true,
			expectVmName:  "",
			config: Config{
				Tags: []Tag{
					{
						Category: "operating-system-class",
						Name:     "Linux",
					},
				},
			},
		},
	}

	for _, testConfig := range dsTestConfigs {
		t.Run(testConfig.name, func(t *testing.T) {
			testConfig.config.ConnectConfig = connectConfig

			ds := Datasource{
				config: testConfig.config,
			}
			err := ds.Configure()
			if err != nil {
				t.Fatalf("Failed to configure datasource: %s", err)
			}

			result, err := ds.Execute()
			if err != nil && !testConfig.expectFailure {
				t.Fatalf("unexpected failure: %s", err)
			}
			if err == nil && testConfig.expectFailure {
				t.Errorf("expected failure, but execution succeeded")
			}
			if err == nil {
				vmName := result.GetAttr("vm_name").AsString()
				if vmName != testConfig.expectVmName {
					t.Errorf("expected vm name `%s`, but got `%s`", testConfig.expectVmName, vmName)
				}
			}
		})
	}
}
