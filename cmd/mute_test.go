package cmd

import (
	"errors"
	"testing"
)

func TestMuteCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "mute command with no flags",
			args: []string{"mute"},
			err:  errors.New(`required flag(s) "description", "name", "pipeline" not set`),
		},
		{
			name: "mute command with name flag without description and pipeline flags",
			args: []string{"mute", "--name", "test-filter"},
			err:  errors.New(`required flag(s) "description", "pipeline" not set`),
		},
		{
			name: "mute command with name and description flags without pipeline flag",
			args: []string{"mute", "--name", "test-filter", "--description", "test filter"},
			err:  errors.New(`required flag(s) "pipeline" not set`),
		},
		{
			name: "mute command with name and pipeline flags without description flag",
			args: []string{"mute", "--name", "test-filter", "--pipeline", "logs"},
			err:  errors.New(`required flag(s) "description" not set`),
		},
		{
			name: "mute command with description flag without name and pipeline flags",
			args: []string{"mute", "--description", "test-filter"},
			err:  errors.New(`required flag(s) "name", "pipeline" not set`),
		},
		{
			name: "mute command with description and pipeline flags without name flag",
			args: []string{"mute", "--description", "test-filter", "--pipeline", "logs"},
			err:  errors.New(`required flag(s) "name" not set`),
		},
		{
			name: "mute command with pipeline flag without description and name flags",
			args: []string{"mute", "--pipeline", "logs"},
			err:  errors.New(`required flag(s) "description", "name" not set`),
		},
	}

	errTests.Run(t)
}
