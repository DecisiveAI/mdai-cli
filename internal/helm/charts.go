package helm

import (
	"embed"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

//go:embed templates/*
var embedFS embed.FS

var chartSpecs = map[string]mdaitypes.ChartSpec{}

func init() {
	certManagerValuesYaml, _ := embedFS.ReadFile("templates/cert-manager-values.yaml")
	opentelemetryOperatorValuesYaml, _ := embedFS.ReadFile("templates/opentelemetry-operator-values.yaml")
	prometheusValuesYaml, _ := embedFS.ReadFile("templates/prometheus-values.yaml")
	mdaiConsoleValuesYaml, _ := embedFS.ReadFile("templates/mdai-console-values.yaml")
	mdaiOperatorValuesYaml, _ := embedFS.ReadFile("templates/mdai-operator-values.yaml")
	mdaiApiValuesYaml, _ := embedFS.ReadFile("templates/mdai-api-values.yaml")

	chartSpecs = make(map[string]mdaitypes.ChartSpec)

	chartSpecs["cert-manager"] = mdaitypes.ChartSpec{
		ReleaseName:     "cert-manager",
		ChartName:       "jetstack/cert-manager",
		Namespace:       "cert-manager",
		Version:         "1.13.1",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(certManagerValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["opentelemetry-operator"] = mdaitypes.ChartSpec{
		ReleaseName: "opentelemetry-operator",
		ChartName:   "mydecisive/opentelemetry-operator",
		// ChartName: "opentelemetry/opentelemetry-operator",
		Namespace: "mdai-otel-nucleus",
		Version:   "0.43.1",
		// Version:         "0.61.0",
		UpgradeCRDs:     true,
		Wait:            false,
		ValuesYaml:      string(opentelemetryOperatorValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["prometheus"] = mdaitypes.ChartSpec{
		ReleaseName:     "prometheus",
		ChartName:       "prometheus-community/prometheus",
		Namespace:       "mdai-otel-nucleus",
		Version:         "25.21.0",
		UpgradeCRDs:     true,
		Wait:            false,
		ValuesYaml:      string(prometheusValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["metrics-server"] = mdaitypes.ChartSpec{
		ReleaseName:     "metrics-server",
		ChartName:       "metrics-server",
		Namespace:       "kube-system",
		Version:         "3.12.1",
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["mdai-api"] = mdaitypes.ChartSpec{
		ReleaseName:     "mdai-api",
		ChartName:       "mydecisive/mdai-api",
		Namespace:       "mdai-otel-nucleus",
		Version:         "0.0.4",
		UpgradeCRDs:     true,
		Wait:            false,
		ValuesYaml:      string(mdaiApiValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["mdai-console"] = mdaitypes.ChartSpec{
		ReleaseName:     "mdai-console",
		ChartName:       "mydecisive/mdai-console",
		Namespace:       "mdai-otel-nucleus",
		Version:         "0.1.1",
		UpgradeCRDs:     true,
		Wait:            false,
		ValuesYaml:      string(mdaiConsoleValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["datalyzer"] = mdaitypes.ChartSpec{
		ReleaseName:     "datalyzer",
		ChartName:       "mydecisive/datalyzer",
		Namespace:       "mdai-otel-nucleus",
		Version:         "0.0.4",
		UpgradeCRDs:     true,
		Wait:            false,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["mdai-operator"] = mdaitypes.ChartSpec{
		ReleaseName:     "mydecisive-engine-operator",
		ChartName:       "mydecisive/mydecisive-engine-operator",
		Namespace:       "mdai-otel-nucleus",
		Version:         "0.0.3",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(mdaiOperatorValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}
}

func getChartSpec(name string) mdaitypes.ChartSpec {
	return chartSpecs[name]
}

func InstallChart(helmChart string) error {
	chartSpec := getChartSpec(helmChart)
	settings := cli.New()
	settings.SetNamespace(chartSpec.Namespace)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), chartSpec.Namespace, "", func(format string, v ...interface{}) { tea.Printf(format, v) }); err != nil {
		return err
	}

	client := action.NewInstall(actionConfig)
	client.ReleaseName = chartSpec.ReleaseName
	client.Namespace = chartSpec.Namespace
	client.CreateNamespace = chartSpec.CreateNamespace
	client.Wait = chartSpec.Wait
	client.Timeout = chartSpec.Timeout

	chartPath, err := client.ChartPathOptions.LocateChart(chartSpec.ChartName, settings)
	if err != nil {
		return err
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	values := map[string]interface{}{}
	if chartSpec.ValuesYaml != "" {
		if err := yaml.Unmarshal([]byte(chartSpec.ValuesYaml), &values); err != nil {
			return errors.Wrap(err, "failed to parse ValuesYaml")
		}
	}

	_, err = client.Run(chart, values)
	return err
}
