package kind

import (
	"embed"
	"fmt"
	"os"
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
)

//go:embed templates/kind-config.yaml
var embedFS embed.FS

type Client struct {
	messages    chan string
	debug       chan string
	errs        chan error
	clusterName string
}

func NewClient(
	messages chan string,
	debug chan string,
	errs chan error,
	clusterName string,
) *Client {
	return &Client{
		messages:    messages,
		debug:       debug,
		errs:        errs,
		clusterName: clusterName,
	}
}

func (c *Client) Install() (string, error) {
	kindRawConfig, _ := embedFS.ReadFile("templates/kind-config.yaml")

	f, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		c.errs <- fmt.Errorf("failed to create temporary kubeconfig file: %w", err)
		return "", fmt.Errorf("failed to create temporary kubeconfig file: %w", err)
	}
	defer os.Remove(f.Name())

	provider := cluster.NewProvider()
	c.messages <- "listing nodes in cluster " + c.clusterName + "..."
	n, err := provider.ListNodes(c.clusterName)
	if err != nil {
		c.errs <- fmt.Errorf("error listing nodes: %w", err)
		return "", fmt.Errorf("error listing nodes: %w", err)
	}
	if len(n) == 0 {
		c.messages <- "cluster " + c.clusterName + " does not exist, creating..."
		if err := provider.Create(c.clusterName,
			cluster.CreateWithDisplayUsage(false),
			cluster.CreateWithDisplaySalutation(false),
			cluster.CreateWithWaitForReady(30*time.Second), // nolint: mnd
			cluster.CreateWithRawConfig(kindRawConfig),
		); err != nil {
			c.errs <- fmt.Errorf("error creating cluster: %w", err)
			return "", fmt.Errorf("error creating cluster: %w", err)
		}
	} else {
		c.messages <- "cluster " + c.clusterName + " already exists"
	}
	c.messages <- "cluster " + c.clusterName + " is ready"
	kubeconfig, _ := provider.KubeConfig(c.clusterName, false)
	f.WriteString(kubeconfig)
	f.Close()

	return kubeconfig, nil
}
