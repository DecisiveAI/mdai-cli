package types

import "time"

type ChartSpec struct {
	ReleaseName     string
	ChartURL        string
	Namespace       string
	Values          map[string]any
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
