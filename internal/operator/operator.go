package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/kubehelper"
	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	"k8s.io/apimachinery/pkg/types"
)

func EnableDatalyzer(ctx context.Context) error {
	return setMeasureVolumes(ctx, true)
}

func DisableDatalyzer(ctx context.Context) error {
	return setMeasureVolumes(ctx, false)
}

func GetOperator(ctx context.Context) (*mydecisivev1.MyDecisiveEngine, error) {
	helper, err := kubehelper.New(kubehelper.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubehelper: %w", err)
	}
	return helper.GetOperator(ctx)
}

func Mute(ctx context.Context, name string, description string, pipelines []string) error {
	var patchBytes []byte
	helper, err := kubehelper.New(kubehelper.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to initialize kubehelper: %w", err)
	}

	tf := mydecisivev1.TelemetryFilter{
		Enabled:        true,
		Name:           name,
		Description:    description,
		MutedPipelines: &pipelines,
	}

	patch := []mutePatch{
		{
			Op:    PatchOpAdd,
			Path:  fmt.Sprintf(MutedPipelinesJSONPath, "-"),
			Value: tf,
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
		if filter.Name == name {
			patch = []mutePatch{
				{
					Op:    PatchOpReplace,
					Path:  fmt.Sprintf(MutedPipelinesJSONPath, i),
					Value: tf,
				},
			}
			break
		}
	}

	if patchBytes, err = json.Marshal(patch); err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	if err := helper.Patch(ctx, types.JSONPatchType, patchBytes); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("Filter name %s is not unique", tf.Name)) {
			return fmt.Errorf(`filter name "%s" already exists in config`, tf.Name)
		}
		for _, pipeline := range *tf.MutedPipelines {
			switch {
			case strings.Contains(err.Error(), fmt.Sprintf("pipeline %s not found in config", pipeline)):
				return fmt.Errorf(`pipeline "%s" not found in config`, pipeline)
			case strings.Contains(err.Error(), fmt.Sprintf("Pipeline %s is muted in several filters", pipeline)):
				return fmt.Errorf(`pipeline "%s" is muted in another filter`, pipeline)
			}
		}
		return fmt.Errorf("failed to patch telemetry filtering: %w", err)
	}
	return nil
}

func Unmute(ctx context.Context, name string, remove bool) error {
	var (
		patchBytes []byte
		err        error
	)

	helper, err := kubehelper.New(kubehelper.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to initialize api: %w", err)
	}

	telemetryFiltering, err := helper.GetTelemetryFiltering(ctx)
	if err != nil {
		return fmt.Errorf("failed to get telemetry filtering: %w", err)
	}
	if telemetryFiltering == nil {
		return fmt.Errorf("filter %s not found", name)
	}

	for i, filter := range *telemetryFiltering.Filters {
		if filter.Name == name {
			filter.Enabled = false
			if remove {
				patchBytes, err = json.Marshal(
					[]mutePatch{
						{
							Op:   PatchOpRemove,
							Path: fmt.Sprintf(MutedPipelinesJSONPath, i),
						},
					})
			} else {
				patchBytes, err = json.Marshal(
					[]mutePatch{
						{
							Op:    PatchOpReplace,
							Path:  fmt.Sprintf(MutedPipelinesJSONPath, i),
							Value: filter,
						},
					})
			}
			if err != nil {
				return fmt.Errorf("failed to marshal patch: %w", err)
			}
			break
		}
	}
	if patchBytes == nil {
		return fmt.Errorf("filter %s not found", name)
	}

	if err := helper.Patch(ctx, types.JSONPatchType, patchBytes); err != nil {
		return fmt.Errorf("failed to patch telemetry filtering: %w", err)
	}
	return nil
}

func UpdateOTELConfig(ctx context.Context, config string) error {
	helper, err := kubehelper.New(kubehelper.WithContext(ctx))
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
	helper, err := kubehelper.New(kubehelper.WithContext(ctx))
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
	helper, err := kubehelper.New(kubehelper.WithContext(ctx))
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
