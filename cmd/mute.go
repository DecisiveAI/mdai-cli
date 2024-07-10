package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var muteCmd = &cobra.Command{
	Use:   "mute",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		filterName, _ := cmd.Flags().GetString("name")
		pipeline, _ := cmd.Flags().GetString("pipeline")
		description, _ := cmd.Flags().GetString("description")

		patchBytes, _ := json.Marshal([]mutePatch{
			{
				Op:   PatchOpAdd,
				Path: fmt.Sprintf(MutedPipelinesJSONPath, "-"),
				Value: mydecisivev1.TelemetryFilter{
					Name:           filterName,
					Description:    description,
					Enabled:        true,
					MutedPipelines: &[]string{pipeline},
				},
			},
		})

		cfg, _ := config.GetConfig()
		dynamicClient, _ := dynamic.NewForConfig(cfg)

		gvr := schema.GroupVersionResource{
			Group:    mdaitypes.MDAIOperatorGroup,
			Version:  mdaitypes.MDAIOperatorVersion,
			Resource: mdaitypes.MDAIOperatorResource,
		}

		//patchBytes := []byte(`[{ "op": "replace", "path": "/spec/telemetryModule/collectors/0/measureVolumes", "value": true }]`)
		//patchBytes := []byte(`[{ "op": "add", "path": "/spec/telemetryModule/collectors/0/telemetryFiltering/filters/-", "value": [{ "name": "test-filter", "enabled": true, "description": "Hello this is filter", "mutedPipelines": ["logs/foobar"] }] }]`)
		//patchBytes = []byte(`[{ "op": "add", "path": "/spec/telemetryModule/collectors/0/telemetryFiltering/filters/-", "value": { "enabled": true, "name": "test-filter", "description": "Hello this is filter", "mutedPipelines": ["logs/foobar"] } }]`)
		//patchBytes := []byte(`[{ "op": "add", "path": "/spec/telemetryModule/collectors/0/telemetryFiltering/filters/-", "value": { "name": "another-filter", "enabled": true, "description": "WHOA two filters?!", "mutedPipelines": ["metrics"] } }]`)

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
		fmt.Println("patched successfully")
	},
}

func init() {
	rootCmd.AddCommand(muteCmd)
	muteCmd.Flags().StringP("pipeline", "p", "", "pipeline to mute")
	muteCmd.Flags().StringP("name", "n", "", "name of the filter")
	muteCmd.Flags().StringP("description", "d", "", "description of the filter")
}
