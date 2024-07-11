package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var muteCmd = &cobra.Command{
	GroupID: "configuration",
	Use:     "mute -n|--name FILTER-NAME -p|--pipeline PIPELINE-NAME -d|--description DESCRIPTION",
	Short:   "mute a telemetry pipeline",
	Long:    `activate (add to pipeline configuration) a telemetry muting filter`,
	Example: `  mdai mute --name test-filter --description "test filter muting" --pipeline "logs"
  mdai mute --name another-filter --description "metrics pipeline muting" --pipeline "metrics"`,
	Run: func(cmd *cobra.Command, _ []string) {
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
			fmt.Println(err)
			return
		}

		cfg, _ := config.GetConfig()
		dynamicClient, _ := dynamic.NewForConfig(cfg)

		gvr := schema.GroupVersionResource{
			Group:    mdaitypes.MDAIOperatorGroup,
			Version:  mdaitypes.MDAIOperatorVersion,
			Resource: mdaitypes.MDAIOperatorResource,
		}

		s := scheme.Scheme
		mydecisivev1.AddToScheme(s)
		k8sClient, _ := client.New(cfg, client.Options{Scheme: s})
		get := mydecisivev1.MyDecisiveEngine{}
		k8sClient.Get(context.TODO(), client.ObjectKey{
			Namespace: Namespace,
			Name:      mdaitypes.MDAIOperatorName,
		}, &get)
		if get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering == nil {
			dynamicClient.Resource(gvr).Namespace(Namespace).Patch(
				context.TODO(),
				mdaitypes.MDAIOperatorName,
				types.JSONPatchType,
				MutedPipelineEmptyFilter,
				metav1.PatchOptions{},
			)
		}
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, err := dynamicClient.Resource(gvr).Namespace(Namespace).Patch(
				context.TODO(),
				mdaitypes.MDAIOperatorName,
				types.JSONPatchType,
				patchBytes,
				metav1.PatchOptions{},
			)
			return err
		}); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("pipeline(s) %v muted successfully as filter %s (%s).\n", pipelines, filterName, description)
	},
}

func init() {
	rootCmd.AddCommand(muteCmd)
	muteCmd.Flags().StringSliceP("pipeline", "p", []string{""}, "pipeline to mute")
	muteCmd.Flags().StringP("name", "n", "", "name of the filter")
	muteCmd.Flags().StringP("description", "d", "", "description of the filter")
	muteCmd.DisableFlagsInUseLine = true
}
