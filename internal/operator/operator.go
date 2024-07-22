package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/kubehelper"
	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	opentelemetry "github.com/decisiveai/opentelemetry-operator/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func EnableDatalyzer() error {
	return setMeasureVolumes(true)
}

func DisableDatalyzer() error {
	return setMeasureVolumes(false)
}

func GetOperator() (*mydecisivev1.MyDecisiveEngine, error) {
	helper, err := kubehelper.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubehelper: %w", err)
	}
	return helper.GetOperator(context.TODO())
}

func GetOTELOperator() (*opentelemetry.OpenTelemetryCollector, error) {
	helper, err := kubehelper.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubehelper: %w", err)
	}
	return helper.GetOTELOperator(context.TODO())
}

func Mute(name string, description string, pipelines []string) error {
	helper, err := kubehelper.New()
	if err != nil {
		return fmt.Errorf("failed to initialize kubehelper: %w", err)
	}

	tf := mydecisivev1.TelemetryFilter{
		Enabled:        true,
		Name:           name,
		Description:    description,
		MutedPipelines: &pipelines,
	}

	patchBytes, err := json.Marshal(
		[]mutePatch{
			{
				Op:    PatchOpAdd,
				Path:  fmt.Sprintf(MutedPipelinesJSONPath, "-"),
				Value: tf,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	telemetryFiltering, err := helper.GetTelemetryFiltering(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get telemetry filtering: %w", err)
	}
	if telemetryFiltering == nil {
		if err := helper.Patch(context.TODO(), types.JSONPatchType, MutedPipelineEmptyFilter); err != nil {
			return fmt.Errorf("failed to patch telemetry filtering: %w", err)
		}
	}

	if err := helper.Patch(context.TODO(), types.JSONPatchType, patchBytes); err != nil {
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

func Unmute(name string, remove bool) error {
	var (
		patchBytes []byte
		err        error
	)

	helper, err := kubehelper.New()
	if err != nil {
		return fmt.Errorf("failed to initialize api: %w", err)
	}

	telemetryFiltering, err := helper.GetTelemetryFiltering(context.TODO())
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

	if err := helper.Patch(context.TODO(), types.JSONPatchType, patchBytes); err != nil {
		return fmt.Errorf("failed to patch telemetry filtering: %w", err)
	}
	return nil
}

func UpdateOTELConfig(config string) error {
	helper, err := kubehelper.New()
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

	if err := helper.Patch(context.TODO(), types.JSONPatchType, patchBytes); err != nil {
		return fmt.Errorf("failed to apply otel collector config: %w", err)
	}

	return nil
}

func Install(manifest []byte) error {
	helper, err := kubehelper.New()
	if err != nil {
		return fmt.Errorf("failed to initialize kubehelper: %w", err)
	}
	return helper.Apply(context.TODO(),
		manifest,
		&schema.GroupVersionKind{
			Group:   mydecisivev1.GroupVersion.Group,
			Version: mydecisivev1.GroupVersion.Version,
			Kind:    "MyDecisiveEngine",
		},
	)
}

func setMeasureVolumes(v bool) error {
	helper, err := kubehelper.New()
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

	if err := helper.Patch(context.TODO(), types.JSONPatchType, patchBytes); err != nil {
		return fmt.Errorf("failed to %s datalyzer: %w", action, err)
	}

	return nil
}
