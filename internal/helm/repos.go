package helm

import (
	"os"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

var repos = []repo.Entry{
	{
		Name: "mydecisive",
		URL:  "https://decisiveai.github.io/mdai-helm-charts",
	},
	{
		Name: "prometheus-community",
		URL:  "https://prometheus-community.github.io/helm-charts",
	},
	{
		Name: "jetstack",
		URL:  "https://charts.jetstack.io",
	},
	{
		Name: "opentelemetry",
		URL:  "https://open-telemetry.github.io/opentelemetry-helm-charts",
	},
}

func AddRepos() error {
	tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
	if err != nil {
		return errors.Wrap(err, "failed to create temp dir")
	}
	defer os.Remove(tmpfile.Name())
	settings := cli.New()
	settings.RepositoryConfig = tmpfile.Name()

	for _, repo := range repos {
		if err := addRepo(repo.Name, repo.URL, settings); err != nil {
			return errors.Wrap(err, "failed to add repo "+repo.Name)
		}
	}
	return nil
}

func addRepo(name, url string, settings *cli.EnvSettings) error {
	file := settings.RepositoryConfig
	repoFile, err := repo.LoadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if repoFile.Has(name) {
		return nil
	}

	entry := repo.Entry{
		Name: name,
		URL:  url,
	}

	repo, err := repo.NewChartRepository(&entry, getter.All(settings))
	if err != nil {
		return err
	}

	if _, err := repo.DownloadIndexFile(); err != nil {
		return err
	}

	repoFile.Update(&entry)

	return repoFile.WriteFile(file, 0644)
}
