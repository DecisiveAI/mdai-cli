package cmd

import (
	"context"
	"fmt"

	"github.com/decisiveai/mdai-cli/internal/oteloperator"
	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	opentelemetry "github.com/decisiveai/opentelemetry-operator/apis/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfigOrDie()
		var group, version, kind string
		configType, _ := cmd.Flags().GetString("config")
		switch configType {
		case "otel":
			group = "opentelemetry.io"
			version = "v1alpha1"
			kind = "OpenTelemetryCollector"
		case "mdai":
			group = "mydecisive.ai"
			version = "v1"
			kind = "MyDecisiveEngine"
		}
		gvk := schema.GroupVersionKind{
			Group:   group,
			Version: version,
			Kind:    kind,
		}
		s := scheme.Scheme
		opentelemetry.AddToScheme(s)
		mydecisivev1.AddToScheme(s)
		k8sClient, _ := client.New(cfg, client.Options{Scheme: s})
		if configType == "mdai" {
			list := mydecisivev1.MyDecisiveEngineList{}
			list.SetGroupVersionKind(gvk)
			if err := k8sClient.List(context.TODO(), &list); err != nil {
				fmt.Printf("error: %+v\n", err)
			}
			fmt.Printf("name           : %+v\n", list.Items[0].Name)
			fmt.Printf("namespace      : %+v\n", list.Items[0].Namespace)
			fmt.Printf("measure volumes: %+v\n", list.Items[0].Spec.TelemetryModule.Collectors[0].MeasureVolumes)
			fmt.Printf("enabled        : %+v\n", list.Items[0].Spec.TelemetryModule.Collectors[0].Enabled)
			// fmt.Printf("config         : %+v\n", list.Items[0].Spec.TelemetryModule.Collectors[0].Spec.Config)

			get := mydecisivev1.MyDecisiveEngine{}
			get.SetGroupVersionKind(gvk)
			if err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "mydecisiveengine-sample-1"}, &get); err != nil {
				fmt.Printf("error: %+v\n", err)
			}
			fmt.Printf("%+v\n", get)
		} else if configType == "otel" {
			config := oteloperator.GetConfig()
			fmt.Println(config)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringP("config", "c", "", "config to get")
}
