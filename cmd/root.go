package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "development"
	GitSha    = "development"
	BuildTime = "development"
)

func Execute() {
	rootCmd, err := NewRootCommand()
	if err != nil {
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "mdai",
		Short: "MyDecisive.ai CLI",
		Long: `
                  -*#%%#*-                
                .#%%%%%%%%#.              
                *%%%%%%%%%%#              
                %%%%%%%%%%%%              
                %%%%%%%%%%%%              
                #%%%%%%%%%%#     .==.     
       .==.     -#=*%%%%#=*=     +%%+     
       +%%+     - ^ %%%% ^ -     +%%+ .%%#
       +%%+  =+- :%%%%%%%%= .-:  +%%+ :%%%
       +%%+ :%%% .%%%%%%%%: %%%: +%%+ :%%%
  +#*  +%%+ :%%% .%%%: %%%: %%%: +%%+  -=:
  %%%: +%%+ :%%% .%%%. %%%: %%%: +%%+     
  %%%: +%%+ .%%% .%%%. %%%- *%#. +%%+     
  -*+  +%%+  :-. .%%%. %%%-       --      
        ::       .%%%. #%%-               
                 .%%%. =%%%-              
                 *%%*   -%%%%%#=          
            :+*#%%%*      -+*#*=          
                
              🐙 MyDecisive.ai  
  
    `,
		RunE: func(cmd *cobra.Command, args []string) error {
			showVersion, _ := cmd.Flags().GetBool("version")
			if showVersion {
				fmt.Printf("version: %s (git sha: %s), built: %s\n", Version, GitSha, BuildTime)
			}
			return nil
		},
	}

	cmd.AddGroup(
		&cobra.Group{ID: "installation", Title: "Installation"},
		&cobra.Group{ID: "configuration", Title: "Configuration"},
	)
	cmd.AddCommand(
		NewConfigureCommand(),
		NewCreateCommand(),
		NewDeleteCommand(),
		NewDemoCommand(),
		NewDisableCommand(),
		NewDocsCommand(),
		NewEnableCommand(),
		NewGetCommand(),
		NewInstallCommand(),
		NewMuteCommand(),
		NewRemoveCommand(),
		NewStatusCommand(),
		NewUninstallCommand(),
		NewUnmuteCommand(),
		NewUpdateCommand(),
	)

	cmd.Flags().Bool("version", false, "Print version information")

	return cmd, nil
}
