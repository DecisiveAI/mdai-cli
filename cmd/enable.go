package cmd

import (
	"fmt"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewEnableCommand() *cobra.Command {
	flags := enableFlags{}
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "enable -m|--module MODULE",
		Short:   "enable a module",
		Long:    `enable one of the supported modules`,
		Example: `  mdai enable --module datalyzer`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			switch flags.module {
			case "datalyzer":
				if err := operator.EnableDatalyzer(ctx); err != nil {
					return err
				}
			default:
				return fmt.Errorf(`module "%s" is not supported for enabling`, flags.module)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s module enabled successfully.\n", flags.module)
			return nil
		},
	}
	cmd.Flags().StringVar(&flags.module, "module", "", "module to enable ["+strings.Join(supportedModules(), ", ")+"]")

	_ = cmd.MarkFlagRequired("module")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
