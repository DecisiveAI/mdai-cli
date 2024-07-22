package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCmdErr struct {
	name string
	args []string
	err  error
}

type testCmdErrs []testCmdErr

func (errTests testCmdErrs) Run(t *testing.T) {
	t.Helper()
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := NewRootCommand()
			require.NoError(t, err)
			cmd.SetErr(new(bytes.Buffer))
			o := new(bytes.Buffer)
			cmd.SetOut(o)
			cmd.SetArgs(tt.args)
			err = cmd.Execute()
			require.Equal(t, "", o.String(), "unexpected output")
			require.Equal(t, tt.err, err)
		})
	}
}

func TestRootCommand(t *testing.T) {
	cmd, err := NewRootCommand()
	require.NoError(t, err)

	require.True(t, cmd.HasSubCommands(), "root command should have subcommands")
	require.True(t, cmd.AllChildCommandsHaveGroup(), "root command should have all child commands belonging to a group")
	require.False(t, cmd.Hidden, "root command should not be hidden")
}
