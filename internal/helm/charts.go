package helm

import (
	"embed"
	"fmt"
	"time"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*
var embedFS embed.FS

var chartSpecs = map[string]mdaitypes.ChartSpec{
	"cert-manager": {
		ReleaseName:     "cert-manager",
		ChartName:       "jetstack/cert-manager",
		Namespace:       "cert-manager",
		Version:         "1.15.0",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, // nolint: mnd
	},

	"opentelemetry-operator": {
		ReleaseName:     "opentelemetry-operator",
		ChartName:       "mydecisive/opentelemetry-operator",
		Namespace:       "mdai",
		Version:         "0.43.1",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, // nolint: mnd
	},

	"prometheus": {
		ReleaseName:     "prometheus",
		ChartName:       "prometheus-community/prometheus",
		Namespace:       "mdai",
		Version:         "25.21.0",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            false,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, // nolint: mnd
	},

	"metrics-server": {
		ReleaseName:     "metrics-server",
		ChartName:       "metrics-server",
		Namespace:       "kube-system",
		Version:         "3.12.1",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, // nolint: mnd
	},

	"mdai-api": {
		ReleaseName:     "mdai-api",
		ChartName:       "mydecisive/mdai-api",
		Namespace:       "mdai",
		Version:         "0.0.4",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            false,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, // nolint: mnd
	},

	"mdai-console": {
		ReleaseName:     "mdai-console",
		ChartName:       "mydecisive/mdai-console",
		Namespace:       "mdai",
		Version:         "0.2.1",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            false,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, // nolint: mnd
	},

	"datalyzer": {
		ReleaseName:     "datalyzer",
		ChartName:       "mydecisive/datalyzer",
		Namespace:       "mdai",
		Version:         "0.0.4",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            false,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, // nolint: mnd
	},

	"mdai-operator": {
		ReleaseName:     "mydecisive-engine-operator",
		ChartName:       "mydecisive/mydecisive-engine-operator",
		Namespace:       "mdai",
		Version:         "0.0.8",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, // nolint: mnd
	},

	"opentelemetry-demo": {
		ReleaseName:     "otel-demo",
		ChartName:       "opentelemetry/opentelemetry-demo",
		Namespace:       "mdai-otel-demo",
		Version:         "0.32.0",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         300 * time.Second, // nolint: mnd
	},
}

func getChartSpec(name string) (*mdaitypes.ChartSpec, error) {
	spec := chartSpecs[name]
	valuesYaml, _ := embedFS.ReadFile("templates/" + name + "-values.yaml")
	if err := yaml.Unmarshal(valuesYaml, &spec.Values); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chart spec %s: %w", name, err)
	}
	return &spec, nil
}
