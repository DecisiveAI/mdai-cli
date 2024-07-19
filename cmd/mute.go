package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func NewMuteCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "mute -n|--name FILTER-NAME -p|--pipeline PIPELINE-NAME -d|--description DESCRIPTION",
		Short:   "mute a telemetry pipeline",
		Long:    `activate (add to pipeline configuration) a telemetry muting filter`,
		Example: `  mdai mute --name test-filter --description "test filter muting" --pipeline "logs"
  mdai mute --name another-filter --description "metrics pipeline muting" --pipeline "metrics"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			filterName, _ := cmd.Flags().GetString("name")
			pipelines, _ := cmd.Flags().GetStringSlice("pipeline")
			description, _ := cmd.Flags().GetString("description")

			patchBytes, err := json.Marshal([]mutePatch{
				{
					Op:   PatchOpAdd,
					Path: fmt.Sprintf(MutedPipelinesJSONPath, "-"),
					Value: mydecisivev1.TelemetryFilter{
						Name:           filterName,
						Description:    description,
						Enabled:        true,
						MutedPipelines: &pipelines,
					},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to marshal patch: %w", err)
			}

			cfg, _ := config.GetConfig()
			dynamicClient, _ := dynamic.NewForConfig(cfg)

			s := scheme.Scheme
			mydecisivev1.AddToScheme(s)
			k8sClient, _ := client.New(cfg, client.Options{Scheme: s})
			get := mydecisivev1.MyDecisiveEngine{}
			if err := k8sClient.Get(context.TODO(), client.ObjectKey{
				Namespace: Namespace,
				Name:      mdaitypes.MDAIOperatorName,
			}, &get); err != nil {
				return fmt.Errorf("failed to get mdai operator: %w", err)
			}
			if get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering == nil {
				if _, err := dynamicClient.Resource(gvr).Namespace(Namespace).Patch(
					context.TODO(),
					mdaitypes.MDAIOperatorName,
					types.JSONPatchType,
					MutedPipelineEmptyFilter,
					metav1.PatchOptions{},
				); err != nil {
					return fmt.Errorf("failed to apply patch: %w", err)
				}
			}
			if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				if _, err := dynamicClient.Resource(gvr).Namespace(Namespace).Patch(
					context.TODO(),
					mdaitypes.MDAIOperatorName,
					types.JSONPatchType,
					patchBytes,
					metav1.PatchOptions{},
				); err != nil {
					return err // nolint: wrapcheck
				}
				return nil
			}); err != nil {
				/*
					pretty up the error message from mdai operator
					* first, check if the error is about a unique filter name
					* then, check if the error is about a pipeline not found
					* then, check if the error is about a pipeline already muted in several filters
					* finally, return the original error message
				*/
				if strings.Contains(err.Error(), fmt.Sprintf("Filter name %s is not unique", filterName)) {
					return fmt.Errorf(`filter name "%s" already exists in config`, filterName)
				}
				for _, pipeline := range pipelines {
					switch {
					case strings.Contains(err.Error(), fmt.Sprintf("pipeline %s not found in config", pipeline)):
						return fmt.Errorf(`pipeline "%s" not found in config`, pipeline)
					case strings.Contains(err.Error(), fmt.Sprintf("Pipeline %s is muted in several filters", pipeline)):
						return fmt.Errorf(`pipeline "%s" is muted in another filter`, pipeline)
					}
				}
				return fmt.Errorf("failed to apply patch: %w", err)
			}
			fmt.Printf("pipeline(s) %v muted successfully as filter %s (%s).\n", pipelines, filterName, description)
			return nil
		},
	}
	cmd.Flags().StringSliceP("pipeline", "p", []string{""}, "pipeline to mute")
	cmd.Flags().StringP("name", "n", "", "name of the filter")
	cmd.Flags().StringP("description", "d", "", "description of the filter")

	cmd.MarkFlagsRequiredTogether("name", "description", "pipeline")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("description")
	_ = cmd.MarkFlagRequired("pipeline")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
