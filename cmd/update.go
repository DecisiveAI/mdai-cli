package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/decisiveai/mdai-cli/internal/editor"
	"github.com/decisiveai/mdai-cli/internal/oteloperator"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/pytimer/k8sutil/apply"
	"github.com/spf13/cobra"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// //go:embed templates/*.yaml
// var embedFS embed.FS

var (
	validPhases = []string{"metrics", "logs", "traces"}
	validBlocks = []string{"receivers", "processors", "exporters"}
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "",
	Long:  "",
	Example: `	-f /path/to/mdai-operator.yaml
	--config=otel
	--config=otel --phase=logs
	--config=otel --block=receivers`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		fileP, _ := cmd.Flags().GetString("file")
		configP, _ := cmd.Flags().GetString("config")
		phaseP, _ := cmd.Flags().GetString("phase")
		blockP, _ := cmd.Flags().GetString("block")

		if fileP != "" && configP != "" {
			return errors.New("cannot specify both --file and --config")
		}

		if !slices.Contains(validPhases, phaseP) {
			return fmt.Errorf("invalid phase: %s", phaseP)
		}

		if !slices.Contains(validBlocks, blockP) {
			return fmt.Errorf("invalid block: %s", blockP)
		}

		/*if phaseP != "" {
			for _, v := range validPhases {
				if v == phaseP {
					continue
				}
				return fmt.Errorf("invalid phase: %s", phaseP)
			}
		}
		if blockP != "" {
			for _, v := range validBlocks {
				if v == blockP {
					continue
				}
				return fmt.Errorf("invalid block: %s", blockP)
			}
		}*/

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fileP, _ := cmd.Flags().GetString("file")
		configP, _ := cmd.Flags().GetString("config")
		phaseP, _ := cmd.Flags().GetString("phase")
		blockP, _ := cmd.Flags().GetString("block")

		if configP != "" {
			var otelConfig string
			var f *os.File
			_ = spinner.New().Title(" fetching current collector configuration 🔧").Type(spinner.Meter).Action(
				func() {
					otelConfig = oteloperator.GetConfig()
					f, _ = os.CreateTemp("", "otelconfig")
					f.WriteString(otelConfig)
					f.Close()
				}).Run()
			defer os.Remove(f.Name())

			m := editor.NewModel(f.Name(), blockP, phaseP)
			if _, err := tea.NewProgram(m).Run(); err != nil {
				fmt.Println("error running program: ", err)
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
			if applyConfig {
				_ = spinner.New().Title(" updating current collector configuration 🔧").Type(spinner.Meter).Action(
					func() {
						cfg := config.GetConfigOrDie()
						dynamicClient, _ := dynamic.NewForConfig(cfg)
						discoveryClient, _ := discovery.NewDiscoveryClientForConfig(cfg)
						otelConfigBytes, _ := os.ReadFile(f.Name())
						mdaiOperator := mdaitypes.NewMDAIOperator()
						mdaiOperator.SetCollectorConfig(string(otelConfigBytes))
						applyYaml, _ := mdaiOperator.ToYaml()
						applyOptions := apply.NewApplyOptions(dynamicClient, discoveryClient)
						if err := applyOptions.Apply(context.TODO(), applyYaml); err != nil {
							panic(fmt.Sprintf("apply error: %v", err))
						}
					}).Run()
			} else {
				fmt.Println("oh well")
			}
		}

		if fileP != "" {
			action := func() {
				cfg := config.GetConfigOrDie()
				dynamicClient, _ := dynamic.NewForConfig(cfg)
				discoveryClient, _ := discovery.NewDiscoveryClientForConfig(cfg)
				applyYaml, _ := os.ReadFile(fileP)
				applyOptions := apply.NewApplyOptions(dynamicClient, discoveryClient)
				if err := applyOptions.Apply(context.TODO(), applyYaml); err != nil {
					panic(fmt.Sprintf("apply error: %v", err))
				}
			}
			_ = spinner.New().Title(" updating current collector configuration 🔧").Action(action).Type(spinner.Meter).Run()
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("file", "f", "", "file to update")
	updateCmd.Flags().StringP("config", "c", "", "config type to update")
	updateCmd.Flags().String("block", "", "block to jump to ["+strings.Join(validBlocks, ", ")+"]")
	updateCmd.Flags().String("phase", "", "phase to jump to ["+strings.Join(validPhases, ", ")+"]")
	updateCmd.Flags().SortFlags = true
}
