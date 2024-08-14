package kind

import (
	"embed"
	"fmt"
	"os"
	"time"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"sigs.k8s.io/kind/pkg/cluster"
)

//go:embed templates/kind-config.yaml
var embedFS embed.FS

type Client struct {
	channels    mdaitypes.Channels
	clusterName string
}

func NewClient(
	channels mdaitypes.Channels,
	clusterName string,
) *Client {
	return &Client{
		channels:    channels,
		clusterName: clusterName,
	}
}

func (c *Client) Create() (string, error) {
	kindRawConfig, _ := embedFS.ReadFile("templates/kind-config.yaml")

	f, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		c.channels.Error(fmt.Errorf("failed to create temporary kubeconfig file: %w", err))
		return "", fmt.Errorf("failed to create temporary kubeconfig file: %w", err)
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()

	provider := cluster.NewProvider()
	c.channels.Message("listing nodes in cluster " + c.clusterName + "...")
	n, err := provider.ListNodes(c.clusterName)
	if err != nil { // the error returned is already wrapped, `errors.Wrap(err, "failed to list nodes")`
		c.channels.Error(err) //nolint: wrapcheck
		return "", err        //nolint: wrapcheck
	}
	if len(n) == 0 {
		c.channels.Message("cluster " + c.clusterName + " does not exist, creating...")
		if err := provider.Create(c.clusterName,
			cluster.CreateWithDisplayUsage(false),
			cluster.CreateWithDisplaySalutation(false),
			cluster.CreateWithWaitForReady(30*time.Second), //nolint: mnd
			cluster.CreateWithRawConfig(kindRawConfig),
		); err != nil {
			c.channels.Error(fmt.Errorf("error creating cluster: %w", err))
			return "", fmt.Errorf("error creating cluster: %w", err)
		}
	} else {
		c.channels.Message("cluster " + c.clusterName + " already exists")
	}
	c.channels.Message("cluster " + c.clusterName + " is ready")
	kubeconfig, _ := provider.KubeConfig(c.clusterName, false)
	if _, err := f.WriteString(kubeconfig); err != nil {
		return "", fmt.Errorf("error writing to temporary kubeconfig file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("error closing temporary kubeconfig file: %w", err)
	}

	return kubeconfig, nil
}

func (c *Client) Delete() error {
	provider := cluster.NewProvider()
	c.channels.Message("listing nodes in cluster " + c.clusterName + "...")
	n, err := provider.ListNodes(c.clusterName)
	if err != nil { // the error returned is already wrapped, `errors.Wrap(err, "failed to list nodes")`
		c.channels.Error(err) //nolint: wrapcheck
		return err            //nolint: wrapcheck
	}
	if len(n) == 0 {
		c.channels.Message("cluster " + c.clusterName + " does not exist, skipping deletion...")
		return nil
	}
	c.channels.Message(fmt.Sprintf("found %d node(s) in cluster %s to delete...", len(n), c.clusterName))
	return provider.Delete(c.clusterName, "")
}
