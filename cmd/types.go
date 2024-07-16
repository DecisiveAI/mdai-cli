package cmd

import (
	"github.com/charmbracelet/lipgloss"
	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
)

const (
	Namespace      = "mdai"
	PatchOpAdd     = "add"
	PatchOpReplace = "replace"
	PatchOpRemove  = "remove"

	DatalyzerJSONPath      = "/spec/telemetryModule/collectors/0/measureVolumes"
	MutedPipelinesJSONPath = "/spec/telemetryModule/collectors/0/telemetryFiltering/filters/%v"
	OtelConfigJSONPath     = "/spec/telemetryModule/collectors/0/spec/config"
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

	MutedPipelineEmptyFilter = []byte(`[{ "op": "add", "path": "/spec/telemetryModule/collectors/0/telemetryFiltering", "value": { "filters": [] } }]`)
)

type mutePatch struct {
	Op    string                       `json:"op"`
	Path  string                       `json:"path"`
	Value mydecisivev1.TelemetryFilter `json:"value,omitempty"`
}

type datalyzerPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value bool   `json:"value"`
}

type mdaiOperatorOtelConfigPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}
