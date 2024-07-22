package operator

import mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"

type datalyzerPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value bool   `json:"value"`
}

type mutePatch struct {
	Op    string                       `json:"op"`
	Path  string                       `json:"path"`
	Value mydecisivev1.TelemetryFilter `json:"value,omitempty"`
}

type otelConfigPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}
