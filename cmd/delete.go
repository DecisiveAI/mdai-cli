package cmd

import (
	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "",
		Long:  ``,
		Run: func(_ *cobra.Command, _ []string) {
		},
	}
	cmd.Hidden = true

	return cmd
}
