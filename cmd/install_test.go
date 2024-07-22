package cmd

import (
	"errors"
	"testing"
)

func TestInstallCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "install command with args",
			args: []string{"install", "demo"},
			err:  errors.New(`unknown command "demo" for "mdai install"`),
		},
	}

	errTests.Run(t)
}
