package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "disable -m|--module MODULE",
		Short:   "disable a module",
		Long:    `disable a module`,
		Example: `  mdai disable --module datalyzer`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			module, _ := cmd.Flags().GetString("module")
			if module != "" && !slices.Contains(SupportedModules, module) {
				return fmt.Errorf(`module "%s" is not supported for disabling`, module)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			module, _ := cmd.Flags().GetString("module")

			switch module {
			case "datalyzer":
				if err := operator.DisableDatalyzer(); err != nil {
					return err
				}
			}

			fmt.Printf("%s module disabled successfully.\n", module)
			return nil
		},
	}
	cmd.Flags().String("module", "", "module to disable ["+strings.Join(SupportedModules, ", ")+"]")

	_ = cmd.MarkFlagRequired("module")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
