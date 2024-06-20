package cmd

import (
	"context"
	"embed"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kind"
	"github.com/decisiveai/mdai-cli/internal/processmanager"
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
	Short: "install MyDecisive Engine",
	Long:  "install MyDecisive Engine",
	PreRun: func(cmd *cobra.Command, args []string) {
		aws, _ := cmd.Flags().GetBool("aws")
		local, _ := cmd.Flags().GetBool("local")
		if aws {
			installationType = "aws"
		}
		if local {
			installationType = "kind"
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		helmCharts := []string{"cert-manager", "opentelemetry-operator", "prometheus", "mdai-api", "mdai-console", "datalyzer", "mdai-operator"}
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
		switch installationType {
		case "kind":
			if clusterName == "" {
				i := huh.NewInput().
					Prompt("cluster name: ").
					Placeholder("mdai-local").
					Value(&clusterName)
				huh.NewForm(huh.NewGroup(i)).Run()
			}
			_ = spinner.New().Title(" creating kubernetes cluster `" + clusterName + "` via kind 🔧").Type(spinner.Meter).Action(func() { kind.Install(clusterName) }).Run()
		}

		addReposFunc := func() error {
			return errors.Wrap(mdaihelm.AddRepos(), "failed to add repos")
		}

		action := func(helmChart string) error {
			return errors.Wrap(mdaihelm.InstallChart(helmChart), "failed to install "+helmChart)
		}

		mdaiOperatorManifestApply := func() error {
			cfg := config.GetConfigOrDie()

			dynamicClient, err := dynamic.NewForConfig(cfg)
			if err != nil {
				return err
			}
			discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
			if err != nil {
				return err
			}
			applyYaml, _ := embedFS.ReadFile("templates/mdai-operator.yaml")
			applyOptions := apply.NewApplyOptions(dynamicClient, discoveryClient)
			return applyOptions.Apply(context.TODO(), applyYaml)
		}

		if _, err := tea.NewProgram(processmanager.NewModel(helmCharts, action, mdaiOperatorManifestApply, addReposFunc)).Run(); err != nil {
			tea.Println("error running program: ", err)
			os.Exit(1)
		}
		return nil
	},
}

var demoCommand = &cobra.Command{
	Use:   "demo",
	Short: "install OpenTelemetry Demo",
	Long:  "install OpenTelemetry Demo",
	RunE: func(cmd *cobra.Command, args []string) error {
		helmCharts := []string{"opentelemetry-demo"}
		if clusterName == "" {
			i := huh.NewInput().
				Prompt("cluster name: ").
				Placeholder("mdai-local").
				Value(&clusterName)
			huh.NewForm(huh.NewGroup(i)).Run()
		}
		_ = spinner.New().Title(" creating kubernetes cluster `" + clusterName + "` via kind 🔧").Type(spinner.Meter).Action(func() { kind.Install(clusterName) }).Run()

		if err := mdaihelm.AddRepos(); err != nil {
			return errors.Wrap(err, "failed to add repos")
		}

		action := func(helmChart string) error {
			return errors.Wrap(mdaihelm.InstallChart(helmChart), "failed to install "+helmChart)
		}

		if _, err := tea.NewProgram(processmanager.NewModel(helmCharts, action, nil, nil)).Run(); err != nil {
			tea.Println("error running program: ", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCommand)
	rootCmd.AddCommand(demoCommand)
	installCommand.Flags().Bool("aws", false, "aws installation type")
	installCommand.Flags().Bool("local", false, "local installation type")
	installCommand.Flags().StringVar(&clusterName, "cluster-name", "", "kubernetes cluster name")
	demoCommand.Flags().StringVar(&clusterName, "cluster-name", "", "kubernetes cluster name")
}
