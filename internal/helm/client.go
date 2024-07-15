package helm

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type Client struct {
	messages       chan string
	debug          chan string
	errs           chan error
	cliEnvSettings *cli.EnvSettings
}

func NewClient(
	messages chan string,
	debug chan string,
	errs chan error,
	repositoryConfig string,
) *Client {
	cliEnvSettings := cli.New()
	cliEnvSettings.RepositoryConfig = repositoryConfig
	return &Client{
		messages:       messages,
		debug:          debug,
		errs:           errs,
		cliEnvSettings: cliEnvSettings,
	}
}

func (c *Client) AddRepos() error {
	for _, repo := range repos {
		if err := c.addRepo(repo.Name, repo.URL); err != nil {
			c.errs <- fmt.Errorf("failed to add repo %s: %w", repo.Name, err)
			return fmt.Errorf("failed to add repo %s: %w", repo.Name, err)
		}
		c.messages <- "added repo " + repo.Name
	}
	return nil
}

func (c *Client) addRepo(name, url string) error {
	file := c.cliEnvSettings.RepositoryConfig
	repoFile, err := repo.LoadFile(file)
	if err != nil && !os.IsNotExist(err) {
		c.errs <- fmt.Errorf("failed to load helm repo index file: %w", err)
		return fmt.Errorf("failed to load helm repo index file: %w", err)
	}

	if repoFile.Has(name) {
		c.messages <- "repo " + name + " already exists. skipping."
		return nil
	}

	entry := repo.Entry{
		Name: name,
		URL:  url,
	}

	repo, err := repo.NewChartRepository(&entry, getter.All(c.cliEnvSettings))
	if err != nil {
		c.errs <- fmt.Errorf("failed to create chart repository: %w", err)
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	if _, err := repo.DownloadIndexFile(); err != nil {
		c.errs <- fmt.Errorf("failed to download helm repo index file: %w", err)
		return fmt.Errorf("failed to download helm repo index file: %w", err)
	}

	repoFile.Update(&entry)
	if err = repoFile.WriteFile(file, 0o644); err != nil { // nolint: mnd
		c.errs <- fmt.Errorf("failed to write helm repo index file: %w", err)
		return fmt.Errorf("failed to write helm repo index file: %w", err)
	}
	return nil
}

func (c *Client) InstallChart(helmchart string) error {
	chartSpec := getChartSpec(helmchart)
	settings := c.cliEnvSettings
	settings.SetNamespace(chartSpec.Namespace)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), chartSpec.Namespace, "", func(format string, v ...interface{}) { c.debug <- fmt.Sprintf(format, v) }); err != nil {
		c.errs <- fmt.Errorf("failed to initialize helm client: %w", err)
		return fmt.Errorf("failed to initialize helm client: %w", err)
	}

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	_, err := histClient.Run(chartSpec.ReleaseName)
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
			c.errs <- fmt.Errorf("failed to locate chart: %w", err)
			return fmt.Errorf("failed to locate chart: %w", err)
		}

		chart, err := loader.Load(chartPath)
		if err != nil {
			c.errs <- fmt.Errorf("failed to load chart: %w", err)
			return fmt.Errorf("failed to load chart: %w", err)
		}

		if _, err = client.Run(chart, chartSpec.Values); err != nil {
			c.errs <- fmt.Errorf("failed to install chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
			return fmt.Errorf("failed to install chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
		}
		c.messages <- "chart " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " installed successfully"
		return nil
	default:
		client := action.NewUpgrade(actionConfig)
		client.Namespace = chartSpec.Namespace
		client.Wait = chartSpec.Wait
		client.Timeout = chartSpec.Timeout

		chartPath, err := client.ChartPathOptions.LocateChart(chartSpec.ChartName, settings)
		if err != nil {
			c.errs <- fmt.Errorf("failed to locate chart: %w", err)
			return fmt.Errorf("failed to locate chart: %w", err)
		}

		chart, err := loader.Load(chartPath)
		if err != nil {
			c.errs <- fmt.Errorf("failed to load chart: %w", err)
			return fmt.Errorf("failed to load chart: %w", err)
		}

		if _, err = client.Run(chartSpec.ReleaseName, chart, chartSpec.Values); err != nil {
			c.errs <- fmt.Errorf("failed to upgrade chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
			return fmt.Errorf("failed to upgrade chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
		}
		c.messages <- "chart " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " upgraded successfully"
		return nil
	}
}

func (c *Client) UninstallChart(helmchart string) error {
	chartSpec := getChartSpec(helmchart)
	settings := c.cliEnvSettings
	settings.SetNamespace(chartSpec.Namespace)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), chartSpec.Namespace, "", func(format string, v ...interface{}) { c.debug <- fmt.Sprintf(format, v) }); err != nil {
		c.errs <- fmt.Errorf("failed to initialize helm client: %w", err)
		return fmt.Errorf("failed to initialize helm client: %w", err)
	}

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(chartSpec.ReleaseName); errors.Is(err, driver.ErrReleaseNotFound) {
		c.messages <- "chart " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " not found. skipping uninstall."
		return nil
	}

	uninstall := action.NewUninstall(actionConfig)

	if _, err := uninstall.Run(chartSpec.ReleaseName); err != nil {
		c.errs <- fmt.Errorf("failed to uninstall chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
		return fmt.Errorf("failed to uninstall chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
	}
	c.messages <- "release " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " uninstalled successfully"

	return nil
}
