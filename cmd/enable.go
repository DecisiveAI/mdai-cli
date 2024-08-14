package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "enable -m|--module MODULE",
		Short:   "enable a module",
		Long:    `enable one of the supported modules`,
		Example: `  mdai enable --module datalyzer`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			module, _ := cmd.Flags().GetString("module")
			if module != "" && !slices.Contains(SupportedModules, module) {
				return fmt.Errorf(`module "%s" is not supported for enabling`, module)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			module, _ := cmd.Flags().GetString("module")

			switch module {
			case "datalyzer":
				if err := operator.EnableDatalyzer(ctx); err != nil {
					return err
				}
			}

			fmt.Printf("%s module enabled successfully.\n", module)
			return nil
		},
	}
	cmd.Flags().String("module", "", "module to enable ["+strings.Join(SupportedModules, ", ")+"]")

	_ = cmd.MarkFlagRequired("module")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
