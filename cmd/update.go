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
	flags := updateFlags{}
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "update [-f FILE] [--config CONFIG-TYPE] [--phase PHASE] [--block BLOCK]",
		Short:   "update a configuration",
		Long:    "update a configuration file or edit a configuration in an editor",
		Example: `	mdai update -f /path/to/mdai-operator.yaml  # update mdai-operator configuration from file
	mdai update --config=otel                   # edit otel collector configuration in $EDITOR
	mdai update --config=otel --phase=logs      # jump to logs block
	mdai update --config=otel --block=receivers # jump to receivers block`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			switch {
			case flags.config != "" && !slices.Contains(supportedUpdateConfigTypes(), flags.config):
				return fmt.Errorf("invalid config type: %s", flags.config)

			case flags.phase != "" && !slices.Contains(supportedPhases(), flags.phase):
				return fmt.Errorf("invalid phase: %s", flags.phase)

			case flags.block != "" && !slices.Contains(supportedBlocks(), flags.block):
				return fmt.Errorf("invalid block: %s", flags.block)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			switch {
			case flags.config != "":
				var otelConfig string

				get, err := operator.GetOperator(ctx)
				if err != nil {
					return err
				}
				otelConfig = get.Spec.TelemetryModule.Collectors[0].Spec.Config
				f, err := os.CreateTemp("", "otelconfig")
				if err != nil {
					return fmt.Errorf("error creating %s config temp file: %w", flags.config, err)
				}
				if _, err := f.WriteString(otelConfig); err != nil {
					return fmt.Errorf("error saving %s config temp file: %w", flags.config, err)
				}
				if err := f.Close(); err != nil {
					return fmt.Errorf("error closing %s config temp file: %w", flags.config, err)
				}

				defer func() {
					_ = os.Remove(f.Name())
				}()

				m := editor.NewModel(f.Name(), flags.block, flags.phase)
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
					fmt.Println(flags.config + " configuration not updated")
					return nil
				}

				otelConfigBytes, _ := os.ReadFile(f.Name())
				if err := operator.UpdateOTELConfig(ctx, string(otelConfigBytes)); err != nil {
					return fmt.Errorf("error updating otel collector configuration: %w", err)
				}
				fmt.Println(flags.config + " configuration updated")

			case flags.file != "":
				otelConfigBytes, err := os.ReadFile(flags.file)
				if err != nil {
					return fmt.Errorf(`error reading file "%s": %w`, flags.file, err)
				}
				if err := operator.UpdateOTELConfig(ctx, string(otelConfigBytes)); err != nil {
					return fmt.Errorf("error updating otel collector configuration: %w", err)
				}
				fmt.Println(flags.config + " configuration updated")
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "file to update")
	cmd.Flags().StringVarP(&flags.config, "config", "c", "", "config type to update ["+strings.Join(supportedUpdateConfigTypes(), ", ")+"]")
	cmd.Flags().StringVar(&flags.block, "block", "", "block to jump to ["+strings.Join(supportedBlocks(), ", ")+"]")
	cmd.Flags().StringVar(&flags.phase, "phase", "", "phase to jump to ["+strings.Join(supportedPhases(), ", ")+"]")

	cmd.MarkFlagsMutuallyExclusive("file", "config")
	cmd.MarkFlagsOneRequired("file", "config")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
