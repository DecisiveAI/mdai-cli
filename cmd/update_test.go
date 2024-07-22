package cmd

import (
	"errors"
	"testing"
)

func TestUpdateCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "update command without config or file flags",
			args: []string{"update"},
			err:  errors.New("at least one of the flags in the group [file config] is required"),
		},
		{
			name: "update command with both config and file flags",
			args: []string{"update", "--config", "otel", "--file", "otel.yaml"},
			err:  errors.New("if any flags in the group [file config] are set none of the others can be; [config file] were all set"),
		},
		{
			name: "update command with invalid block flag",
			args: []string{"update", "--config", "otel", "--block", "foo"},
			err:  errors.New("invalid block: foo"),
		},
		{
			name: "update command with invalid phase flag",
			args: []string{"update", "--config", "otel", "--phase", "foo"},
			err:  errors.New("invalid phase: foo"),
		},
	}

	errTests.Run(t)
}
