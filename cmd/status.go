package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	//"github.com/decisiveai/opentelemetry-operator/apis/v1alpha1"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type deployment struct {
	name      string
	namespace string
}

var deployments = []deployment{
	{name: "datalyzer-deployment", namespace: "mdai-otel-nucleus"},
	{name: "mdai-api", namespace: "mdai-otel-nucleus"},
	{name: "mdai-console", namespace: "mdai-otel-nucleus"},
	{name: "prometheus-server", namespace: "mdai-otel-nucleus"},
	{name: "prometheus-kube-state-metrics", namespace: "mdai-otel-nucleus"},
	{name: "test-collector-collector", namespace: "mdai-otel-nucleus"},
	{name: "mydecisive-engine-operator-controller-manager", namespace: "mdai-otel-nucleus"},
	{name: "opentelemetry-operator", namespace: "mdai-otel-nucleus"},
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
		actionConfig := new(action.Configuration)
		settings := cli.New()
		if err := actionConfig.Init(settings.RESTClientGetter(), "", "secrets", nil); err != nil {
			panic(err)
		}
		client := action.NewList(actionConfig)
		client.AllNamespaces = true

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
				lastIndex := strings.LastIndex(d.Labels["helm.sh/chart"], "-")
				release = d.Labels["helm.sh/chart"][:lastIndex]
				version = d.Labels["helm.sh/chart"][lastIndex+1:]
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
