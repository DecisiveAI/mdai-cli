package helm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"golang.org/x/mod/semver"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type Client struct {
	channels    mdaitypes.Channels
	envSettings *cli.EnvSettings
}

type ClientOption func(*Client)

func WithContext(ctx context.Context) ClientOption {
	return func(client *Client) {
		if kubeconfig, ok := ctx.Value(mdaitypes.Kubeconfig{}).(string); ok {
			client.envSettings.KubeConfig = kubeconfig
		}
		if kubecontext, ok := ctx.Value(mdaitypes.Kubecontext{}).(string); ok {
			client.envSettings.KubeContext = kubecontext
		}
	}
}

func WithChannels(channels mdaitypes.Channels) ClientOption {
	return func(client *Client) {
		client.channels = channels
	}
}

func WithRepositoryConfig(repositoryConfig string) ClientOption {
	return func(client *Client) {
		client.envSettings.RepositoryConfig = repositoryConfig
	}
}

func NewClient(options ...ClientOption) *Client {
	client := new(Client)
	client.envSettings = cli.New()
	for _, option := range options {
		option(client)
	}

	return client
}

func (c *Client) AddRepos() error {
	for _, repo := range repos {
		if err := c.addRepo(repo.Name, repo.URL); err != nil {
			c.channels.Error(fmt.Errorf("failed to add repo %s: %w", repo.Name, err))
			return fmt.Errorf("failed to add repo %s: %w", repo.Name, err)
		}
		c.channels.Message("added repo " + repo.Name)
	}
	return nil
}

func (c *Client) addRepo(name, url string) error {
	file := c.envSettings.RepositoryConfig
	repoFile, err := helmrepo.LoadFile(file)
	if err != nil && !os.IsNotExist(err) {
		c.channels.Error(fmt.Errorf("failed to load helm repo index file: %w", err))
		return fmt.Errorf("failed to load helm repo index file: %w", err)
	}

	if repoFile.Has(name) {
		c.channels.Message("repo " + name + " already exists. skipping.")
		return nil
	}

	entry := helmrepo.Entry{
		Name: name,
		URL:  url,
	}

	repo, err := helmrepo.NewChartRepository(&entry, getter.All(c.envSettings))
	if err != nil {
		c.channels.Error(fmt.Errorf("failed to create chart repository: %w", err))
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	if _, err := repo.DownloadIndexFile(); err != nil {
		c.channels.Error(fmt.Errorf("failed to download helm repo index file: %w", err))
		return fmt.Errorf("failed to download helm repo index file: %w", err)
	}

	repoFile.Update(&entry)
	if err = repoFile.WriteFile(file, 0o644); err != nil { // nolint: mnd
		c.channels.Error(fmt.Errorf("failed to write helm repo index file: %w", err))
		return fmt.Errorf("failed to write helm repo index file: %w", err)
	}
	return nil
}

func (c *Client) InstallChart(helmchart string) error {
	chartSpec, err := getChartSpec(helmchart)
	if err != nil {
		return fmt.Errorf("failed to get chart spec: %w", err)
	}
	settings := c.envSettings
	settings.SetNamespace(chartSpec.Namespace)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), chartSpec.Namespace, "", func(format string, v ...interface{}) { c.channels.Debug(fmt.Sprintf(format, v)) }); err != nil {
		c.channels.Error(fmt.Errorf("failed to initialize helm client: %w", err))
		return fmt.Errorf("failed to initialize helm client: %w", err)
	}

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	_, err = histClient.Run(chartSpec.ReleaseName)
	switch {
	case errors.Is(err, driver.ErrReleaseNotFound):
		client := action.NewInstall(actionConfig)
		client.ReleaseName = chartSpec.ReleaseName
		client.Namespace = chartSpec.Namespace
		client.CreateNamespace = chartSpec.CreateNamespace
		client.Wait = chartSpec.Wait
		client.Timeout = chartSpec.Timeout

		chartPath, err := client.ChartPathOptions.LocateChart(chartSpec.ChartName, settings)
		if err != nil {
			c.channels.Error(fmt.Errorf("failed to locate chart: %w", err))
			return fmt.Errorf("failed to locate chart: %w", err)
		}

		chart, err := loader.Load(chartPath)
		if err != nil {
			c.channels.Error(fmt.Errorf("failed to load chart: %w", err))
			return fmt.Errorf("failed to load chart: %w", err)
		}

		if _, err = client.Run(chart, chartSpec.Values); err != nil {
			c.channels.Error(fmt.Errorf("failed to install chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err))
			return fmt.Errorf("failed to install chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
		}
		c.channels.Message("chart " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " installed successfully")
		return nil
	default:
		client := action.NewUpgrade(actionConfig)
		client.Namespace = chartSpec.Namespace
		client.Wait = chartSpec.Wait
		client.Timeout = chartSpec.Timeout

		chartPath, err := client.ChartPathOptions.LocateChart(chartSpec.ChartName, settings)
		if err != nil {
			c.channels.Error(fmt.Errorf("failed to locate chart: %w", err))
			return fmt.Errorf("failed to locate chart: %w", err)
		}

		chart, err := loader.Load(chartPath)
		if err != nil {
			c.channels.Error(fmt.Errorf("failed to load chart: %w", err))
			return fmt.Errorf("failed to load chart: %w", err)
		}

		if _, err = client.Run(chartSpec.ReleaseName, chart, chartSpec.Values); err != nil {
			c.channels.Error(fmt.Errorf("failed to upgrade chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err))
			return fmt.Errorf("failed to upgrade chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
		}
		c.channels.Message("chart " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " upgraded successfully")
		return nil
	}
}

func (c *Client) UninstallChart(helmchart string) error {
	chartSpec, err := getChartSpec(helmchart)
	if err != nil {
		return fmt.Errorf("failed to get chart spec: %w", err)
	}
	settings := c.envSettings
	settings.SetNamespace(chartSpec.Namespace)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), chartSpec.Namespace, "", func(format string, v ...interface{}) { c.channels.Debug(fmt.Sprintf(format, v)) }); err != nil {
		c.channels.Error(fmt.Errorf("failed to initialize helm client: %w", err))
		return fmt.Errorf("failed to initialize helm client: %w", err)
	}

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(chartSpec.ReleaseName); errors.Is(err, driver.ErrReleaseNotFound) {
		c.channels.Message("chart " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " not found. skipping uninstall.")
		return nil
	}

	uninstall := action.NewUninstall(actionConfig)
	if _, err := uninstall.Run(chartSpec.ReleaseName); err != nil {
		c.channels.Error(fmt.Errorf("failed to uninstall chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err))
		return fmt.Errorf("failed to uninstall chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
	}
	c.channels.Message("release " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " uninstalled successfully")

	return nil
}
