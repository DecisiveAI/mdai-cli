package operator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/decisiveai/mdai-cli/internal/kubehelper"
	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	"k8s.io/apimachinery/pkg/types"
)

var WithContext = kubehelper.WithContext

func EnableDatalyzer(ctx context.Context) error {
	return setMeasureVolumes(ctx, true)
}

func DisableDatalyzer(ctx context.Context) error {
	return setMeasureVolumes(ctx, false)
}

func GetOperator(ctx context.Context) (*mydecisivev1.MyDecisiveEngine, error) {
	helper, err := kubehelper.New(WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubehelper: %w", err)
	}
	return helper.GetOperator(ctx)
}

func CreateTelemetryFilter(ctx context.Context, options ...TelemetryFilterOption) error {
	newTelemetryFilter := new(telemetryFilter)
	options = append(options, WithEnable())
	for _, option := range options {
		option(newTelemetryFilter)
	}

	helper, err := kubehelper.New(WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to initialize kubehelper: %w", err)
	}

	patch := []mutePatch{
		{
			Op:    PatchOpAdd,
			Path:  fmt.Sprintf(MutedPipelinesJSONPath, "-"),
			Value: newTelemetryFilter.filter,
		},
	}

	telemetryFiltering, err := helper.GetTelemetryFiltering(ctx)
	if err != nil {
		return fmt.Errorf("failed to get telemetry filtering: %w", err)
	}
	if telemetryFiltering == nil {
		if err := helper.Patch(ctx, types.JSONPatchType, MutedPipelineEmptyFilter); err != nil {
			return fmt.Errorf("failed to patch telemetry filtering: %w", err)
		}
		telemetryFiltering, err = helper.GetTelemetryFiltering(ctx)
		if err != nil {
			return fmt.Errorf("failed to get telemetry filtering: %w", err)
		}
	}

	for i, filter := range *telemetryFiltering.Filters {
		if filter.Name == newTelemetryFilter.filter.Name {
			patch = []mutePatch{
				{
					Op:    PatchOpReplace,
					Path:  fmt.Sprintf(MutedPipelinesJSONPath, i),
					Value: newTelemetryFilter.filter,
				},
			}
			break
		}
	}

	var patchBytes []byte
	if patchBytes, err = json.Marshal(patch); err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	if err := helper.Patch(ctx, types.JSONPatchType, patchBytes); err != nil {
		return fmt.Errorf("failed to patch telemetry filtering: %w", err)
	}
	return nil
}

func RemoveTelemetryFilter(ctx context.Context, options ...TelemetryFilterOption) error {
	options = append(options, WithRemove())
	return toggleTelemetryFilter(ctx, options...)
}

func EnableTelemetryFilter(ctx context.Context, options ...TelemetryFilterOption) error {
	options = append(options, WithEnable())
	return toggleTelemetryFilter(ctx, options...)
}

func DisableTelemetryFilter(ctx context.Context, options ...TelemetryFilterOption) error {
	options = append(options, WithDisable())
	return toggleTelemetryFilter(ctx, options...)
}

func UpdateOTELConfig(ctx context.Context, config string) error {
	helper, err := kubehelper.New(WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to initialize kubehelper: %w", err)
	}

	patchBytes, err := json.Marshal(
		[]otelConfigPatch{
			{
				Op:    PatchOpAdd,
				Path:  OtelConfigJSONPath,
				Value: config,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	if err := helper.Patch(ctx, types.JSONPatchType, patchBytes); err != nil {
		return fmt.Errorf("failed to apply otel collector config: %w", err)
	}

	return nil
}

func Install(ctx context.Context, manifest []byte) error {
	helper, err := kubehelper.New(WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to initialize kubehelper: %w", err)
	}
	gvk := mydecisivev1.GroupVersion.WithKind("MyDecisiveEngine")
	return helper.Apply(ctx,
		manifest,
		&gvk,
	)
}

func setMeasureVolumes(ctx context.Context, v bool) error {
	helper, err := kubehelper.New(WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to initialize kubehelper: %w", err)
	}

	patchBytes, err := json.Marshal(
		[]datalyzerPatch{
			{
				Op:    PatchOpReplace,
				Path:  DatalyzerJSONPath,
				Value: v,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to marshal datalyzer patch: %w", err)
	}

	action := "disable"
	if v {
		action = "enable"
	}

	if err := helper.Patch(ctx, types.JSONPatchType, patchBytes); err != nil {
		return fmt.Errorf("failed to %s datalyzer: %w", action, err)
	}

	return nil
}

func toggleTelemetryFilter(ctx context.Context, options ...TelemetryFilterOption) error {
	var (
		patchBytes []byte
		err        error
	)

	tf := new(telemetryFilter)
	for _, option := range options {
		option(tf)
	}

	helper, err := kubehelper.New(WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to initialize api: %w", err)
	}

	telemetryFiltering, err := helper.GetTelemetryFiltering(ctx)
	if err != nil {
		return fmt.Errorf("failed to get telemetry filtering: %w", err)
	}
	if telemetryFiltering == nil {
		return fmt.Errorf("filter %s not found", tf.filter.Name)
	}

	for i, filter := range *telemetryFiltering.Filters {
		if filter.Name != tf.filter.Name {
			continue
		}
		filter.Enabled = tf.filter.Enabled
		var patch []mutePatch

		if tf.remove {
			patch = []mutePatch{
				{
					Op:   PatchOpRemove,
					Path: fmt.Sprintf(MutedPipelinesJSONPath, i),
				},
			}
		} else {
			patch = []mutePatch{
				{
					Op:    PatchOpReplace,
					Path:  fmt.Sprintf(MutedPipelinesJSONPath, i),
					Value: filter,
				},
			}
		}
		patchBytes, err = json.Marshal(patch)
		if err != nil {
			return fmt.Errorf("failed to marshal patch: %w", err)
		}
		break
	}
	if patchBytes == nil {
		return fmt.Errorf(`filter "%s" not found`, tf.filter.Name)
	}

	if err := helper.Patch(ctx, types.JSONPatchType, patchBytes); err != nil {
		return fmt.Errorf("failed to patch telemetry filtering: %w", err)
	}
	return nil
}
