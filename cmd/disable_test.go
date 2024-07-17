package cmd

import (
	"errors"
	"testing"
)

func TestDisableCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "disable command with no args",
			args: []string{"disable"},
			err:  errors.New(`required flag(s) "module" not set`),
		},
		{
			name: "disable command with invalid module",
			args: []string{"disable", "--module", "foo"},
			err:  errors.New(`module "foo" is not supported for disabling`),
		},
	}

	errTests.Run(t)
}
