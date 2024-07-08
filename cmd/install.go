package cmd

import (
	"context"
	"embed"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kind"
	"github.com/decisiveai/mdai-cli/internal/viewport"
	"github.com/pkg/errors"
	"github.com/pytimer/k8sutil/apply"
	"github.com/spf13/cobra"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

//go:embed templates/*
var embedFS embed.FS

var (
	installationType string
	clusterName      string
)

var installCommand = &cobra.Command{
	Use:   "install",
	Short: "install MyDecisive Nucleus",
	Long:  "install MyDecisive Nucleus",
	PreRun: func(cmd *cobra.Command, args []string) {
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
	},
	RunE: func(cmd *cobra.Command, args []string) error {
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

		helmcharts := []string{"cert-manager", "opentelemetry-operator", "prometheus", "mdai-api", "mdai-console", "datalyzer", "mdai-operator"}
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

		go func() error {
			switch installationType {
			case "kind":
				task <- "creating kubernetes cluster via kind"
				kindclient := kind.NewClient(messages, debug, errs, clusterName)
				if _, err := kindclient.Install(); err != nil {
					errs <- errors.Wrap(err, "failed to create kubernetes cluster")
					return errors.Wrap(err, "failed to create kubernetes cluster")
				}
			}

			tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
			if err != nil {
				errs <- errors.Wrap(err, "failed to create temp dir")
				return errors.Wrap(err, "failed to create temp dir")
			}
			defer os.Remove(tmpfile.Name())
			helmclient := mdaihelm.NewClient(messages, debug, errs, tmpfile.Name())
			task <- "adding helm repos"
			if err := helmclient.AddRepos(); err != nil {
				errs <- errors.Wrap(err, "failed to add helm repos")
				return errors.Wrap(err, "failed to add helm repos")
			}
			for _, helmchart := range helmcharts {
				task <- "installing helm chart " + helmchart
				if err := helmclient.InstallChart(helmchart); err != nil {
					errs <- errors.Wrap(err, "failed to install helm chart "+helmchart)
					return errors.Wrap(err, "failed to install helm chart "+helmchart)
				}
			}

			cfg, err := config.GetConfig()
			if err != nil {
				errs <- errors.Wrap(err, "failed to get kubernetes config")
				return errors.Wrap(err, "failed to get kubernetes config")
			}

			dynamicClient, err := dynamic.NewForConfig(cfg)
			if err != nil {
				errs <- errors.Wrap(err, "failed to create dynamic client")
				return errors.Wrap(err, "failed to create dynamic client")
			}
			discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
			if err != nil {
				errs <- errors.Wrap(err, "failed to create discovery client")
				return errors.Wrap(err, "failed to create discovery client")
			}

			applyOptions := apply.NewApplyOptions(dynamicClient, discoveryClient)
			applyYaml, _ := embedFS.ReadFile("templates/mdai-operator.yaml")
			task <- "applying mdai-operator manifest"
			if err := applyOptions.Apply(context.TODO(), applyYaml); err != nil {
				errs <- errors.Wrap(err, "failed to apply mdai-operator manifest")
				return errors.Wrap(err, "failed to apply mdai-operator manifest")
			}

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
			return errors.Wrap(err, "failed to run tea program")
		}

		return nil
	},
}

var demoCommand = &cobra.Command{
	Use:   "demo",
	Short: "install OpenTelemetry Demo",
	Long:  "install OpenTelemetry Demo",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		helmcharts := []string{"opentelemetry-demo"}

		go func() error {
			task <- "creating kubernetes cluster via kind"
			kindclient := kind.NewClient(messages, debug, errs, clusterName)
			if _, err := kindclient.Install(); err != nil {
				errs <- errors.Wrap(err, "failed to create kubernetes cluster")
				return errors.Wrap(err, "failed to create kubernetes cluster")
			}

			tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
			if err != nil {
				errs <- errors.Wrap(err, "failed to create temp dir")
				return errors.Wrap(err, "failed to create temp dir")
			}
			defer os.Remove(tmpfile.Name())
			helmclient := mdaihelm.NewClient(messages, debug, errs, tmpfile.Name())
			task <- "adding helm repos"
			helmclient.AddRepos()
			for _, helmchart := range helmcharts {
				task <- "installing helm chart " + helmchart
				if err := helmclient.InstallChart(helmchart); err != nil {
					errs <- errors.Wrap(err, "failed to install helm chart "+helmchart)
					return errors.Wrap(err, "failed to install helm chart "+helmchart)
				}
			}

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
			return errors.Wrap(err, "failed to run tea program")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCommand)
	rootCmd.AddCommand(demoCommand)
	//installCommand.Flags().Bool("aws", false, "aws installation type")
	//installCommand.Flags().Bool("local", false, "local installation type")
	installCommand.Flags().StringVar(&clusterName, "cluster-name", "mdai-local", "kubernetes cluster name")
	installCommand.Flags().Bool("debug", false, "debug mode")
	installCommand.Flags().Bool("quiet", false, "quiet mode")
	demoCommand.Flags().StringVar(&clusterName, "cluster-name", "mdai-local", "kubernetes cluster name")
}
