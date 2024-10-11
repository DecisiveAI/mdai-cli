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
	"mdai-cluster": {
		ReleaseName:     "mdai-cluster",
		ChartURL:        "https://github.com/DecisiveAI/mdai-helm-charts/raw/gh-pages/%s-%s.tgz",
		Namespace:       "mdai",
		Version:         "0.0.3",
		Values:          map[string]any{},
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         120 * time.Second, //nolint: mnd
	},
}

func getChartSpec(name string) (*mdaitypes.ChartSpec, error) {
	spec, ok := chartSpecs[name]
	if !ok {
		return nil, fmt.Errorf("chart %s not found", name)
	}
	valuesYaml, _ := embedFS.ReadFile("templates/" + name + "-values.yaml")
	if err := yaml.Unmarshal(valuesYaml, &spec.Values); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chart spec %s: %w", name, err)
	}
	return &spec, nil
}
