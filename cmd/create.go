package cmd

import (
	"github.com/spf13/cobra"
)

func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "",
		Long:  "",
		Run: func(_ *cobra.Command, _ []string) {
		},
	}
	cmd.Hidden = true

	return cmd
}
