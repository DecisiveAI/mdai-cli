package helm

import (
	"embed"
	"time"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
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
	mdaiAPIValuesYaml, _ := embedFS.ReadFile("templates/mdai-api-values.yaml")
	opentelemetryDemoValuesYaml, _ := embedFS.ReadFile("templates/opentelemetry-demo-values.yaml")

	chartSpecs = make(map[string]mdaitypes.ChartSpec)

	chartSpecs["cert-manager"] = mdaitypes.ChartSpec{
		ReleaseName:     "cert-manager",
		ChartName:       "jetstack/cert-manager",
		Namespace:       "cert-manager",
		Version:         "1.15.0",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(certManagerValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, // nolint: gomnd
	}

	chartSpecs["opentelemetry-operator"] = mdaitypes.ChartSpec{
		ReleaseName: "opentelemetry-operator",
		ChartName:   "mydecisive/opentelemetry-operator",
		// ChartName: "opentelemetry/opentelemetry-operator",
		Namespace: "mdai",
		Version:   "0.43.1",
		// Version:         "0.61.0",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(opentelemetryOperatorValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["prometheus"] = mdaitypes.ChartSpec{
		ReleaseName:     "prometheus",
		ChartName:       "prometheus-community/prometheus",
		Namespace:       "mdai",
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
		Namespace:       "mdai",
		Version:         "0.0.4",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(mdaiAPIValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["mdai-console"] = mdaitypes.ChartSpec{
		ReleaseName:     "mdai-console",
		ChartName:       "mydecisive/mdai-console",
		Namespace:       "mdai",
		Version:         "0.1.1",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(mdaiConsoleValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["datalyzer"] = mdaitypes.ChartSpec{
		ReleaseName:     "datalyzer",
		ChartName:       "mydecisive/datalyzer",
		Namespace:       "mdai",
		Version:         "0.0.4",
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["mdai-operator"] = mdaitypes.ChartSpec{
		ReleaseName:     "mydecisive-engine-operator",
		ChartName:       "mydecisive/mydecisive-engine-operator",
		Namespace:       "mdai",
		Version:         "0.0.6",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(mdaiOperatorValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["opentelemetry-demo"] = mdaitypes.ChartSpec{
		ReleaseName:     "otel-demo",
		ChartName:       "opentelemetry/opentelemetry-demo",
		Namespace:       "mdai-otel-demo",
		Version:         "0.32.0",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(opentelemetryDemoValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         300 * time.Second, // nolint: gomnd
	}
}

func getChartSpec(name string) mdaitypes.ChartSpec {
	return chartSpecs[name]
}
