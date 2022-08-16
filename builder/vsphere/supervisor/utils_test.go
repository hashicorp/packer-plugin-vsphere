package supervisor_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
)

func TestCheckRequiredStates(t *testing.T) {
	state := newBasicTestState(nil)
	err := supervisor.CheckRequiredStates(state, "logger")
	if err != nil {
		t.Errorf("Expected no error but got: %s", err.Error())
	}

	state.Put("test-key-1", "test-val-1")
	state.Put("test-key-2", "test-val-2")
	err = supervisor.CheckRequiredStates(state, "test-key-1", "test-key-2")
	if err != nil {
		t.Errorf("Expected no error but got: %s", err.Error())
	}

	expectErr := supervisor.CheckRequiredStates(state, "test-key-non-exist")
	if expectErr == nil {
		t.Errorf("Expected error but got nil")
	}
}

// Utility functions that are used in multiple test code.

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
