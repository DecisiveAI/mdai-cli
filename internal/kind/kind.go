package kind

import (
	"embed"
	"os"
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
)

//go:embed templates/kind-config.yaml
var embedFS embed.FS

func Install(clusterName string) string {
	kindRawConfig, _ := embedFS.ReadFile("templates/kind-config.yaml")

	f, _ := os.CreateTemp("", "kubeconfig")
	defer os.Remove(f.Name())

	provider := cluster.NewProvider()
	if err := provider.Create(clusterName,
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
		cluster.CreateWithWaitForReady(30*time.Second),
		cluster.CreateWithRawConfig(kindRawConfig),
	); err != nil {
		if err.Error() == `node(s) already exist for a cluster with the name \`+clusterName+`"` {
			return ""
		}
		panic(err)
	}
	kubeconfig, _ := provider.KubeConfig(clusterName, false)

	f.WriteString(kubeconfig)
	f.Close()

	return kubeconfig
}
