package types

import "time"

type ChartSpec struct {
	ReleaseName     string
	ChartName       string
	Namespace       string
	ValuesYaml      string
	Version         string
	CreateNamespace bool
	Replace         bool
	Wait            bool
	Timeout         time.Duration
	SkipCRDs        bool
	UpgradeCRDs     bool
	Force           bool
	Recreate        bool
}
