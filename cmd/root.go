package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const mdaiLogo = `
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
  
`

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
	if err := rootCmd.ExecuteContext(rootCmd.Context()); err != nil {
		os.Exit(1)
	}
}

func NewRootCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "mdai",
		Short: "MyDecisive.ai CLI",
		Long:  mdaiLogo,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			kubeconfig := viper.GetString("kubeconfig")
			kubecontext := viper.GetString("kubecontext")

			if kubeconfig == "" {
				if home := homedir.HomeDir(); home != "" {
					kubeconfig = filepath.Join(home, ".kube", "config")
				}
			}
			apiConfig, err := clientcmd.LoadFromFile(kubeconfig)
			if err != nil {
				return fmt.Errorf("error loading kubeconfig: %w", err)
			}
			if kubecontext == "" {
				kubecontext = apiConfig.CurrentContext
			}
			if _, exists := apiConfig.Contexts[kubecontext]; !exists {
				return fmt.Errorf("context '%s' does not exist in kubeconfig `%s`", kubecontext, kubeconfig)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, mdaitypes.Kubeconfig{}, kubeconfig)
			ctx = context.WithValue(ctx, mdaitypes.Kubecontext{}, kubecontext)
			cmd.SetContext(ctx)
			return nil
		},
		Version: fmt.Sprintf("version: %s (git sha: %s), built: %s", Version, GitSha, BuildTime),
	}

	addGroups(cmd)
	addCommands(cmd)

	viper.AutomaticEnv()

	cmd.PersistentFlags().String("kubeconfig", "", "Path to a kubeconfig")
	_ = viper.BindPFlag("kubeconfig", cmd.PersistentFlags().Lookup("kubeconfig"))
	cmd.PersistentFlags().String("kubecontext", "", "Kubernetes context to use")
	_ = viper.BindPFlag("kubecontext", cmd.PersistentFlags().Lookup("kubecontext"))

	cmd.SilenceUsage = true
	cmd.DisableFlagsInUseLine = true

	err := cmd.RegisterFlagCompletionFunc("kubecontext", kubecontextFlagCompletionFunc)

	return cmd, err
}

func addGroups(cmd *cobra.Command) {
	cmd.AddGroup(
		&cobra.Group{ID: "installation", Title: "Installation"},
		&cobra.Group{ID: "configuration", Title: "Configuration"},
	)
}

func addCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		NewConfigureCommand(),
		NewCreateCommand(),
		NewDeleteCommand(),
		NewDisableCommand(),
		NewDocsCommand(),
		NewEnableCommand(),
		NewFilterCommand(),
		NewGetCommand(),
		NewInstallCommand(),
		NewOutdatedCommand(),
		NewRemoveCommand(),
		NewStatusCommand(),
		NewUninstallCommand(),
		NewUpdateCommand(),
		// THESE ARE A FAKE COMMAND
		NewDynamicVariablesCommand(),
		NewTieredStorageCommand(),
	)
}

func kubecontextFlagCompletionFunc(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: cmd.Context().Value(mdaitypes.Kubeconfig{}).(string)}
	if config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).RawConfig(); err == nil {
		var completions []string
		for name, ctx := range config.Contexts {
			completion := name
			if name != ctx.Cluster {
				completion += " (" + ctx.Cluster + ")"
			}
			completions = append(completions, completion)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
