package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

func NewDocsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "generate documentation",
		Long:  ``,
		Run: func(cmd *cobra.Command, _ []string) {
			md, _ := cmd.Flags().GetBool("md")
			yaml, _ := cmd.Flags().GetBool("yaml")
			rst, _ := cmd.Flags().GetBool("rst")
			man, _ := cmd.Flags().GetBool("man")

			rootCmd, _ := NewRootCommand()

			if md {
				doc.GenMarkdownTree(rootCmd, "docs/md") // nolint: errcheck
			}

			if yaml {
				doc.GenYamlTree(rootCmd, "docs/yaml") // nolint: errcheck
			}

			if rst {
				doc.GenReSTTree(rootCmd, "docs/rst") // nolint: errcheck
			}

			if man {
				header := &doc.GenManHeader{
					Title:   "MDAI CLI",
					Section: "1",
				}
				doc.GenManTree(rootCmd, header, "docs/man") // nolint: errcheck
			}
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
