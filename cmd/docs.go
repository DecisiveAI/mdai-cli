package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

func NewDocsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "generate documentation",
		Long:  ``,
		RunE: func(cmd *cobra.Command, _ []string) error {
			md, _ := cmd.Flags().GetBool("md")
			yaml, _ := cmd.Flags().GetBool("yaml")
			rst, _ := cmd.Flags().GetBool("rst")
			man, _ := cmd.Flags().GetBool("man")

			rootCmd, _ := NewRootCommand()

			if md {
				if err := doc.GenMarkdownTree(rootCmd, "docs/md"); err != nil {
					return fmt.Errorf("failed to generate markdown documentation: %w", err)
				}
			}

			if yaml {
				if err := doc.GenYamlTree(rootCmd, "docs/yaml"); err != nil {
					return fmt.Errorf("failed to generate yaml documentation: %w", err)
				}
			}

			if rst {
				if err := doc.GenReSTTree(rootCmd, "docs/rst"); err != nil {
					return fmt.Errorf("failed to generate rst documentation: %w", err)
				}
			}

			if man {
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
	cmd.Flags().Bool("yaml", false, "generate YAML documentation")
	cmd.Flags().Bool("markdown", true, "generate Markdown documentation")
	cmd.Flags().Bool("restructured", false, "generate ReStructuredText documentation")
	cmd.Flags().Bool("man", false, "generate man page documentation")

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
