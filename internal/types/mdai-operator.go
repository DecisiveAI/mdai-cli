package types

import (
	"gopkg.in/yaml.v3"
)

type MDAIOperator struct {
	APIVersion string               `yaml:"apiVersion"`
	Kind       string               `yaml:"kind"`
	Metadata   MDAIOperatorMetadata `yaml:"metadata"`
	Spec       MDAIOperatorSpec     `yaml:"spec"`
}

type MDAIOperatorSpec struct {
	TelemetryModule MDAIOperatorTelemetryModule `yaml:"telemetryModule"`
}

type MDAIOperatorLabels struct {
	AppKubernetesIoName      string `yaml:"app.kubernetes.io/name"`
	AppKubernetesIoInstance  string `yaml:"app.kubernetes.io/instance"`
	AppKubernetesIoPartOf    string `yaml:"app.kubernetes.io/part-of"`
	AppKubernetesIoManagedBy string `yaml:"app.kubernetes.io/managed-by"`
	AppKubernetesIoCreatedBy string `yaml:"app.kubernetes.io/created-by"`
}

type MDAIOperatorTelemetryModule struct {
	Attributes MDAIOperatorAttributes  `yaml:"attributes"`
	Collectors []MDAIOperatorCollector `yaml:"collectors"`
}

type MDAIOperatorAttributes struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type MDAIOperatorMetadata struct {
	Labels MDAIOperatorLabels `yaml:"labels"`
	Name   string             `yaml:"name"`
}

type MDAIOperatorCollectorSpec struct {
	Replicas int `yaml:"replicas"`
	Ports    []struct {
		Name     string `yaml:"name"`
		Port     int    `yaml:"port"`
		Protocol string `yaml:"protocol,omitempty"`
	} `yaml:"ports"`
	Config string `yaml:"config"`
}

type MDAIOperatorCollector struct {
	Name           string                    `yaml:"name"`
	Enabled        bool                      `yaml:"enabled"`
	MeasureVolumes bool                      `yaml:"measureVolumes"`
	Spec           MDAIOperatorCollectorSpec `yaml:"spec"`
}

const (
	MDAIOperatorName     = "mydecisiveengine-sample-1"
	MDAIOperatorGroup    = "mydecisive.ai"
	MDAIOperatorVersion  = "v1"
	MDAIOperatorResource = "mydecisiveengines"
	MDAIOperatorKind     = "MyDecisiveEngine"
)

func NewMDAIOperator() MDAIOperator {
	m := MDAIOperator{
		APIVersion: "mydecisive.ai/v1",
		Kind:       "MyDecisiveEngine",
		Metadata: MDAIOperatorMetadata{
			Labels: MDAIOperatorLabels{
				AppKubernetesIoName:      "mydecisiveengine",
				AppKubernetesIoInstance:  "mydecisiveengine-sample",
				AppKubernetesIoPartOf:    "mydecisive-engine-operator",
				AppKubernetesIoManagedBy: "kustomize",
				AppKubernetesIoCreatedBy: "mydecisive-engine-operator",
			},
			Name: "mydecisiveengine-sample-1",
		},
		Spec: MDAIOperatorSpec{
			TelemetryModule: MDAIOperatorTelemetryModule{
				Attributes: MDAIOperatorAttributes{
					Name:    "telemetry",
					Version: "0.0.1",
				},
				Collectors: []MDAIOperatorCollector{
					{
						Name:           "test-collector",
						Enabled:        true,
						MeasureVolumes: true,
						Spec: MDAIOperatorCollectorSpec{
							Replicas: 2,
						},
					},
				},
			},
		},
	}
	return m
}

func (m *MDAIOperator) ToYaml() ([]byte, error) {
	return yaml.Marshal(m) // nolint: wrapcheck
}

func (m *MDAIOperator) SetCollectorConfig(config string) {
	m.Spec.TelemetryModule.Collectors[0].Spec.Config = config
}
