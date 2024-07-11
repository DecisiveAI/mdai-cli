package cmd

import (
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "",
	Long:  "",
	Run: func(_ *cobra.Command, _ []string) {
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
	configureCmd.Hidden = true
}
