package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kubehelper"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/decisiveai/mdai-cli/internal/viewport"
	"github.com/spf13/cobra"
)

func NewUninstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "installation",
		Use:     "uninstall",
		Short:   "uninstall MyDecisive Cluster",
		Long:    "uninstall MyDecisive Cluster",
		Example: `  mdai uninstall --kubecontext kind-mdai-local # uninstall from kind cluster mdai-local
  mdai uninstall --debug                   # uninstall in debug mode
  mdai uninstall --quiet                   # uninstall in quiet mode
  mdai uninstall --confirm                 # uninstall, with confirmation`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			confirm, _ := cmd.Flags().GetBool("confirm")

			if !confirm {
				kubeconfig := ctx.Value(mdaitypes.Kubeconfig{}).(string)
				kubecontext := ctx.Value(mdaitypes.Kubecontext{}).(string)
				if err := huh.NewConfirm().
					Title("Uninstall MDAI from this cluster?").
					Description(fmt.Sprintf("kubeconfig: %s\nkubecontext: %s\n", kubeconfig, kubecontext)).
					Negative("No!").
					Affirmative("Yes.").
					Value(&confirm).Run(); err != nil {
					return fmt.Errorf("uninstall failed: %w", err)
				}
			}
			if !confirm {
				return fmt.Errorf("aborting uninstallation")
			}
			channels := mdaitypes.NewChannels()
			defer channels.Close()

			debugMode, _ := cmd.Flags().GetBool("debug")
			quietMode, _ := cmd.Flags().GetBool("quiet")

			modes := mdaitypes.NewModes(debugMode, quietMode)

			go func() {
				tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
				if err != nil {
					channels.Error(fmt.Errorf("failed to create temp dir: %w", err))
					return
				}
				defer func() {
					if err := os.Remove(tmpfile.Name()); err != nil {
						channels.Error(fmt.Errorf("failed to remove temp file: %w", err))
					}
				}()
				helmclient := mdaihelm.NewClient(
					mdaihelm.WithContext(ctx),
					mdaihelm.WithChannels(channels),
					mdaihelm.WithRepositoryConfig(tmpfile.Name()),
				)
				for _, helmchart := range mdaiHelmcharts {
					channels.Task("uninstalling helm chart " + helmchart)
					if err := helmclient.UninstallChart(helmchart); err != nil {
						channels.Error(fmt.Errorf("failed to uninstall helm chart %s: %w", helmchart, err))
						return
					}
				}
				channels.Message("helm charts uninstalled successfully.")

				helper, err := kubehelper.New(kubehelper.WithContext(ctx))
				if err != nil {
					channels.Error(fmt.Errorf("failed to initialize kubehelper: %w", err))
					return
				}

				for _, crd := range crds {
					channels.Task("deleting crd " + crd)
					if err = helper.DeleteCRD(ctx, crd); err != nil {
						channels.Message("CRD " + crd + " not found, skipping deletion.")
						continue
					}
					channels.Message("CRD " + crd + " deleted successfully.")
				}
				channels.Message("CRDs deleted successfully.")
				channels.Done()
			}()

			p := tea.NewProgram(
				viewport.InitialModel(
					channels,
					modes,
				),
			)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("failed to run program: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().Bool("debug", false, "debug mode")
	cmd.Flags().Bool("quiet", false, "quiet mode")
	cmd.Flags().Bool("confirm", false, "confirm uninstallation")

	cmd.MarkFlagsMutuallyExclusive("debug", "quiet")

	return cmd
}
