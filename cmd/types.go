package cmd

import mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"

const (
	Namespace      = "mdai"
	PatchOpAdd     = "add"
	PatchOpReplace = "replace"

	DatalyzerJSONPath      = "/spec/telemetryModule/collectors/0/measureVolumes"
	MutedPipelinesJSONPath = "/spec/telemetryModule/collectors/0/telemetryFiltering/filters/%v"
)

var (
	SupportedModules         = []string{"datalyzer"}
	SupportedConfigTypes     = []string{"mdai", "otel"}
	MutedPipelineEmptyFilter = []byte(`[{ "op": "add", "path": "/spec/telemetryModule/collectors/0/telemetryFiltering", "value": { "filters": [] } }]`)
)

type mutePatch struct {
	Op    string                       `json:"op"`
	Path  string                       `json:"path"`
	Value mydecisivev1.TelemetryFilter `json:"value"`
}

type datalyzerPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value bool   `json:"value"`
}