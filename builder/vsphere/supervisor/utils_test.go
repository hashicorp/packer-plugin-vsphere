package supervisor_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

// TODO: Add testings for utils.go

// Utility functions used in other testings.

func newBasicTestState(writer *bytes.Buffer) *multistep.BasicStateBag {
	state := new(multistep.BasicStateBag)
	ui := &packersdk.BasicUi{
		Writer: writer,
	}
	state.Put("logger", &supervisor.PackerLogger{UI: ui})

	return state
}

func checkOutputLines(t *testing.T, writer *bytes.Buffer, expectedLines []string) {
	for _, expected := range expectedLines {
		actual, err := writer.ReadString('\n')
		actual = strings.TrimSpace(actual)
		if err != nil {
			t.Fatalf("Failed to read line from writer, err: %s", err.Error())
		}
		if actual != expected {
			t.Fatalf("Expected output '%s' but got '%s'", expected, actual)
		}
	}
}
