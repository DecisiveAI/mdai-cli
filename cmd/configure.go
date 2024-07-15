package cmd

import (
	"github.com/spf13/cobra"
)

func NewConfigureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "",
		Long:  "",
		Run: func(_ *cobra.Command, _ []string) {
		},
	}
	cmd.Hidden = true

	return cmd
}
