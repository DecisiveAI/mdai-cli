package cmd

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/decisiveai/mdai-cli/internal/kubehelper"
	"github.com/spf13/cobra"
)

type TieredStorageOutputAddFlags struct {
	key             string
	tier            string
	capacity        string
	retentionPeriod string
	format          string
	description     string
	pipelines       string
	location        string
}

func NewTieredStorageCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "tiered_storage",
		Short:   "manage tiered storages",
		Long:    `manage tiered storages`,
	}

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	cmd.AddCommand(
		NewTieredStorageAddCommand(),
		NewTieredStorageListCommand(),
		NewTieredStorageRemoveCommand(),
	)

	return cmd
}

func NewTieredStorageListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list tiered storages",
		Long:  `list tiered storages`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			helper, err := kubehelper.New(kubehelper.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("failed to initialize kubehelper: %w", err)
			}

			// Define headers for the table
			headers := []string{"Key", "Tier", "Status", "Description", "Pipelines", "Capacity", "Retention Period", "Format", "Location"}
			var rows [][]string

			// Fetch the ConfigMap
			configMap, err := helper.GetConfigMap(ctx, "tiered-storage", "mdai")
			if err != nil {
				return fmt.Errorf("failed to fetch tiered storages configmap: %w", err)
			}

			// Iterate through the ConfigMap and extract fields
			for k, v := range configMap.Data {
				// Assume `v` is a serialized JSON string or similar
				// Parse the value (JSON, YAML, etc.)
				var storageData struct {
					Tier            string   `json:"tier"`
					Status          string   `json:"status"`
					Description     string   `json:"description"`
					Pipelines       []string `json:"pipelines"`
					Capacity        string   `json:"capacity"`
					RetentionPeriod string   `json:"retention_period"`
					Format          string   `json:"format"`
					Location        string   `json:"location"`
				}

				// Parse JSON data from the ConfigMap
				if err != nil {
					return fmt.Errorf("failed to parse storage data for key %s: %w", k, v, err)
				}

				// Build the row for the table
				row := []string{
					k,                       // Key
					storageData.Tier,        // Tier
					storageData.Status,      // Status
					storageData.Description, // Description
					strings.Join(storageData.Pipelines, ", "), // Pipelines (comma-separated)
					storageData.Capacity,                      // Capacity
					storageData.RetentionPeriod,               // Retention Period
					storageData.Format,                        // Format
					storageData.Location,                      // Location
				}
				rows = append(rows, row)
			}

			// Print the tiered storage data in a table format
			printTieredStorage := func(headers []string, rows [][]string) {
				if len(rows) == 0 {
					fmt.Println("No tiered storage found.")
					return
				}

				// Render the table
				TieredStorageOutput := table.New().
					BorderHeader(false).
					Border(lipgloss.HiddenBorder()).
					StyleFunc(func(row, col int) lipgloss.Style {
						if row == 0 {
							return HeaderStyle
						}
						return lipgloss.NewStyle()
					}).
					Headers(headers...).
					Rows(rows...)
				fmt.Println(TieredStorageOutput)
			}

			// Call the function to display the data
			printTieredStorage(headers, rows)
			return nil
		},
	}

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func (f TieredStorageOutputAddFlags) successString() string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, `tiered storage added successfully, "%s"="%s".`, f.key, f.tier)
	_, _ = fmt.Fprintln(&sb)
	return sb.String()
}

func NewTieredStorageAddCommand() *cobra.Command {
	flags := TieredStorageOutputAddFlags{}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "add a tiered storage",
		Long:  `add a tiered storage`,
		Example: `add 
## Add a key and name of the tiered storage you'd like to add
--key some_key 
--tier some_tier
## Add the capacity, retention period, and format of the tiered storage
--capacity some_capacity 
--retention-period some_retention_period 
--format some_format 
## Add the description of the tiered storage
--description this is a description 
## Add the pipelines in the tiered storage
--pipelines some_pipelines 
--location some_location`,

		/*PreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.MarkFlagsRequiredTogether("key", "tier", "capacity", "retention-period", "format", "pipelines", "location")
			return nil
		},*/
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get the flag values
			keyFlag, _ := cmd.Flags().GetString("key")
			tierFlag, _ := cmd.Flags().GetString("tier")
			capacityFlag, _ := cmd.Flags().GetString("capacity")
			retentionPeriodFlag, _ := cmd.Flags().GetString("retention-period")
			formatFlag, _ := cmd.Flags().GetString("format")
			descriptionFlag, _ := cmd.Flags().GetString("description")
			pipelinesFlag, _ := cmd.Flags().GetStringSlice("pipelines")
			locationFlag, _ := cmd.Flags().GetString("location")

			// Use the flags in the form or fallback to form input
			var (
				key             = keyFlag
				tier            = tierFlag
				capacity        = capacityFlag
				retentionPeriod = retentionPeriodFlag
				format          = formatFlag
				description     = descriptionFlag
				pipelines       = pipelinesFlag
				location        = locationFlag
			)
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Name as key for storage tier (ex. some_tier)").
						Value(&key).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Key cannot be empty")
							}
							return nil
						}),
					huh.NewInput().
						Title("Tier of storage").
						Value(&tier).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Tier cannot be empty")
							}
							return nil
						}),
				),

				huh.NewGroup(
					huh.NewInput().
						Title("Capacity of storage tier").
						Value(&capacity).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Capacity cannot be empty")
							}
							return nil
						}),
					huh.NewInput().
						Title("Retention period of storage tier").
						Value(&retentionPeriod).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Retention period cannot be empty")
							}
							return nil
						}),
				),

				huh.NewGroup(
					huh.NewInput().
						Title("Format for storage tier (ex. iceberg)").
						Value(&format).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Format cannot be empty")
							}
							return nil
						}),
					huh.NewText().
						Title("Description (optional)").
						Value(&description).
						Validate(func(str string) error {
							return nil
						}),
				),

				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Pipelines").
						Options(
							huh.NewOption("traces", "traces"),
							huh.NewOption("metrics", "metrics"),
							huh.NewOption("Logs", "logs").Selected(true),
						).
						Limit(3).
						Value(&pipelines),
					huh.NewInput().
						Title("Location of storage tier").
						Value(&location).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Location cannot be empty")
							}
							return nil
						}),
				),
			) // This is where the form closes

			err := form.Run()
			if err != nil {
				fmt.Println("Storage tier failed due to", err)
				os.Exit(1)
			}

			// Map data to ConfigMap
			configMapData := map[string]string{
				"key":             key,
				"tier":            tier,
				"capacity":        capacity,
				"retentionPeriod": retentionPeriod,
				"format":          format,
				"description":     description,
				"pipelines":       strings.Join(pipelines, ", "),
				"location":        location,
			}

			ctx := cmd.Context()
			helper, err := kubehelper.New(kubehelper.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("failed to create kube helper: %w", err)
			}

			if _, err := helper.UpdateConfigMap(cmd.Context(), "dynamic-variables", "mdai", configMapData); err != nil {
				return fmt.Errorf("failed to add dynamic variable: %w", err)
			}

			fmt.Println(flags.successString())
			fmt.Printf("Key: %s\nTier: %s\nCapacity: %s\nRetention Period: %s\nFormat: %s\nDescription: %s\nPipelines: %v\nLocation: %s\n",
				key, tier, capacity, retentionPeriod, format, description, pipelines, location)

			return nil
		},
	}

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewTieredStorageRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "remove a tiered storage",
		Long:    `remove a tiered storage`,
		Example: `  remove --key some_key`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			TieredStorageOutputKey, _ := cmd.Flags().GetString("key")
			ctx := cmd.Context()
			helper, err := kubehelper.New(kubehelper.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("failed to initialize kubehelper: %w", err)
			}
			configMap, err := helper.GetConfigMap(ctx, "tiered-storage", "mdai")
			if err != nil {
				return fmt.Errorf("failed to fetch tiered storages configmap: %w", err)
			}
			if _, found := configMap.Data[TieredStorageOutputKey]; !found {
				return fmt.Errorf("tiered storage %s not found in configmap", TieredStorageOutputKey)
			}
			delete(configMap.Data, TieredStorageOutputKey)
			if _, err = helper.SetConfigMap(ctx, "mdai", configMap); err != nil {
				return fmt.Errorf("failed to remove tiered storage: %w", err)
			}

			fmt.Printf(`"%s"tiered storage removed successfully.`, TieredStorageOutputKey)
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringP("key", "k", "", "name of the tiered storage")
	_ = cmd.MarkFlagRequired("key")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
