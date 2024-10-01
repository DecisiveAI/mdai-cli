package cmd

import (
	"fmt"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewDisableCommand() *cobra.Command {
	flags := disableFlags{}
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "disable -m|--module MODULE",
		Short:   "disable a module",
		Long:    `disable a module`,
		Example: `  mdai disable --module datalyzer`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			switch flags.module {
			case "datalyzer":
				if err := operator.DisableDatalyzer(ctx); err != nil {
					return err
				}
			default:
				return fmt.Errorf(`module "%s" is not supported for disabling`, flags.module)
			}

			fmt.Printf("%s module disabled successfully.\n", flags.module)
			return nil
		},
	}
	cmd.Flags().StringVar(&flags.module, "module", "", "module to disable ["+strings.Join(supportedModules(), ", ")+"]")

	_ = cmd.MarkFlagRequired("module")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
