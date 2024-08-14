package cmd

import (
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/operator"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/decisiveai/mdai-cli/internal/viewport"
	"github.com/spf13/cobra"
)

//go:embed templates/*
var embedFS embed.FS

func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "installation",
		Use:     "install [--cluster-name CLUSTER-NAME] [--debug] [--quiet]",
		Short:   "install MyDecisive Cluster",
		Long:    "install MyDecisive Cluster",
		Example: `  mdai install --kubecontext kind-mdai-local # install on kind cluster mdai-local
  mdai install --debug                   # install in debug mode
  mdai install --quiet                   # install in quiet mode
  mdai install --confirm                 # install, with confirmation`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			confirm, _ := cmd.Flags().GetBool("confirm")

			if !confirm {
				kubeconfig := ctx.Value(mdaitypes.Kubeconfig{}).(string)
				kubecontext := ctx.Value(mdaitypes.Kubecontext{}).(string)
				if err := huh.NewConfirm().
					Title("Install MDAI into this cluster?").
					Description(fmt.Sprintf("kubeconfig: %s\nkubecontext: %s\n", kubeconfig, kubecontext)).
					Affirmative("Yes!").
					Negative("No.").
					Value(&confirm).Run(); err != nil {
					return fmt.Errorf("install failed: %w", err)
				}
			}
			if !confirm {
				return fmt.Errorf("aborting installation")
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
				channels.Task("adding helm repos")
				if err := helmclient.AddRepos(); err != nil {
					channels.Error(fmt.Errorf("failed to add helm repos: %w", err))
					return
				}
				for _, helmchart := range mdaiHelmcharts {
					channels.Task("installing helm chart " + helmchart)
					if err := helmclient.InstallChart(helmchart); err != nil {
						channels.Error(fmt.Errorf("failed to install helm chart %s: %w", helmchart, err))
						return
					}
				}

				manifest, _ := embedFS.ReadFile("templates/mdai-operator.yaml")
				if err := operator.Install(ctx, manifest); err != nil {
					channels.Error(fmt.Errorf("failed to apply mdai operator manifest: %w", err))
					return
				}

				channels.Message("installation completed successfully")
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
	cmd.Flags().Bool("confirm", false, "confirm installation")

	cmd.MarkFlagsMutuallyExclusive("debug", "quiet")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
