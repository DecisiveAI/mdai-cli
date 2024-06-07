package helm

import (
	"embed"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
)

//go:embed templates/*
var embedFS embed.FS

var chartSpecs = map[string]helmclient.ChartSpec{}

func init() {
	certManagerValuesYaml, _ := embedFS.ReadFile("templates/cert-manager-values.yaml")
	opentelemetryOperatorValuesYaml, _ := embedFS.ReadFile("templates/opentelemetry-operator-values.yaml")
	prometheusValuesYaml, _ := embedFS.ReadFile("templates/prometheus-values.yaml")

	chartSpecs = make(map[string]helmclient.ChartSpec)

	chartSpecs["cert-manager"] = helmclient.ChartSpec{
		ReleaseName:     "cert-manager",
		ChartName:       "jetstack/cert-manager",
		Namespace:       "cert-manager",
		Version:         "1.13.1",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(certManagerValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["opentelemetry-operator"] = helmclient.ChartSpec{
		ReleaseName: "opentelemetry-operator",
		// ChartName:       "mydecisive/opentelemetry-operator",
		ChartName: "opentelemetry/opentelemetry-operator",
		Namespace: "opentelemetry-operator-system",
		// Version:         "0.43.1",
		Version:         "0.61.0",
		UpgradeCRDs:     true,
		Wait:            true,
		ValuesYaml:      string(opentelemetryOperatorValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["prometheus"] = helmclient.ChartSpec{
		ReleaseName:     "prometheus",
		ChartName:       "prometheus-community/prometheus",
		Namespace:       "default",
		Version:         "25.21.0",
		UpgradeCRDs:     true,
		Wait:            false,
		ValuesYaml:      string(prometheusValuesYaml),
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["metrics-server"] = helmclient.ChartSpec{
		ReleaseName:     "metrics-server",
		ChartName:       "metrics-server",
		Namespace:       "kube-system",
		Version:         "3.12.1",
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["mdai-api"] = helmclient.ChartSpec{
		ReleaseName:     "mdai-api",
		ChartName:       "mydecisive/mdai-api",
		Namespace:       "default",
		Version:         "0.0.3",
		UpgradeCRDs:     true,
		Wait:            false,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["mdai-console"] = helmclient.ChartSpec{
		ReleaseName: "mdai-console",
		ChartName:   "mydecisive/mdai-console",
		Namespace:   "default",
		Version:     "0.0.7",
		UpgradeCRDs: true,
		Wait:        false,
		/*
					      values: {
			        'ingress': {
			          'userPoolArn': mdaiUserPool.userPoolArn,
			          'userPoolClientId': mdaiAppClient.userPoolClientId,
			          'userPoolDomain': config.MDAI_COGNITO.USER_POOL_DOMAIN,
			        },
			        'env': {
			          'MDAI_UI_ACM_ARN': process.env.MDAI_UI_ACM_ARN,
			        }
			      }
		*/
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["datalyzer"] = helmclient.ChartSpec{
		ReleaseName:     "datalyzer",
		ChartName:       "mydecisive/datalyzer",
		Namespace:       "default",
		Version:         "0.0.1",
		UpgradeCRDs:     true,
		Wait:            false,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}

	chartSpecs["mdai-operator"] = helmclient.ChartSpec{
		ReleaseName:     "mydecisive-engine-operator",
		ChartName:       "mydecisive/mydecisive-engine-operator",
		Namespace:       "mydecisive-engine-operator-system",
		Version:         "0.1.0",
		UpgradeCRDs:     true,
		Wait:            true,
		Replace:         true,
		CreateNamespace: true,
		Timeout:         60 * time.Second, // nolint: gomnd
	}
}

func GetChartSpec(name string) helmclient.ChartSpec {
	return chartSpecs[name]
}
