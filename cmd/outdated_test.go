package cmd

import (
	"errors"
	"testing"
)

func TestOutdatedCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "outdated command with args",
			args: []string{"outdated", "packages"},
			err:  errors.New(`unknown command "packages" for "mdai outdated"`),
		},
	}

	errTests.Run(t)
}
