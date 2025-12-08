// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"net"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepHTTPIPDiscover_Run(t *testing.T) {
	state := new(multistep.BasicStateBag)
	step := new(StepHTTPIPDiscover)

	// without setting HTTPIP
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("unexpected action: expected '%#v', but returned '%#v'", multistep.ActionContinue, action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("unexpected error: expected no error")
	}
	_, ok := state.GetOk("http_ip")
	if !ok {
		t.Fatalf("unexpected state: '%s' not found", "http_ip")
	}

	// setting HTTPIP
	ip := "10.0.2.2"
	step = &StepHTTPIPDiscover{
		HTTPIP: ip,
	}
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("unexpected action: expected '%#v', but returned '%#v'", multistep.ActionContinue, action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("unexpected error: expected no error")
	}
	httpIp, ok := state.GetOk("http_ip")
	if !ok {
		t.Fatalf("unexpected state: '%s' not found", "http_ip")
	}
	if httpIp != ip {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", ip, httpIp)
	}

	_, ipNet, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	step = new(StepHTTPIPDiscover)
	step.Network = ipNet

	// without setting HTTPIP with Network
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("unexpected action: expected '%#v', but returned '%#v'", multistep.ActionContinue, action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("unexpected error: expected no error")
	}
	_, ok = state.GetOk("http_ip")
	if !ok {
		t.Fatalf("unexpected state: '%s' not found", "http_ip")
	}

	// setting HTTPIP with Network
	step = &StepHTTPIPDiscover{
		HTTPIP:  ip,
		Network: ipNet,
	}
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("unexpected action: expected '%#v', but returned '%#v'", multistep.ActionContinue, action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("unexpected error: expected no error")
	}
	httpIp, ok = state.GetOk("http_ip")
	if !ok {
		t.Fatalf("unexpected state: '%s' not found", "http_ip")
	}
	if httpIp != ip {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", ip, httpIp)
	}
}
