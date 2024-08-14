package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kubehelper"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployment struct {
	name      string
	namespace string
}

var deployments = []deployment{
	{name: "datalyzer-deployment", namespace: "mdai"},
	{name: "mdai-console", namespace: "mdai"},
	{name: "prometheus-server", namespace: "mdai"},
	{name: "prometheus-kube-state-metrics", namespace: "mdai"},
	{name: "gateway-collector", namespace: "mdai"},
	{name: "mydecisive-engine-operator-controller-manager", namespace: "mdai"},
	{name: "opentelemetry-operator", namespace: "mdai"},
	{name: "cert-manager", namespace: "cert-manager"},
	{name: "cert-manager-cainjector", namespace: "cert-manager"},
	{name: "cert-manager-webhook", namespace: "cert-manager"},
}

func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show kubernetes deployment status",
		Long:  `show installed helm charts, deployments with their statuses`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			helmclient := mdaihelm.NewClient(mdaihelm.WithContext(ctx))
			releases, err := helmclient.Releases()
			if err != nil {
				return fmt.Errorf("failed to get releases from cluster: %w", err)
			}
			t := table.New().
				BorderHeader(false).
				Border(lipgloss.HiddenBorder()).
				StyleFunc(func(row, col int) lipgloss.Style {
					switch {
					case row == 0:
						return HeaderStyle
					case row%2 == 0:
						return EvenRowStyle
					default:
						return OddRowStyle
					}
				}).
				Headers("NAMESPACE", "RELEASE", "CHART", "VERSION", "APPVERSION", "FIRST DEPLOY", "LAST DEPLOY") // .
			for _, rel := range releases {
				t.Row(rel.Namespace, rel.Name, rel.Chart.Metadata.Name, rel.Chart.Metadata.Version, rel.Chart.Metadata.AppVersion, rel.Info.FirstDeployed.String(), rel.Info.LastDeployed.String())
			}

			fmt.Println(t)
			fmt.Printf("kubeconfig: %s\nkubecontext: %s\n",
				PurpleStyle.Render(ctx.Value(mdaitypes.Kubeconfig{}).(string)),
				PurpleStyle.Render(ctx.Value(mdaitypes.Kubecontext{}).(string)),
			)

			for _, deployment := range deployments {
				helper, err := kubehelper.New(kubehelper.WithContext(ctx))
				if err != nil {
					return fmt.Errorf("failed creating kubehelper: %w", err)
				}
				d, err := helper.GetDeployment(ctx, deployment.name, deployment.namespace)
				if err != nil {
					continue
				}
				labelSelector := metav1.FormatLabelSelector(d.Spec.Selector)
				var release, version string
				if _, ok := d.Labels["helm.sh/chart"]; ok {
					lastIndex := strings.LastIndex(d.Labels["helm.sh/chart"], "-")
					release = d.Labels["helm.sh/chart"][:lastIndex]
					version = d.Labels["helm.sh/chart"][lastIndex+1:]
				}
				fmt.Printf("Deployment: %s (%s) [%s]\n",
					LightPurpleStyle.Render(deployment.name),
					PurpleStyle.Render(release),
					PurpleStyle.Render(version),
				)

				pod, err := helper.GetPodByLabel(ctx, deployment.namespace, labelSelector)
				if err != nil {
					continue
				}
				for _, p := range pod.Items {
					fmt.Printf("  Pod: %s\n", WhiteStyle.Render(p.Name))
					for _, containerStatus := range p.Status.ContainerStatuses {
						image := containerStatus.Image
						lastPullTime := containerStatus.State.Running.StartedAt.Time
						fmt.Printf("    Container: %s\n", WhiteStyle.Render(containerStatus.Name))
						fmt.Printf("      Image: %s\n", WhiteStyle.Render(image))
						fmt.Printf("      Last Pull: %s\n", WhiteStyle.Render(lastPullTime.Format(time.RFC3339)))
					}
				}
			}
			return nil
		},
	}
	cmd.Hidden = true
	return cmd
}
