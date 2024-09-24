package cmd

const (
	DisabledString = "✗"
	EnabledString  = "✓"
	NoDataString   = "--"
)

func supportedModules() []string {
	return []string{"datalyzer"}
}

func supportedGetConfigTypes() []string {
	return []string{"mdai", "otel"}
}

func supportedUpdateConfigTypes() []string {
	return []string{"otel"}
}

func supportedPhases() []string {
	return []string{"metrics", "logs", "traces"}
}

func supportedBlocks() []string {
	return []string{"receivers", "processors", "exporters"}
}

func customResourceDefinitions() []string {
	return []string{
		"mydecisiveengines.mydecisive.ai",

		"opentelemetrycollectors.opentelemetry.io",
		"instrumentations.opentelemetry.io",
		"opampbridges.opentelemetry.io",
	}
}

func pipelineFilterHeaders() []string {
	return []string{"NAME", "DESCRIPTION", "ENABLED", "MUTED PIPELINES"}
}

func filterServiceHeaders() []string {
	return []string{"NAME", "DESCRIPTION", "ENABLED", "FILTERED PIPELINES", "FILTERED TELEMETRY", "SERVICE PATTERN"}
}
