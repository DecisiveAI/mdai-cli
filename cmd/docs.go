package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "generate documentation",
	Long:  ``,
	Run: func(cmd *cobra.Command, _ []string) {
		md, _ := cmd.Flags().GetBool("md")
		yaml, _ := cmd.Flags().GetBool("yaml")
		rst, _ := cmd.Flags().GetBool("rst")
		man, _ := cmd.Flags().GetBool("man")

		if md {
			doc.GenMarkdownTree(rootCmd, "docs/md")
		}

		if yaml {
			doc.GenYamlTree(rootCmd, "docs/yaml")
		}

		if rst {
			doc.GenReSTTree(rootCmd, "docs/rst")
		}

		if man {
			header := &doc.GenManHeader{
				Title:   "MDAI CLI",
				Section: "1",
			}
			doc.GenManTree(rootCmd, header, "docs/man")
		}
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().Bool("yaml", false, "generate YAML documentation")
	docsCmd.Flags().Bool("markdown", true, "generate Markdown documentation")
	docsCmd.Flags().Bool("restructured", false, "generate ReStructuredText documentation")
	docsCmd.Flags().Bool("man", false, "generate man page documentation")

	docsCmd.Flags().SetNormalizeFunc(func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "md":
			name = "markdown"
		case "rst":
			name = "restructured"
		}
		return pflag.NormalizedName(name)
	})

	docsCmd.Hidden = true
}
