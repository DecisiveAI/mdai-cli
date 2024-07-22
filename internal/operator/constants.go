package operator

const (
	PatchOpAdd             = "add"
	PatchOpRemove          = "remove"
	PatchOpReplace         = "replace"
	DatalyzerJSONPath      = "/spec/telemetryModule/collectors/0/measureVolumes"
	MutedPipelinesJSONPath = "/spec/telemetryModule/collectors/0/telemetryFiltering/filters/%v"
	OtelConfigJSONPath     = "/spec/telemetryModule/collectors/0/spec/config"
)

var MutedPipelineEmptyFilter = []byte(`[{ "op": "add", "path": "/spec/telemetryModule/collectors/0/telemetryFiltering", "value": { "filters": [] } }]`)
