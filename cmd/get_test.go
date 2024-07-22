package cmd

import (
	"errors"
	"testing"
)

func TestGetCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "get command without config flag",
			args: []string{"get"},
			err:  errors.New(`required flag(s) "config" not set`),
		},
		{
			name: "get command with invalid config flag",
			args: []string{"get", "--config", "foo"},
			err:  errors.New("config type foo is not supported"),
		},
	}

	errTests.Run(t)
}
