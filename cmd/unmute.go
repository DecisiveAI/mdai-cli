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

var unmuteCmd = &cobra.Command{
	GroupID: "configuration",
	Use:     "unmute -n|--name FILTER-NAME",
	Short:   "unmute a telemetry muting filter",
	Long:    ``,
	Example: `  mdai unmute --name test-filter # unmute the filter with name test-filter`,
	Run: func(cmd *cobra.Command, args []string) {
		var patchBytes []byte
		filterName, _ := cmd.Flags().GetString("name")

		cfg := config.GetConfigOrDie()
		s := scheme.Scheme
		mydecisivev1.AddToScheme(s)
		k8sClient, _ := client.New(cfg, client.Options{Scheme: s})
		get := mydecisivev1.MyDecisiveEngine{}
		if err := k8sClient.Get(context.TODO(), client.ObjectKey{
			Namespace: Namespace,
			Name:      mdaitypes.MDAIOperatorName,
		}, &get); err != nil {
			fmt.Printf("error: %+v\n", err)
			return
		}

		if get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering == nil {
			fmt.Printf("filter %s not found\n", filterName)
			return
		}

		for i, filter := range *get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters {
			if filter.Name == filterName {
				filter.Enabled = false
				patchBytes, _ = json.Marshal([]mutePatch{
					{
						Op:    PatchOpReplace,
						Path:  fmt.Sprintf(MutedPipelinesJSONPath, i),
						Value: filter,
					},
				})
				break
			}
		}
		if patchBytes == nil {
			fmt.Printf("filter %s not found.\n", filterName)
			return
		}

		dynamicClient, _ := dynamic.NewForConfig(cfg)
		gvr := schema.GroupVersionResource{
			Group:    mdaitypes.MDAIOperatorGroup,
			Version:  mdaitypes.MDAIOperatorVersion,
			Resource: mdaitypes.MDAIOperatorResource,
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
		fmt.Printf("%s filter unmuted successfully.\n", filterName)
	},
}

func init() {
	rootCmd.AddCommand(unmuteCmd)
	unmuteCmd.Flags().StringP("name", "n", "", "name of the filter")
	unmuteCmd.DisableFlagsInUseLine = true
}
