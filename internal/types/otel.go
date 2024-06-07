package types

type OtelConfig struct {
	Receivers  map[string]ReceiverConfig  `yaml:"receivers"`
	Exporters  map[string]ExporterConfig  `yaml:"exporters"`
	Processors map[string]ProcessorConfig `yaml:"processors"`
	Pipelines  map[string]PipelineConfig  `yaml:"service"`
}

type ReceiverConfig struct {
	Endpoint string `yaml:"endpoint"`
}

type ExporterConfig struct {
	Endpoint string `yaml:"endpoint"`
}

type ProcessorConfig struct {
	Attributes map[string]string `yaml:"attributes"`
}

type PipelineConfig struct {
	Receivers  []string `yaml:"receivers"`
	Processors []string `yaml:"processors"`
	Exporters  []string `yaml:"exporters"`
}
