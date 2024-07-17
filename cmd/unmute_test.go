package cmd

import (
	"errors"
	"testing"
)

func TestUnmuteCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "unmute command with no args",
			args: []string{"unmute"},
			err:  errors.New(`required flag(s) "name" not set`),
		},
	}

	errTests.Run(t)
}
