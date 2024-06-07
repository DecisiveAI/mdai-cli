package helm

import "helm.sh/helm/v3/pkg/repo"

var ChartRepos = []repo.Entry{
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
