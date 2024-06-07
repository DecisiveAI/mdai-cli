package cmd

import (
	"context"
	"embed"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kind"
	"github.com/decisiveai/mdai-cli/internal/processmanager"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/pytimer/k8sutil/apply"
	"github.com/spf13/cobra"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

//go:embed templates/*
var embedFS embed.FS

var installCommand = &cobra.Command{
	Use:   "install",
	Short: "install MyDecisive Engine",
	Long:  "install MyDecisive Engine",
	RunE: func(cmd *cobra.Command, args []string) error {
		helmCharts := []string{"cert-manager", "opentelemetry-operator", "prometheus", "mdai-api", "mdai-console", "datalyzer", "mdai-operator"}

		var installationType string
		s := huh.NewSelect[string]().
			Title("Installation Type").
			Options(
				huh.NewOption("Local Installation via kind", "kind"),
				huh.NewOption("AWS Installation via eks", "aws"),
			).
			Value(&installationType)

		huh.NewForm(huh.NewGroup(s)).Run()
		var kubeconfig string
		switch installationType {
		case "kind":
			_ = spinner.New().Title(" creating kubernetes cluster via kind 🔧").Type(spinner.Meter).Action(func() { kubeconfig = kind.Install() }).Run()
		}
		opt := &helmclient.KubeConfClientOptions{
			Options: &helmclient.Options{
				RepositoryCache:  os.TempDir() + "/.helmcache",
				RepositoryConfig: os.TempDir() + "/.helmrepo",
				Debug:            false,
				// DebugLog: func(format string, v ...interface{}) {
				// Change this to your own logger. Default is 'log.Printf(format, v...)'.
				// },
			},
			KubeContext: "",
			KubeConfig:  []byte(kubeconfig),
		}
		opt.Options.DebugLog = func(_ string, _ ...interface{}) {}

		helmClient, _ := helmclient.NewClientFromKubeConf(opt, helmclient.Timeout(time.Second*60))
		for _, chartRepo := range mdaihelm.ChartRepos {
			var err error
			_ = spinner.New().Title(" adding " + chartRepo.Name + " helm chart repo 🔧").Type(spinner.Meter).Action(
				func() {
					err = helmClient.AddOrUpdateChartRepo(chartRepo)
				},
			).Run()
			if err != nil {
				return err
			}
		}

		action := func(helmChart string) error {
			chartSpec := mdaihelm.GetChartSpec(helmChart)
			opt := &helmclient.KubeConfClientOptions{
				Options: &helmclient.Options{
					Namespace:        chartSpec.Namespace,
					RepositoryCache:  os.TempDir() + "/.helmcache",
					RepositoryConfig: os.TempDir() + "/.helmrepo",
					Debug:            false,
					// DebugLog: func(format string, v ...interface{}) {
					// Change this to your own logger. Default is 'log.Printf(format, v...)'.
					// },
				},
				KubeContext: "",
				KubeConfig:  []byte(kubeconfig),
			}
			opt.Options.DebugLog = func(_ string, _ ...interface{}) {}

			helmClient, _ := helmclient.NewClientFromKubeConf(opt, helmclient.Timeout(time.Second*60))
			_, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil)
			return err
		}
		if _, err := tea.NewProgram(processmanager.NewModel(helmCharts, action)).Run(); err != nil {
			tea.Println("Error running program:", err)
			os.Exit(1)
		}

		cfg := config.GetConfigOrDie()
		dynamicClient, _ := dynamic.NewForConfig(cfg)
		discoveryClient, _ := discovery.NewDiscoveryClientForConfig(cfg)
		applyYaml, _ := embedFS.ReadFile("templates/mdai-operator.yaml")
		applyOptions := apply.NewApplyOptions(dynamicClient, discoveryClient)
		return applyOptions.Apply(context.TODO(), applyYaml)
	},
}

func init() {
	rootCmd.AddCommand(installCommand)
}
