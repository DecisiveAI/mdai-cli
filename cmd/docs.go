package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

func NewDocsCommand() *cobra.Command {
	flags := docsFlags{}
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "generate documentation",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			rootCmd, err := NewRootCommand()
			if err != nil {
				return fmt.Errorf("failed to create root command: %w", err)
			}

			if flags.md {
				if err := doc.GenMarkdownTree(rootCmd, "docs/md"); err != nil {
					return fmt.Errorf("failed to generate markdown documentation: %w", err)
				}
			}

			if flags.yaml {
				if err := doc.GenYamlTree(rootCmd, "docs/yaml"); err != nil {
					return fmt.Errorf("failed to generate yaml documentation: %w", err)
				}
			}

			if flags.rst {
				if err := doc.GenReSTTree(rootCmd, "docs/rst"); err != nil {
					return fmt.Errorf("failed to generate rst documentation: %w", err)
				}
			}

			if flags.man {
				header := &doc.GenManHeader{
					Title:   "MDAI CLI",
					Section: "1",
				}
				if err := doc.GenManTree(rootCmd, header, "docs/man"); err != nil {
					return fmt.Errorf("failed to generate man documentation: %w", err)
				}
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&flags.yaml, "yaml", false, "generate YAML documentation")
	cmd.Flags().BoolVar(&flags.md, "markdown", true, "generate Markdown documentation")
	cmd.Flags().BoolVar(&flags.rst, "restructured", false, "generate ReStructuredText documentation")
	cmd.Flags().BoolVar(&flags.man, "man", false, "generate man page documentation")

	cmd.Flags().SetNormalizeFunc(func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "md":
			name = "markdown"
		case "rst":
			name = "restructured"
		}
		return pflag.NormalizedName(name)
	})

	cmd.Hidden = true

	return cmd
}
