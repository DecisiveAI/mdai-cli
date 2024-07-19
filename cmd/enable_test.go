package cmd

import (
	"errors"
	"testing"
)

func TestEnableCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "enable command with no flags",
			args: []string{"enable"},
			err:  errors.New(`required flag(s) "module" not set`),
		},
		{
			name: "enable command with invalid module flag",
			args: []string{"enable", "--module", "foo"},
			err:  errors.New(`module "foo" is not supported for enabling`),
		},
	}

	errTests.Run(t)
}
