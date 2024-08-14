package cmd

import (
	"errors"
	"testing"
)

func TestUninstallCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "uninstall command with args",
			args: []string{"uninstall", "demo"},
			err:  errors.New(`unknown command "demo" for "mdai uninstall"`),
		},
		{
			name: "uninstall with both --debug and --quiet",
			args: []string{"uninstall", "--debug", "--quiet"},
			err:  errors.New(`if any flags in the group [debug quiet] are set none of the others can be; [debug quiet] were all set`),
		},
	}

	errTests.Run(t)
}
