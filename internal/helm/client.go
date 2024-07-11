package helm

import (
	"fmt"
	"os"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
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
			c.errs <- errors.Wrap(err, "failed to add repo "+repo.Name)
			return errors.Wrap(err, "failed to add repo "+repo.Name)
		}
		c.messages <- "added repo " + repo.Name
	}
	return nil
}

func (c *Client) addRepo(name, url string) error {
	file := c.cliEnvSettings.RepositoryConfig
	repoFile, err := repo.LoadFile(file)
	if err != nil && !os.IsNotExist(err) {
		c.errs <- errors.Wrap(err, "failed to load Helm repo index file")
		return err
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
		c.errs <- errors.Wrap(err, "failed to create chart repository")
		return err
	}

	if _, err := repo.DownloadIndexFile(); err != nil {
		c.errs <- errors.Wrap(err, "failed to download Helm repo index file")
		return err
	}

	repoFile.Update(&entry)
	err = repoFile.WriteFile(file, 0644)
	c.errs <- errors.Wrap(err, "failed to write Helm repo index file")
	return err
}

func (c *Client) InstallChart(helmchart string) error {
	chartSpec := getChartSpec(helmchart)
	settings := c.cliEnvSettings
	settings.SetNamespace(chartSpec.Namespace)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), chartSpec.Namespace, "", func(format string, v ...interface{}) { c.debug <- fmt.Sprintf(format, v) }); err != nil {
		c.errs <- errors.Wrap(err, "failed to initialize Helm client")
		return errors.Wrap(err, "failed to initialize Helm client")
	}

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	_, err := histClient.Run(chartSpec.ReleaseName)
	switch err {
	case driver.ErrReleaseNotFound:
		client := action.NewInstall(actionConfig)
		client.ReleaseName = chartSpec.ReleaseName
		client.Namespace = chartSpec.Namespace
		client.CreateNamespace = chartSpec.CreateNamespace
		client.Wait = chartSpec.Wait
		client.Timeout = chartSpec.Timeout

		chartPath, err := client.ChartPathOptions.LocateChart(chartSpec.ChartName, settings)
		if err != nil {
			c.errs <- errors.Wrap(err, "failed to locate chart")
			return errors.Wrap(err, "failed to locate chart")
		}

		chart, values, err := c.getChartAndValues(chartSpec, chartPath)
		if err != nil {
			c.errs <- errors.Wrap(err, "failed to get chart and values")
			return errors.Wrap(err, "failed to get chart and values")
		}

		if _, err = client.Run(chart, values); err != nil {
			c.errs <- errors.Wrap(err, "failed to install chart "+chartSpec.ReleaseName+" in namespace "+chartSpec.Namespace)
			return errors.Wrap(err, "failed to install chart "+chartSpec.ReleaseName+" in namespace "+chartSpec.Namespace)
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
			c.errs <- errors.Wrap(err, "failed to locate chart")
			return errors.Wrap(err, "failed to locate chart")
		}

		chart, values, err := c.getChartAndValues(chartSpec, chartPath)
		if err != nil {
			c.errs <- errors.Wrap(err, "failed to get chart and values")
			return errors.Wrap(err, "failed to get chart and values")
		}

		if _, err = client.Run(chartSpec.ReleaseName, chart, values); err != nil {
			c.errs <- errors.Wrap(err, "failed to upgrade chart "+chartSpec.ReleaseName+" in namespace "+chartSpec.Namespace)
			return errors.Wrap(err, "failed to upgrade chart "+chartSpec.ReleaseName+" in namespace "+chartSpec.Namespace)

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
		c.errs <- errors.Wrap(err, "failed to initialize Helm client")
		return errors.Wrap(err, "failed to initialize Helm client")
	}

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(chartSpec.ReleaseName); err == driver.ErrReleaseNotFound {
		c.messages <- "chart " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " not found. skipping uninstall."
		return nil
	}

	uninstall := action.NewUninstall(actionConfig)

	if _, err := uninstall.Run(chartSpec.ReleaseName); err != nil {
		c.errs <- errors.Wrap(err, "failed to uninstall chart "+chartSpec.ReleaseName+" in namespace "+chartSpec.Namespace)
		return errors.Wrap(err, "failed to uninstall chart "+chartSpec.ReleaseName+" in namespace "+chartSpec.Namespace)
	}
	c.messages <- "release " + chartSpec.ReleaseName + " in namespace " + chartSpec.Namespace + " uninstalled successfully"

	return nil
}

func (c *Client) getChartAndValues(chartSpec mdaitypes.ChartSpec, chartPath string) (*chart.Chart, map[string]any, error) {
	chart, err := loader.Load(chartPath)
	if err != nil {
		c.errs <- errors.Wrap(err, "failed to load chart")
		return nil, nil, errors.Wrap(err, "failed to load chart")
	}

	values := map[string]any{}
	if chartSpec.ValuesYaml != "" {
		if err := yaml.Unmarshal([]byte(chartSpec.ValuesYaml), &values); err != nil {
			c.errs <- errors.Wrap(err, "failed to parse ValuesYaml")
			return nil, nil, errors.Wrap(err, "failed to parse ValuesYaml")
		}
	}
	return chart, values, nil
}
