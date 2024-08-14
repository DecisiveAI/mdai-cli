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
		{
			name: "install with both --debug and --quiet",
			args: []string{"install", "--debug", "--quiet"},
			err:  errors.New(`if any flags in the group [debug quiet] are set none of the others can be; [debug quiet] were all set`),
		},
	}

	errTests.Run(t)
}
