package cmd

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "",
	Long:  "",
	Run: func(_ *cobra.Command, _ []string) {
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Hidden = true
}
