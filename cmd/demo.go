package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kind"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/decisiveai/mdai-cli/internal/viewport"
	"github.com/spf13/cobra"
)

func NewDemoCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "installation",
		Use:     "demo [--cluster-name=CLUSTER-NAME] [--uninstall]",
		Short:   "install OpenTelemetry Demo",
		Long:    "install OpenTelemetry Demo",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var action func()

			channels := mdaitypes.NewChannels()
			defer channels.Close()

			clusterName, _ := cmd.Flags().GetString("cluster-name")
			uninstall, _ := cmd.Flags().GetBool("uninstall")

			modes := mdaitypes.NewModes(false, false)

			helmcharts := []string{"opentelemetry-demo"}

			switch uninstall {
			case true:
				action = func() {
					tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
					if err != nil {
						channels.Error(fmt.Errorf("failed to create temp dir: %w", err))
						return
					}
					defer os.Remove(tmpfile.Name())
					helmclient := mdaihelm.NewClient(channels, tmpfile.Name())
					for _, helmchart := range helmcharts {
						channels.Task("uninstalling helm chart " + helmchart)
						if err := helmclient.UninstallChart(helmchart); err != nil {
							channels.Error(fmt.Errorf("failed to uninstall helm chart %s: %w", helmchart, err))
							return
						}
					}
					channels.Done()
				}
			case false:
				action = func() {
					channels.Task("creating kubernetes cluster via kind")
					kindclient := kind.NewClient(channels, clusterName)
					if _, err := kindclient.Install(); err != nil {
						channels.Error(fmt.Errorf("failed to create kubernetes cluster: %w", err))
						return
					}

					tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
					if err != nil {
						channels.Error(fmt.Errorf("failed to create temp dir: %w", err))
						return
					}
					defer os.Remove(tmpfile.Name())
					helmclient := mdaihelm.NewClient(channels, tmpfile.Name())
					channels.Task("adding helm repos")
					if err := helmclient.AddRepos(); err != nil {
						channels.Error(fmt.Errorf("failed to add helm repos: %w", err))
						return
					}
					for _, helmchart := range helmcharts {
						channels.Task("installing helm chart " + helmchart)
						if err := helmclient.InstallChart(helmchart); err != nil {
							channels.Error(fmt.Errorf("failed to install helm chart %s %w", helmchart, err))
							return
						}
					}
					channels.Done()
				}
			}

			go action()

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
	cmd.Flags().String("cluster-name", "mdai-local", "kubernetes cluster name")
	cmd.Flags().Bool("uninstall", false, "uninstall OpenTelemetry Demo")

	return cmd
}
