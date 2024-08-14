package cmd

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	Namespace = "mdai"
)

var (
	purple  = lipgloss.NewStyle().Foreground(lipgloss.Color("#BF40BF"))
	white   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	lpurple = lipgloss.NewStyle().Foreground(lipgloss.Color("#800080"))
)

var (
	SupportedModules = []string{"datalyzer"}

	SupportedGetConfigTypes    = []string{"mdai", "otel"}
	SupportedUpdateConfigTypes = []string{"otel"}

	SupportedPhases = []string{"metrics", "logs", "traces"}
	SupportedBlocks = []string{"receivers", "processors", "exporters"}

	mdaiHelmcharts = []string{"cert-manager", "prometheus", "opentelemetry-operator", "mydecisive-engine-operator", "mdai-console", "datalyzer"}
	crds           = []string{
		"opentelemetrycollectors.opentelemetry.io",
		"instrumentations.opentelemetry.io",
		"opampbridges.opentelemetry.io",

		"certificaterequests.cert-manager.io",
		"certificates.cert-manager.io",
		"challenges.acme.cert-manager.io",
		"clusterissuers.cert-manager.io",
		"issuers.cert-manager.io",
		"orders.acme.cert-manager.io",
	}
)
