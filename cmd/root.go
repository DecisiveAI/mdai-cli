package cmd

import (
	"os"

	"github.com/spf13/cobra"
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

	return cmd, nil
}
