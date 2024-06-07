package cmd

import (
	"context"
	"fmt"

	"github.com/pytimer/k8sutil/apply"
	"github.com/spf13/cobra"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		module, _ := cmd.Flags().GetString("module")
		cfg := config.GetConfigOrDie()
		dynamicClient, _ := dynamic.NewForConfig(cfg)
		discoveryClient, _ := discovery.NewDiscoveryClientForConfig(cfg)
		applyOptions := apply.NewApplyOptions(dynamicClient, discoveryClient)
		/*patchBytes, err := embedFS.ReadFile("templates/mdai-operator.yaml")
		if err != nil {
			fmt.Printf("failed to read file: %v", err)
		}*/
		switch module {
		case "datalyzer":
			patchBytes, _ := embedFS.ReadFile("templates/mdai-operator-patch-disable-datalyzer.yaml")

			if err := applyOptions.WithServerSide(true).Apply(context.TODO(), patchBytes); err != nil {
				fmt.Printf("apply error: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
	disableCmd.Flags().String("module", "", "module to disable")
}
