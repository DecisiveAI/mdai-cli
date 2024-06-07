package kind

import (
	"embed"
	"os"
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
)

//go:embed templates/kind-config.yaml
var embedFS embed.FS

func Install() string {
	kindRawConfig, _ := embedFS.ReadFile("templates/kind-config.yaml")

	f, _ := os.CreateTemp("", "kubeconfig")
	defer os.Remove(f.Name())

	provider := cluster.NewProvider()
	if err := provider.Create("mdai-local",
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
		cluster.CreateWithWaitForReady(30*time.Second),
		cluster.CreateWithRawConfig(kindRawConfig),
	); err != nil {
		panic(err)
	}
	kubeconfig, _ := provider.KubeConfig("mdai-local", false)

	f.WriteString(kubeconfig)
	f.Close()

	return kubeconfig
}
