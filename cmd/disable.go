package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var disableCmd = &cobra.Command{
	GroupID: "configuration",
	Use:     "disable -m|--module MODULE",
	Short:   "disable a module",
	Long:    `disable a module`,
	Example: `  mdai disable --module datalyzer`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		module, _ := cmd.Flags().GetString("module")
		if module == "" {
			return fmt.Errorf("module is required")
		}
		if !slices.Contains(SupportedModules, module) {
			return fmt.Errorf("module %s is not supported for disabling", module)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var patchBytes []byte
		module, _ := cmd.Flags().GetString("module")
		cfg, _ := config.GetConfig()
		dynamicClient, _ := dynamic.NewForConfig(cfg)

		gvr := schema.GroupVersionResource{
			Group:    mdaitypes.MDAIOperatorGroup,
			Version:  mdaitypes.MDAIOperatorVersion,
			Resource: mdaitypes.MDAIOperatorResource,
		}
		switch module {
		case "datalyzer":
			patchBytes, _ = json.Marshal([]datalyzerPatch{
				{
					Op:    PatchOpReplace,
					Path:  DatalyzerJSONPath,
					Value: false,
				},
			})
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
		fmt.Printf("%s module disabled successfully.\n", module)
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
	disableCmd.Flags().String("module", "", "module to disable ["+strings.Join(SupportedModules, ", ")+"]")
	disableCmd.DisableFlagsInUseLine = true
}
