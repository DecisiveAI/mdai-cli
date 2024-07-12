package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kind"
	"github.com/decisiveai/mdai-cli/internal/viewport"
	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	GroupID: "installation",
	Use:     "demo [--cluster-name=CLUSTER-NAME] [--uninstall]",
	Short:   "install OpenTelemetry Demo",
	Long:    "install OpenTelemetry Demo",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var action func() error

		messages := make(chan string)
		debug := make(chan string)
		errs := make(chan error)
		done := make(chan bool)
		task := make(chan string)
		defer func() {
			close(messages)
			close(debug)
			close(errs)
			close(task)
			close(done)
		}()

		debugMode := false
		quietMode := false
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		uninstall, _ := cmd.Flags().GetBool("uninstall")

		helmcharts := []string{"opentelemetry-demo"}

		switch uninstall {
		case true:
			action = func() error {
				tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
				if err != nil {
					errs <- fmt.Errorf("failed to create temp dir: %w", err)
					return fmt.Errorf("failed to create temp dir: %w", err)
				}
				defer os.Remove(tmpfile.Name())
				helmclient := mdaihelm.NewClient(messages, debug, errs, tmpfile.Name())
				for _, helmchart := range helmcharts {
					task <- "uninstalling helm chart " + helmchart
					if err := helmclient.UninstallChart(helmchart); err != nil {
						errs <- fmt.Errorf("failed to uninstall helm chart %s: %w", helmchart, err)
						return fmt.Errorf("failed to uninstall helm chart %s: %w", helmchart, err)
					}
				}
				done <- true
				return nil
			}
		case false:
			action = func() error {
				task <- "creating kubernetes cluster via kind"
				kindclient := kind.NewClient(messages, debug, errs, clusterName)
				if _, err := kindclient.Install(); err != nil {
					errs <- fmt.Errorf("failed to create kubernetes cluster: %w", err)
					return fmt.Errorf("failed to create kubernetes cluster: %w", err)
				}

				tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
				if err != nil {
					errs <- fmt.Errorf("failed to create temp dir: %w", err)
					return fmt.Errorf("failed to create temp dir: %w", err)
				}
				defer os.Remove(tmpfile.Name())
				helmclient := mdaihelm.NewClient(messages, debug, errs, tmpfile.Name())
				task <- "adding helm repos"
				helmclient.AddRepos()
				for _, helmchart := range helmcharts {
					task <- "installing helm chart " + helmchart
					if err := helmclient.InstallChart(helmchart); err != nil {
						errs <- fmt.Errorf("failed to install helm chart %s %w", helmchart, err)
						return fmt.Errorf("failed to install helm chart %s: %w", helmchart, err)
					}
				}

				done <- true
				return nil
			}
		}
		go action()

		p := tea.NewProgram(
			viewport.InitialModel(
				messages,
				debug,
				errs,
				done,
				task,
				debugMode,
				quietMode,
			),
		)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run program: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)
	demoCmd.Flags().String("cluster-name", "mdai-local", "kubernetes cluster name")
	demoCmd.Flags().Bool("uninstall", false, "uninstall OpenTelemetry Demo")
}