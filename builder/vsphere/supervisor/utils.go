package supervisor

import (
	"fmt"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type PackerLogger struct {
	ui packersdk.Ui
}

func (pl *PackerLogger) Info(msg string, args ...interface{}) {
	pl.ui.Message(fmt.Sprintf(msg, args...))
}
func (pl *PackerLogger) Error(msg string, args ...interface{}) {
	pl.ui.Error(fmt.Sprintf(msg, args...))
}
