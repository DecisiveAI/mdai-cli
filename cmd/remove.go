package cmd

import (
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "",
	Long:  ``,
	Run: func(_ *cobra.Command, _ []string) {
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Hidden = true
}
