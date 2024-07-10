package kind

import (
	"embed"
	"os"
	"time"

	"github.com/pkg/errors"

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
		c.errs <- errors.Wrap(err, "failed to create temporary kubeconfig file")
		return "", errors.Wrap(err, "failed to create temporary kubeconfig file")
	}
	defer os.Remove(f.Name())

	provider := cluster.NewProvider()
	c.messages <- "listing nodes in cluster " + c.clusterName + "..."
	n, err := provider.ListNodes(c.clusterName)
	if err != nil {
		c.errs <- errors.Wrap(err, "error listing nodes")
		return "", errors.Wrap(err, "error listing nodes")
	}
	if len(n) == 0 {
		c.messages <- "cluster " + c.clusterName + " does not exist, creating..."
		if err := provider.Create(c.clusterName,
			cluster.CreateWithDisplayUsage(false),
			cluster.CreateWithDisplaySalutation(false),
			cluster.CreateWithWaitForReady(30*time.Second),
			cluster.CreateWithRawConfig(kindRawConfig),
		); err != nil {
			c.errs <- errors.Wrap(err, "error creating cluster")
			return "", errors.Wrap(err, "error creating cluster")
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
