package operator

import mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"

type telemetryFilter struct {
	remove bool
	filter mydecisivev1.TelemetryFilter
}

type TelemetryFilterOption func(*telemetryFilter)

func WithRemove() TelemetryFilterOption {
	return func(tf *telemetryFilter) {
		tf.remove = true
	}
}

func WithEnable() TelemetryFilterOption {
	return func(tf *telemetryFilter) {
		tf.filter.Enabled = true
	}
}

func WithDisable() TelemetryFilterOption {
	return func(tf *telemetryFilter) {
		tf.filter.Enabled = false
	}
}

func WithName(name string) TelemetryFilterOption {
	return func(tf *telemetryFilter) {
		tf.filter.Name = name
	}
}

func WithDescription(description string) TelemetryFilterOption {
	return func(tf *telemetryFilter) {
		tf.filter.Description = description
	}
}

func WithService(service string) TelemetryFilterOption {
	return func(tf *telemetryFilter) {
		if tf.filter.FilteredServices == nil {
			tf.filter.FilteredServices = &mydecisivev1.FilteredServices{}
		}
		tf.filter.FilteredServices.ServiceNamePattern = service
	}
}

func WithServicePipeline(pipeline []string) TelemetryFilterOption {
	return func(tf *telemetryFilter) {
		if tf.filter.FilteredServices == nil {
			tf.filter.FilteredServices = &mydecisivev1.FilteredServices{}
		}
		tf.filter.FilteredServices.Pipelines = &pipeline
	}
}

func WithPipeline(pipeline []string) TelemetryFilterOption {
	return func(tf *telemetryFilter) {
		tf.filter.MutedPipelines = &pipeline
	}
}

func WithTelemetry(telemetry []string) TelemetryFilterOption {
	return func(tf *telemetryFilter) {
		tf.filter.FilteredServices.TelemetryTypes = &telemetry
	}
}
