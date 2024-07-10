package cmd

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/oteloperator"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "",
	Long:  "",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		configType, _ := cmd.Flags().GetString("config")
		if configType == "" {
			return fmt.Errorf("module is required")
		}
		if !slices.Contains(SupportedConfigTypes, configType) {
			return fmt.Errorf("config type %s is not supported", configType)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		configType, _ := cmd.Flags().GetString("config")
		cfg := config.GetConfigOrDie()
		s := scheme.Scheme
		mydecisivev1.AddToScheme(s)
		k8sClient, _ := client.New(cfg, client.Options{Scheme: s})
		switch configType {
		case "mdai":
			get := mydecisivev1.MyDecisiveEngine{}
			if err := k8sClient.Get(context.TODO(), client.ObjectKey{
				Namespace: Namespace,
				Name:      mdaitypes.MDAIOperatorName,
			}, &get); err != nil {
				fmt.Printf("error: %+v\n", err)
				return
			}
			fmt.Printf("name           : %s\n", get.Name)
			fmt.Printf("namespace      : %s\n", get.Namespace)
			fmt.Printf("measure volumes: %v\n", get.Spec.TelemetryModule.Collectors[0].MeasureVolumes)
			fmt.Printf("enabled        : %v\n", get.Spec.TelemetryModule.Collectors[0].Enabled)
			if get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering != nil {
				fmt.Println("filters")
				for _, filter := range *get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters {
					fmt.Printf("\tname       : %s\n", filter.Name)
					fmt.Printf("\tdescription: %s\n", filter.Description)
					fmt.Printf("\tenabled    : %v\n", filter.Enabled)
					fmt.Printf("\tpipelines  : %s\n", strings.Join(*filter.MutedPipelines, ", "))
				}
			}
		case "otel":
			config := oteloperator.GetConfig()
			fmt.Println(config)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringP("config", "c", "", "configuration to get ["+strings.Join(SupportedConfigTypes, ", ")+"]")
}
