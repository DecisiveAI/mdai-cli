package helm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"golang.org/x/mod/semver"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type Client struct {
	envSettings *cli.EnvSettings
	logger      *log.Logger
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
		client.logger = log.FromContext(ctx)
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

func (c *Client) InstallChart(helmchart string) error {
	chartSpec, err := getChartSpec(helmchart)
	if err != nil {
		return fmt.Errorf("failed to get chart spec: %w", err)
	}

	actionConfig, settings, err := c.getActionConfig(chartSpec.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get action config: %w", err)
	}

	helmChart, err := loadChart(fmt.Sprintf(chartSpec.ChartURL, chartSpec.ReleaseName, chartSpec.Version), settings)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	getClient := action.NewGet(actionConfig)
	helmRelease, err := getClient.Run(chartSpec.ReleaseName)
	if errors.Is(err, driver.ErrReleaseNotFound) {
		installClient := action.NewInstall(actionConfig)
		installClient.ReleaseName = chartSpec.ReleaseName
		installClient.Namespace = chartSpec.Namespace
		installClient.CreateNamespace = chartSpec.CreateNamespace
		installClient.Wait = chartSpec.Wait
		installClient.Timeout = chartSpec.Timeout

		if _, err = installClient.Run(helmChart, chartSpec.Values); err != nil {
			return fmt.Errorf("failed to install chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
		}
		return nil
	}

	if semver.Compare(helmRelease.Chart.Metadata.Version, chartSpec.Version) < 0 {
		upgradeClient := action.NewUpgrade(actionConfig)
		upgradeClient.Namespace = chartSpec.Namespace
		upgradeClient.Wait = chartSpec.Wait
		upgradeClient.Timeout = chartSpec.Timeout

		if _, err = upgradeClient.Run(chartSpec.ReleaseName, helmChart, chartSpec.Values); err != nil {
			return fmt.Errorf("failed to upgrade chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
		}
	}
	return nil
}

func (c *Client) UninstallChart(helmchart string) error {
	chartSpec, err := getChartSpec(helmchart)
	if err != nil {
		return fmt.Errorf("failed to get chart spec: %w", err)
	}

	actionConfig, _, err := c.getActionConfig(chartSpec.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get action config: %w", err)
	}

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(chartSpec.ReleaseName); errors.Is(err, driver.ErrReleaseNotFound) {
		return nil
	}

	uninstallClient := action.NewUninstall(actionConfig)
	if _, err := uninstallClient.Run(chartSpec.ReleaseName); err != nil {
		return fmt.Errorf("failed to uninstall chart %s in namespace %s: %w", chartSpec.ReleaseName, chartSpec.Namespace, err)
	}

	return nil
}

func (c *Client) Outdated() ([][]string, error) {
	releases, err := c.Releases()
	if err != nil {
		return nil, err
	}

	var rows [][]string
	expectedCharts := []string{"prometheus", "opentelemetry-operator", "mydecisive-engine-operator", "mdai-console", "datalyzer"}
	seenCharts := make(map[string]bool, len(expectedCharts))

	for _, rel := range releases {
		chartSpec, err := getChartSpec(rel.Name)
		if chartSpec == nil || err != nil {
			continue
		}
		seenCharts[rel.Name] = true
		var outdated string
		current := rel.Chart.Metadata.Version
		if !strings.HasPrefix(current, "v") {
			current = "v" + current
		}
		wanted := chartSpec.Version
		if !strings.HasPrefix(wanted, "v") {
			wanted = "v" + wanted
		}
		switch semver.Compare(current, wanted) {
		case -1: // v < w
			outdated = "✗"
		case 0, 1: // v == w, v > w
			outdated = "✓"
		}
		rows = append(rows, []string{outdated, rel.Name, current, wanted})
	}

	for _, rel := range expectedCharts {
		if ok := seenCharts[rel]; !ok {
			chartSpec, err := getChartSpec(rel)
			if chartSpec == nil || err != nil {
				continue
			}
			rows = append(rows, []string{"✗", rel, "", "v" + chartSpec.Version})
		}
	}
	return rows, nil
}

func (c *Client) Releases() ([]*release.Release, error) {
	actionConfig := new(action.Configuration)
	settings := c.envSettings
	if err := actionConfig.Init(settings.RESTClientGetter(), "", "", nil); err != nil {
		return nil, fmt.Errorf("failed to initialize helm client: %w", err)
	}
	client := action.NewList(actionConfig)
	client.AllNamespaces = true

	releases, err := client.Run()
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func loadChart(chartURL string, settings *cli.EnvSettings) (*chart.Chart, error) {
	chartPath, err := (&action.ChartPathOptions{}).LocateChart(chartURL, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}
	helmChart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}
	return helmChart, nil
}

func (c *Client) getActionConfig(namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	settings := c.envSettings
	settings.SetNamespace(namespace)
	actionConfig := new(action.Configuration)

	logFunc := func(format string, v ...interface{}) { c.logger.Debugf(format+"\r", v...) }
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "", logFunc); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize helm client: %w", err)
	}
	return actionConfig, settings, nil
}
