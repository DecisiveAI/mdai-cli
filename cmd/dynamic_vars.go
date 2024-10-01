package cmd

import (
	"errors"
	"fmt"
	"strings"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func NewDynamicVariablesCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "dynamic_variables",
		Short:   "manage dynamic variables",
		Long:    `manage dynamic variables`,
	}

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	cmd.AddCommand(
		NewDynamicVariablesAddCommand(),
		NewDynamicVariablesListCommand(),
		NewDynamicVariablesRemoveCommand(),
	)

	return cmd
}

// Define DynamicVariables as a list of key-value pairs (map[string]string)
type DynamicVariables []map[string]string

func GetMockDynamicVars() DynamicVariables {
	// Initialize DynamicVariables with some key-value pairs
	vars := DynamicVariables{
		{"key": "environment", "value": "production"},
		{"key": "version", "value": "1.0.0"},
		{"key": "region", "value": "us-west-1"},
	}
	return vars
}

func NewDynamicVariablesListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list dynamic variables",
		Long:  `list dynamic variables`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dynamicVars := GetMockDynamicVars()

			var rows [][]string
			headers := []string{"Key", "Value", "Status"}

			for _, dynamicVar := range dynamicVars {
				row := []string{dynamicVar["key"], dynamicVar["value"], "Enabled"}
				rows = append(rows, row)
			}

			printDynamicVariables := func(headers []string, rows [][]string) {
				if len(rows) == 0 {
					return
				}
				dynamicVariablesOutput := table.New().
					BorderHeader(false).
					Border(lipgloss.HiddenBorder()).
					StyleFunc(func(row, col int) lipgloss.Style {
						switch {
						case row == 0:
							return HeaderStyle
						default:
							return lipgloss.NewStyle()
						}
					}).
					Headers(headers...).
					Rows(rows...)
				fmt.Println(dynamicVariablesOutput)
			}

			printDynamicVariables(headers, rows)
			return nil
		},
	}

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

type dynamicVariableAddFlags struct {
	key   string
	value string
}


func (f dynamicVariableAddFlags) successString() string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, `dynamic variable added successfully as: "%s" (%s).`, f.key, f.value)
	_, _ = fmt.Fprintln(&sb)
	return sb.String()
}

func NewDynamicVariablesAddCommand() *cobra.Command {
	f := dynamicVariableAddFlags{}

	cmd := &cobra.Command{
		Use:     "add",
		Short:   "add a dynamic variable",
		Long:    `add a dynamic variable`,
		Example: `  add --key some_key --value s0m3v@lu3@5@`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.MarkFlagsRequiredTogether("key", "value")

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().NFlag() == 0 {
				return errors.New(cmd.UsageString())
			}

			fmt.Println(f.successString())
			return nil
		},
	}

	cmd.Flags().StringVarP(&f.key, "key", "k", "", "key of the dynamic variable")
	cmd.Flags().StringVarP(&f.value, "value", "v", "", "value of the dynamic variable")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewDynamicVariablesRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "remove a dynamic variable",
		Long:    `remove a dynamic variable`,
		Example: `  remove --key some_key`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dynamicVariableKey, _ := cmd.Flags().GetString("key")

			fmt.Printf(`"%s" dynamic variable removed successfully.`, dynamicVariableKey)
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringP("key", "k", "", "name of the dynamic variable")
	_ = cmd.MarkFlagRequired("key")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
