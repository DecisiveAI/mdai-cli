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
	Name         string   `json:"-"`
	Tier         string   `json:"tier"`
	Store        string   `json:"store"`
	Capacity     string   `json:"capacity"`
	CapacityType string   `json:"capacity_type"`
	Duration     string   `json:"duration"`
	DurationType string   `json:"duration_type"`
	Format       string   `json:"format"`
	Description  string   `json:"description"`
	Pipelines    []string `json:"pipelines"`
}

func (f TieredStorageOutputAddFlags) SuccessString() string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, `tiered storage added successfully, "%s"`, f.Name)
	fmt.Printf("Name: %s\nTier: %s\nStore: %s\nCapacity: %s %s\nDuration: %s %s\nFormat: %s\nDescription: %s\nPipelines: %v",
		f.Name, f.Store, f.Tier, f.Capacity, f.CapacityType, f.Duration, f.DurationType, f.Format, f.Description, f.Pipelines)
	return sb.String()
}
