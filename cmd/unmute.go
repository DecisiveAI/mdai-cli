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
	Long:    `deactivate (delete from pipeline configuration) a telemetry muting filter`,
	Example: `  mdai unmute --name test-filter # unmute the filter with name test-filter`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		var (
			patchBytes []byte
			err        error
		)
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
			return fmt.Errorf("failed to get mdai operator: %w", err)
		}

		if get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering == nil {
			return fmt.Errorf("filter %s not found", filterName)
		}

		for i, filter := range *get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters {
			if filter.Name == filterName {
				filter.Enabled = false
				patchBytes, err = json.Marshal([]mutePatch{
					{
						Op:    PatchOpReplace,
						Path:  fmt.Sprintf(MutedPipelinesJSONPath, i),
						Value: filter,
					},
				})
				if err != nil {
					return fmt.Errorf("failed to marshal patch: %w", err)
				}
				break
			}
		}
		if patchBytes == nil {
			return fmt.Errorf("filter %s not found", filterName)
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
			return fmt.Errorf("failed to apply patch: %w", err)
		}); err != nil {
			return err // nolint: wrapcheck
		}
		fmt.Printf("%s filter unmuted successfully.\n", filterName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(unmuteCmd)
	unmuteCmd.Flags().StringP("name", "n", "", "name of the filter")
	unmuteCmd.DisableFlagsInUseLine = true
	unmuteCmd.SilenceUsage = true
}
