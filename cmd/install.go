package cmd

import (
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kind"
	"github.com/decisiveai/mdai-cli/internal/operator"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/decisiveai/mdai-cli/internal/viewport"
	"github.com/spf13/cobra"
)

//go:embed templates/*
var embedFS embed.FS

var installationType string

func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "installation",
		Use:     "install [--cluster-name CLUSTER-NAME] [--debug] [--quiet]",
		Short:   "install MyDecisive Cluster",
		Long:    "install MyDecisive Cluster",
		Example: `  mdai install --kubecontext kind-mdai-local # install on kind cluster mdai-local
  mdai install --debug                   # install in debug mode
  mdai install --quiet                   # install in quiet mode`,
		Args: cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			/*
				aws, _ := cmd.Flags().GetBool("aws")
				local, _ := cmd.Flags().GetBool("local")
				if aws {
					installationType = "aws"
				}
				if local {
					installationType = "kind"
				}
			*/
			installationType = "kind"

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			channels := mdaitypes.NewChannels()
			defer channels.Close()

			debugMode, _ := cmd.Flags().GetBool("debug")
			quietMode, _ := cmd.Flags().GetBool("quiet")
			clusterName, _ := cmd.Flags().GetString("cluster-name")

			modes := mdaitypes.NewModes(debugMode, quietMode)
			/*
				if installationType == "" {
					s := huh.NewSelect[string]().
						Title("Installation Type").
						Options(
							huh.NewOption("Local Installation via kind", "kind"),
							huh.NewOption("AWS Installation via eks", "aws"),
						).
						Value(&installationType)

					huh.NewForm(huh.NewGroup(s)).Run()
				}
			*/

			go func() {
				switch installationType {
				case "kind":
					channels.Task("creating kubernetes cluster via kind")
					kindclient := kind.NewClient(channels, clusterName)
					if _, err := kindclient.Create(); err != nil {
						channels.Error(fmt.Errorf("failed to create kubernetes cluster: %w", err))
						return
					}
				}

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
	// cmd.Flags().Bool("aws", false, "aws installation type")
	// cmd.Flags().Bool("local", false, "local installation type")
	cmd.Flags().String("cluster-name", "mdai-local", "kubernetes cluster name")
	cmd.Flags().Bool("debug", false, "debug mode")
	cmd.Flags().Bool("quiet", false, "quiet mode")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
