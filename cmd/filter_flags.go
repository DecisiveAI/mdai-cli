package cmd

import (
	"fmt"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/operator"
)

type filterAddFlags struct {
	name        string
	description string
	pipeline    []string
	service     string
	telemetry   []string
}

type filterListFlags struct {
	onlyService  bool
	onlyPipeline bool
}

type filterDisableFlags struct {
	filterName string
}

type filterEnableFlags struct {
	filterName string
}

type filterRemoveFlags struct {
	filterName string
}

func (flags filterAddFlags) toTelemetryFilterOptions() []operator.TelemetryFilterOption {
	funcs := []operator.TelemetryFilterOption{
		WithName(flags.name),
		WithDescription(flags.description),
	}
	if flags.service != "" {
		funcs = append(funcs, WithService(flags.service))
		if len(flags.pipeline) > 0 {
			funcs = append(funcs, WithServicePipeline(flags.pipeline))
		}
		if len(flags.telemetry) > 0 {
			funcs = append(funcs, WithTelemetry(flags.telemetry))
		}
	} else if len(flags.pipeline) > 0 {
		funcs = append(funcs, WithPipeline(flags.pipeline))
	}

	return funcs
}

func (flags filterAddFlags) successString() string {
	var sb strings.Builder
	if flags.service != "" {
		_, _ = fmt.Fprintf(&sb, `service pattern "%s" added successfully as filter "%s" (%s)`, flags.service, flags.name, flags.description)
		if len(flags.pipeline) > 0 {
			_, _ = fmt.Fprintf(&sb, " for pipelines %v\n", flags.pipeline)
		}
		if len(flags.telemetry) > 0 {
			_, _ = fmt.Fprintf(&sb, " for telemetry types %v\n", flags.telemetry)
		}
	} else {
		_, _ = fmt.Fprintf(&sb, `pipeline(s) %v added successfully as filter "%s" (%s).`, flags.pipeline, flags.name, flags.description)
		_, _ = fmt.Fprintln(&sb)
	}
	return sb.String()
}
