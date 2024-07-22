package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			channels := mdaitypes.NewChannels()
			defer channels.Close()

			debugMode, _ := cmd.Flags().GetBool("debug")
			quietMode, _ := cmd.Flags().GetBool("quiet")
			// clusterName, _ := cmd.Flags().GetString("cluster-name")

			modes := mdaitypes.NewModes(debugMode, quietMode)

			go func() {
				tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
				if err != nil {
					channels.Error(fmt.Errorf("failed to create temp dir: %w", err))
					return
				}
				defer os.Remove(tmpfile.Name())
				helmclient := mdaihelm.NewClient(channels, tmpfile.Name())
				for _, helmchart := range mdaiHelmcharts {
					channels.Task("uninstalling helm chart " + helmchart)
					if err := helmclient.UninstallChart(helmchart); err != nil {
						channels.Error(fmt.Errorf("failed to uninstall helm chart %s: %w", helmchart, err))
						return
					}
				}
				channels.Message("helm charts uninstalled successfully.")

				helper, err := kubehelper.New()
				if err != nil {
					channels.Error(fmt.Errorf("failed to initialize kubehelper: %w", err))
					return
				}

				for _, crd := range crds {
					channels.Task("deleting crd " + crd)
					if err = helper.DeleteCRD(context.TODO(), crd); err != nil {
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
	cmd.Flags().String("cluster-name", "mdai-local", "kubernetes cluster name")
	cmd.Flags().Bool("debug", false, "debug mode")
	cmd.Flags().Bool("quiet", false, "quiet mode")

	return cmd
}
