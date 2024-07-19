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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func NewEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "enable -m|--module MODULE",
		Short:   "enable a module",
		Long:    `enable one of the supported modules`,
		Example: `  mdai enable --module datalyzer`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			module, _ := cmd.Flags().GetString("module")
			if module != "" && !slices.Contains(SupportedModules, module) {
				return fmt.Errorf(`module "%s" is not supported for enabling`, module)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			var (
				patchBytes []byte
				err        error
			)
			module, _ := cmd.Flags().GetString("module")
			cfg, _ := config.GetConfig()
			dynamicClient, _ := dynamic.NewForConfig(cfg)

			switch module {
			case "datalyzer":
				patchBytes, err = json.Marshal([]datalyzerPatch{
					{
						Op:    PatchOpReplace,
						Path:  DatalyzerJSONPath,
						Value: true,
					},
				})
				if err != nil {
					return fmt.Errorf("failed to marshal datalyzer patch: %w", err)
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
				return fmt.Errorf("failed to apply patch: %w", err)
			}
			fmt.Printf("%s module enabled successfully.\n", module)
			return nil
		},
	}
	cmd.Flags().String("module", "", "module to enable ["+strings.Join(SupportedModules, ", ")+"]")

	_ = cmd.MarkFlagRequired("module")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
