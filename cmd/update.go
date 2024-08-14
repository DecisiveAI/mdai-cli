package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/decisiveai/mdai-cli/internal/editor"
	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "update [-f FILE] [--config CONFIG-TYPE] [--phase PHASE] [--block BLOCK]",
		Short:   "update a configuration",
		Long:    "update a configuration file or edit a configuration in an editor",
		Example: `	mdai update -f /path/to/mdai-operator.yaml  # update mdai-operator configuration from file
	mdai update --config=otel                   # edit otel collector configuration in $EDITOR
	mdai update --config=otel --phase=logs      # jump to logs block
	mdai update --config=otel --block=receivers # jump to receivers block`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			configP, _ := cmd.Flags().GetString("config")
			phaseP, _ := cmd.Flags().GetString("phase")
			blockP, _ := cmd.Flags().GetString("block")

			switch {
			case configP != "" && !slices.Contains(SupportedUpdateConfigTypes, configP):
				return fmt.Errorf("invalid config type: %s", configP)

			case phaseP != "" && !slices.Contains(SupportedPhases, phaseP):
				return fmt.Errorf("invalid phase: %s", phaseP)

			case blockP != "" && !slices.Contains(SupportedBlocks, blockP):
				return fmt.Errorf("invalid block: %s", blockP)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			fileP, _ := cmd.Flags().GetString("file")
			configP, _ := cmd.Flags().GetString("config")
			phaseP, _ := cmd.Flags().GetString("phase")
			blockP, _ := cmd.Flags().GetString("block")

			switch {
			case configP != "":
				var otelConfig string

				get, err := operator.GetOperator(ctx)
				if err != nil {
					return err
				}
				otelConfig = get.Spec.TelemetryModule.Collectors[0].Spec.Config
				f, err := os.CreateTemp("", "otelconfig")
				if err != nil {
					return fmt.Errorf("error creating %s config temp file: %w", configP, err)
				}
				if _, err := f.WriteString(otelConfig); err != nil {
					return fmt.Errorf("error saving %s config temp file: %w", configP, err)
				}
				if err := f.Close(); err != nil {
					return fmt.Errorf("error closing %s config temp file: %w", configP, err)
				}

				defer os.Remove(f.Name())

				m := editor.NewModel(f.Name(), blockP, phaseP)
				if _, err := tea.NewProgram(m).Run(); err != nil {
					return err
				}

				var applyConfig bool
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("apply config?").
							Value(&applyConfig).
							Affirmative("yes!").
							Negative("no."),
					),
				)
				if err := form.Run(); err != nil {
					return err
				}
				if !applyConfig {
					fmt.Println(configP + " configuration not updated")
					return nil
				}

				otelConfigBytes, _ := os.ReadFile(f.Name())
				if err := operator.UpdateOTELConfig(ctx, string(otelConfigBytes)); err != nil {
					return fmt.Errorf("error updating otel collector configuration: %w", err)
				}
				fmt.Println(configP + " configuration updated")

			case fileP != "":
				otelConfigBytes, err := os.ReadFile(fileP)
				if err != nil {
					return fmt.Errorf(`error reading file "%s": %w`, fileP, err)
				}
				if err := operator.UpdateOTELConfig(ctx, string(otelConfigBytes)); err != nil {
					return fmt.Errorf("error updating otel collector configuration: %w", err)
				}
				fmt.Println(configP + " configuration updated")
			}

			return nil
		},
	}
	cmd.Flags().StringP("file", "f", "", "file to update")
	cmd.Flags().StringP("config", "c", "", "config type to update ["+strings.Join(SupportedUpdateConfigTypes, ", ")+"]")
	cmd.Flags().String("block", "", "block to jump to ["+strings.Join(SupportedBlocks, ", ")+"]")
	cmd.Flags().String("phase", "", "phase to jump to ["+strings.Join(SupportedPhases, ", ")+"]")

	cmd.MarkFlagsMutuallyExclusive("file", "config")
	cmd.MarkFlagsOneRequired("file", "config")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
