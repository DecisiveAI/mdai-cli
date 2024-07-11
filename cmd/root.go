package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var mdaiHelmcharts = []string{"cert-manager", "opentelemetry-operator", "prometheus", "mdai-api", "mdai-console", "datalyzer", "mdai-operator"}

var rootCmd = &cobra.Command{
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddGroup(
		&cobra.Group{ID: "installation", Title: "Installation"},
		&cobra.Group{ID: "configuration", Title: "Configuration"},
	)
}
