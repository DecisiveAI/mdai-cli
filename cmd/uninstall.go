package cmd

import (
	"context"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/viewport"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var uninstallCmd = &cobra.Command{
	GroupID: "installation",
	Use:     "uninstall",
	Short:   "uninstall MyDecisive Cluster",
	Long:    "uninstall MyDecisive Cluster",
	RunE: func(cmd *cobra.Command, _ []string) error {
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

		debugMode, _ := cmd.Flags().GetBool("debug")
		quietMode, _ := cmd.Flags().GetBool("quiet")
		// clusterName, _ := cmd.Flags().GetString("cluster-name")

		go func() error {
			tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
			if err != nil {
				errs <- errors.Wrap(err, "failed to create temp dir")
				return errors.Wrap(err, "failed to create temp dir")
			}
			defer os.Remove(tmpfile.Name())
			helmclient := mdaihelm.NewClient(messages, debug, errs, tmpfile.Name())
			for _, helmchart := range mdaiHelmcharts {
				task <- "uninstalling helm chart " + helmchart
				if err := helmclient.UninstallChart(helmchart); err != nil {
					errs <- errors.Wrap(err, "failed to uninstall helm chart "+helmchart)
					return errors.Wrap(err, "failed to uninstall helm chart "+helmchart)
				}
			}
			messages <- "helm charts uninstalled successfully."

			cfg := config.GetConfigOrDie()
			apiExtensionsClientset, _ := apiextensionsclient.NewForConfig(cfg)
			crds := []string{
				"opentelemetrycollectors.opentelemetry.io", "instrumentations.opentelemetry.io", "opampbridges.opentelemetry.io",
				"certificaterequests.cert-manager.io", "certificates.cert-manager.io", "challenges.acme.cert-manager.io", "clusterissuers.cert-manager.io", "issuers.cert-manager.io", "orders.acme.cert-manager.io",
			}

			for _, crd := range crds {
				task <- "deleting crd " + crd
				if err = apiExtensionsClientset.ApiextensionsV1().CustomResourceDefinitions().Delete(
					context.TODO(),
					crd,
					metav1.DeleteOptions{},
				); err != nil {
					messages <- "CRD " + crd + " not found, skipping deletion."
				} else {
					messages <- "CRD " + crd + " deleted successfully."
				}
			}
			messages <- "CRDs deleted successfully."

			done <- true
			return nil
		}()

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
			return errors.Wrap(err, "failed to run program")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().String("cluster-name", "mdai-local", "kubernetes cluster name")
	uninstallCmd.Flags().Bool("debug", false, "debug mode")
	uninstallCmd.Flags().Bool("quiet", false, "quiet mode")
}
