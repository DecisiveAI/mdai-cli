package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/kind/pkg/cluster"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		configMapList := &corev1.ConfigMapList{}
		cfg := config.GetConfigOrDie()
		k8sClient, _ := client.New(cfg, client.Options{})
		selectorLabels := map[string]string{
			"app.kubernetes.io/managed-by": "opentelemetry-operator",
			"app.kubernetes.io/instance":   "default.test-collector", //  naming.Truncate("%s.%s", 63, instance.Namespace, instance.Name),
			"app.kubernetes.io/part-of":    "opentelemetry",
			"app.kubernetes.io/component":  "opentelemetry-collector", // component,
		}
		listOps := &client.ListOptions{
			Namespace:     "default",
			LabelSelector: labels.SelectorFromSet(selectorLabels),
			// LabelSelector: labels.SelectorFromSet(manifestutils.SelectorLabels(params.OtelCol.ObjectMeta, collector.ComponentOpenTelemetryCollector)),
		}
		err := k8sClient.List(context.TODO(), configMapList, listOps)
		if err != nil {
			fmt.Printf("error listing ConfigMaps: %v", err)
		}
		for i, item := range configMapList.Items {
			fmt.Printf("config %d: %s\n", i, item.Data["collector.yaml"])
		}
		fmt.Printf("total configurations %d\n", len(configMapList.Items))
		provider := cluster.NewProvider()
		kubeconfig, _ := provider.KubeConfig("mdai-local", false)
		opt := &helmclient.KubeConfClientOptions{
			Options: &helmclient.Options{
				Namespace:        "default",
				RepositoryCache:  os.TempDir() + "/.helmcache",
				RepositoryConfig: os.TempDir() + "/.helmrepo",
				Debug:            false,
				// DebugLog: func(format string, v ...interface{}) {
				//  Change this to your own logger. Default is 'log.Printf(format, v...)'.
				// },
			},
			KubeContext: "",
			KubeConfig:  []byte(kubeconfig),
		}
		opt.Options.DebugLog = func(_ string, _ ...interface{}) {}

		helmClient, _ := helmclient.NewClientFromKubeConf(opt, helmclient.Timeout(time.Second*60))
		releases, _ := helmClient.ListDeployedReleases()
		for _, release := range releases {
			fmt.Printf("%s [%s] (%s)\n", release.Name, release.Chart.Metadata.Version, release.Chart.Metadata.AppVersion)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
