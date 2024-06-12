package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/decisiveai/opentelemetry-operator/apis/v1alpha1"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type deployment struct {
	name      string
	namespace string
}

var deployments = []deployment{
	{name: "datalyzer-deployment", namespace: "default"},
	{name: "mdai-api", namespace: "default"},
	{name: "mdai-console", namespace: "default"},
	{name: "prometheus-server", namespace: "default"},
	{name: "prometheus-kube-state-metrics", namespace: "default"},
	{name: "test-collector-collector", namespace: "default"},
	{name: "mydecisive-engine-operator-controller-manager", namespace: "mydecisive-engine-operator-system"},
	{name: "opentelemetry-operator", namespace: "opentelemetry-operator-system"},
	{name: "cert-manager", namespace: "cert-manager"},
	{name: "cert-manager-cainjector", namespace: "cert-manager"},
	{name: "cert-manager-webhook", namespace: "cert-manager"},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfigOrDie()
		scheme := runtime.NewScheme()
		v1alpha1.AddToScheme(scheme)
		k8sClient, _ := client.New(cfg, client.Options{Scheme: scheme})
		//"github.com/open-telemetry/opentelemetry-operator/apis/v1alpha1"
		collectors := &v1alpha1.OpenTelemetryCollectorList{}
		opts := client.MatchingLabels(map[string]string{
			"app.kubernetes.io/managed-by": "opentelemetry-operator", // mydecisive-engine-operator",
		})
		err := k8sClient.List(context.TODO(), collectors, opts)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("%+v\n", collectors)
		os.Exit(0)
		/*
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
		*/
		//provider := cluster.NewProvider()
		//kubeconfig, _ := provider.KubeConfig("mdai-local", false)
		actionConfig := new(action.Configuration)
		settings := cli.New()
		if err := actionConfig.Init(settings.RESTClientGetter(), "", "secrets", nil); err != nil {
			panic(err)
		}
		client := action.NewList(actionConfig)
		client.AllNamespaces = true
		registry.NewClient()
		ch := chart.Chart{}
		action.NewInstall(actionConfig).Run(&ch, nil)

		releases, _ := client.Run()
		for _, release := range releases {
			fmt.Printf("Namespace: %s, Release Name: %s, Chart: %s, Version: %s, AppVersion: %s, First Deployed: %s, Last Deployed: %s\n", release.Namespace, release.Name, release.Chart.Metadata.Name, release.Chart.Metadata.Version, release.Chart.Metadata.AppVersion, release.Info.FirstDeployed, release.Info.LastDeployed)
		}

		clientset, _ := kubernetes.NewForConfig(cfg)

		for _, deployment := range deployments {
			d, _ := clientset.AppsV1().Deployments(deployment.namespace).Get(context.TODO(), deployment.name, metav1.GetOptions{})
			labelSelector := metav1.FormatLabelSelector(d.Spec.Selector)
			var release, version string
			if _, ok := d.Labels["helm.sh/chart"]; ok {
				helmInfo := strings.Split(d.Labels["helm.sh/chart"], "-")
				release = helmInfo[0]
				version = helmInfo[1]
			}
			fmt.Printf("Deployment: %s (%s) [%s]\n", deployment.name, release, version)

			pod, _ := clientset.CoreV1().Pods(deployment.namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
			for _, pod := range pod.Items {
				fmt.Printf("  Pod: %s\n", pod.Name)
				for _, containerStatus := range pod.Status.ContainerStatuses {
					image := containerStatus.Image
					lastPullTime := containerStatus.State.Running.StartedAt.Time

					fmt.Printf("    Container: %s\n", containerStatus.Name)
					fmt.Printf("      Image: %s\n", image)
					fmt.Printf("      Last Pull: %s\n", lastPullTime.Format(time.RFC3339))
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
