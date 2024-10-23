package types

import (
	"fmt"
	"strings"
)

type (
	Kubeconfig  struct{}
	Kubecontext struct{}
)

type TieredStorageOutputAddFlags struct {
	Key             string   `json:"-"`
	Tier            string   `json:"tier"`
	Capacity        string   `json:"capacity"`
	RetentionPeriod string   `json:"retention_period"`
	Format          string   `json:"format"`
	Description     string   `json:"description"`
	Pipelines       []string `json:"pipelines"`
	Location        string   `json:"location"`
}

func (f TieredStorageOutputAddFlags) SuccessString() string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, `tiered storage added successfully, "%s"`, f.Key)
	fmt.Printf("Key: %s\nTier: %s\nCapacity: %s\nRetention Period: %s\nFormat: %s\nDescription: %s\nPipelines: %v\nLocation: %s\n",
		f.Key, f.Tier, f.Capacity, f.RetentionPeriod, f.Format, f.Description, f.Pipelines, f.Location)
	return sb.String()
}
