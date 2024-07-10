package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		module, _ := cmd.Flags().GetString("module")
		if module == "" {
			return fmt.Errorf("module is required")
		}
		if !slices.Contains(SupportedModules, module) {
			return fmt.Errorf("module %s is not supported for enabling", module)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.GetConfig()
		dynamicClient, _ := dynamic.NewForConfig(cfg)

		gvr := schema.GroupVersionResource{
			Group:    mdaitypes.MDAIOperatorGroup,
			Version:  mdaitypes.MDAIOperatorVersion,
			Resource: mdaitypes.MDAIOperatorResource,
		}

		patchBytes, _ := json.Marshal([]datalyzerPatch{
			{
				Op:    PatchOpReplace,
				Path:  DatalyzerJSONPath,
				Value: true,
			},
		})
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
	rootCmd.AddCommand(enableCmd)
	enableCmd.Flags().String("module", "", "module to enable ["+strings.Join(SupportedModules, ", ")+"]")
}
