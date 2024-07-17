package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/decisiveai/mdai-cli/internal/editor"
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

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "update [-f FILE] [--config CONFIG-TYPE] [--phase PHASE] [--block BLOCK]",
		Short:   "update a configuration",
		Long:    "update a configuration file or edit a configuration in an editor",
		Example: `	mdai update -f /path/to/mdai-operator.yaml  # update mdai-operator configuration from file
	mdai update --config=otel                   # edit otel collector configuration in $EDITOR
	mdai update --config=otel --phase=logs      # jump to logs block
	mdai update --config=otel --block=receivers # jump to receivers block`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			configP, _ := cmd.Flags().GetString("config")
			phaseP, _ := cmd.Flags().GetString("phase")
			blockP, _ := cmd.Flags().GetString("block")

			switch {
			case configP != "" && !slices.Contains(SupportedUpdateConfigTypes, configP):
				return fmt.Errorf("invalid config type: %s", configP)

			case phaseP != "" && !slices.Contains(SupportedPhases, phaseP):
				return fmt.Errorf("invalid phase: %s", phaseP)

			case blockP != "" && !slices.Contains(SupportedBlocks, blockP):
				return fmt.Errorf("invalid block: %s", blockP)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, _ []string) {
			fileP, _ := cmd.Flags().GetString("file")
			configP, _ := cmd.Flags().GetString("config")
			phaseP, _ := cmd.Flags().GetString("phase")
			blockP, _ := cmd.Flags().GetString("block")

			switch {
			case configP != "":
				var otelConfig string

				cfg := config.GetConfigOrDie()
				s := scheme.Scheme
				mydecisivev1.AddToScheme(s)
				k8sClient, _ := client.New(cfg, client.Options{Scheme: s})
				get := mydecisivev1.MyDecisiveEngine{}
				if err := k8sClient.Get(context.TODO(), client.ObjectKey{
					Namespace: Namespace,
					Name:      mdaitypes.MDAIOperatorName,
				}, &get); err != nil {
					fmt.Printf("error getting %s config: %v\n", configP, err)
					return
				}
				otelConfig = get.Spec.TelemetryModule.Collectors[0].Spec.Config
				f, err := os.CreateTemp("", "otelconfig")
				if err != nil {
					fmt.Printf("error saving %s config temp file: %+v\n", configP, err)
					return
				}
				f.WriteString(otelConfig)
				f.Close()

				defer os.Remove(f.Name())

				m := editor.NewModel(f.Name(), blockP, phaseP)
				if _, err := tea.NewProgram(m).Run(); err != nil {
					fmt.Printf("error running program: %v\n", err)
					os.Exit(1)
				}
				var applyConfig bool
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("apply config?").
							Value(&applyConfig).
							Affirmative("yes!").
							Negative("no."),
					),
				)
				form.Run()
				if !applyConfig {
					fmt.Println(configP + " configuration not updated")
					return
				}
				dynamicClient, _ := dynamic.NewForConfig(cfg)
				gvr := schema.GroupVersionResource{
					Group:    mdaitypes.MDAIOperatorGroup,
					Version:  mdaitypes.MDAIOperatorVersion,
					Resource: mdaitypes.MDAIOperatorResource,
				}
				otelConfigBytes, _ := os.ReadFile(f.Name())
				patchBytes, err := json.Marshal([]mdaiOperatorOtelConfigPatch{
					{
						Op:    PatchOpReplace,
						Path:  OtelConfigJSONPath,
						Value: string(otelConfigBytes),
					},
				})
				if err != nil {
					fmt.Printf("failed to marshal mdai operator patch: %v\n", err)
					return
				}

				if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					if _, err := dynamicClient.Resource(gvr).Namespace(Namespace).Patch(
						context.TODO(),
						mdaitypes.MDAIOperatorName,
						types.JSONPatchType,
						patchBytes,
						metav1.PatchOptions{},
					); err != nil {
						return fmt.Errorf("failed to apply patch: %w", err)
					}
					return nil
				}); err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(configP + " configuration updated")

			case fileP != "":
				cfg := config.GetConfigOrDie()
				dynamicClient, _ := dynamic.NewForConfig(cfg)
				gvr := schema.GroupVersionResource{
					Group:    mdaitypes.MDAIOperatorGroup,
					Version:  mdaitypes.MDAIOperatorVersion,
					Resource: mdaitypes.MDAIOperatorResource,
				}
				otelConfigBytes, _ := os.ReadFile(fileP)
				patchBytes, err := json.Marshal([]mdaiOperatorOtelConfigPatch{
					{
						Op:    PatchOpReplace,
						Path:  OtelConfigJSONPath,
						Value: string(otelConfigBytes),
					},
				})
				if err != nil {
					fmt.Printf("failed to marshal mdai operator patch: %v\n", err)
					return
				}

				if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					if _, err := dynamicClient.Resource(gvr).Namespace(Namespace).Patch(
						context.TODO(),
						mdaitypes.MDAIOperatorName,
						types.JSONPatchType,
						patchBytes,
						metav1.PatchOptions{},
					); err != nil {
						return fmt.Errorf("failed to apply patch: %w", err)
					}
					return nil
				}); err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(configP + " configuration updated")
			}
		},
	}
	cmd.Flags().StringP("file", "f", "", "file to update")
	cmd.Flags().StringP("config", "c", "", "config type to update ["+strings.Join(SupportedUpdateConfigTypes, ", ")+"]")
	cmd.Flags().String("block", "", "block to jump to ["+strings.Join(SupportedBlocks, ", ")+"]")
	cmd.Flags().String("phase", "", "phase to jump to ["+strings.Join(SupportedPhases, ", ")+"]")

	cmd.MarkFlagsMutuallyExclusive("file", "config")
	cmd.MarkFlagsOneRequired("file", "config")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
