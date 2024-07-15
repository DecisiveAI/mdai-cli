package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCmd(t *testing.T) {
	cmd := NewGetCommand()
	cmd.SetErr(new(bytes.Buffer))

	t.Run("get command without config", func(t *testing.T) {
		cmd.SetArgs([]string{"get"})
		err := cmd.Execute()
		require.Equal(t, err, errors.New("config is required"))
	})

	t.Run("get command with invalid config", func(t *testing.T) {
		cmd.SetArgs([]string{"get", "--config", "foo"})
		err := cmd.Execute()
		require.Equal(t, err, errors.New("config type foo is not supported"))
	})
}
