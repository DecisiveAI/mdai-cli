package cmd

import (
	"errors"
	"testing"
)

func TestMuteCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "mute command with no args",
			args: []string{"mute"},
			err:  errors.New(`required flag(s) "description", "name", "pipeline" not set`),
		},
		{
			name: "mute command with name without description and pipeline",
			args: []string{"mute", "--name", "test-filter"},
			err:  errors.New(`required flag(s) "description", "pipeline" not set`),
		},
		{
			name: "mute command with name and description without pipeline",
			args: []string{"mute", "--name", "test-filter", "--description", "test filter"},
			err:  errors.New(`required flag(s) "pipeline" not set`),
		},
		{
			name: "mute command with name and pipeline without description",
			args: []string{"mute", "--name", "test-filter", "--pipeline", "logs"},
			err:  errors.New(`required flag(s) "description" not set`),
		},
		{
			name: "mute command with description without name and pipeline",
			args: []string{"mute", "--description", "test-filter"},
			err:  errors.New(`required flag(s) "name", "pipeline" not set`),
		},
		{
			name: "mute command with description and pipeline without name",
			args: []string{"mute", "--description", "test-filter", "--pipeline", "logs"},
			err:  errors.New(`required flag(s) "name" not set`),
		},
		{
			name: "mute command with pipeline without description and name",
			args: []string{"mute", "--pipeline", "logs"},
			err:  errors.New(`required flag(s) "description", "name" not set`),
		},
	}

	errTests.Run(t)
}
