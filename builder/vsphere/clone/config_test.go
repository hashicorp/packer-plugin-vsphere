// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clone

import (
	"testing"
	"time"
)

func TestCloneConfig_MinimalConfig(t *testing.T) {
	c := new(Config)
	warns, errs := c.Prepare(minimalConfig())
	testConfigOk(t, warns, errs)
}

func TestCloneConfig_MandatoryParameters(t *testing.T) {
	params := []string{"vcenter_server", "username", "password", "template", "vm_name", "host"}
	for _, param := range params {
		raw := minimalConfig()
		raw[param] = ""
		c := new(Config)
		warns, err := c.Prepare(raw)
		testConfigErr(t, param, warns, err)
	}
}

func TestCloneConfig_Timeout(t *testing.T) {
	raw := minimalConfig()
	raw["shutdown_timeout"] = "3m"
	conf := new(Config)
	warns, err := conf.Prepare(raw)
	testConfigOk(t, warns, err)
	if conf.Timeout != 3*time.Minute {
		t.Fatalf("unexpected result: expected '3m', but returned '%v'", conf.Timeout)
	}
}

func TestCloneConfig_RAMReservation(t *testing.T) {
	raw := minimalConfig()
	raw["RAM_reservation"] = 1000
	raw["RAM_reserve_all"] = true
	c := new(Config)
	warns, err := c.Prepare(raw)
	testConfigErr(t, "RAM_reservation", warns, err)
}

func minimalConfig() map[string]interface{} {
	return map[string]interface{}{
		"vcenter_server": "vcenter.example.com",
		"username":       "administrator@vsphere.local",
		"password":       "VMw@re1!",
		"template":       "ubuntu",
		"vm_name":        "vm-01",
		"host":           "esxi-01.example.com",
		"ssh_username":   "root",
		"ssh_password":   "VMw@re1!",
	}
}

func testConfigOk(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Errorf("unexpected warning: %#v", warns)
	}
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func testConfigErr(t *testing.T, context string, warns []string, err error) {
	if len(warns) > 0 {
		t.Errorf("unexpected warning: %#v", warns)
	}
	if err == nil {
		t.Errorf("unexpected result: expected '%s', but returned 'nil'", context)
	}
}
