// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere_template

import (
	"testing"
)

func getTestConfig() Config {
	return Config{
		Username: "administrator@vsphere.local",
		Password: "password",
		Host:     "vcenter.example.com",
	}
}

func TestConfigure_Good(t *testing.T) {
	var p PostProcessor

	config := getTestConfig()

	err := p.Configure(config)
	if err != nil {
		t.Errorf("error: %s", err)
	}
}

func TestConfigure_ReRegisterVM(t *testing.T) {
	var p PostProcessor

	config := getTestConfig()

	err := p.Configure(config)
	if err != nil {
		t.Errorf("error: %s", err)
	}

	if p.config.ReregisterVM.False() {
		t.Errorf("error: should be unset, not false")
	}
}
