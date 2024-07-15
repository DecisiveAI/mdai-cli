package cmd

import (
	"github.com/spf13/cobra"
)

func NewRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "",
		Long:  ``,
		Run: func(_ *cobra.Command, _ []string) {
		},
	}
	cmd.Hidden = true

	return cmd
}
