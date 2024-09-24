package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/operator"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/spf13/cobra"
)

//go:embed templates/*
var embedFS embed.FS

func NewInstallCommand() *cobra.Command {
	flags := installFlags{}
	cmd := &cobra.Command{
		GroupID: "installation",
		Use:     "install [--cluster-name CLUSTER-NAME] [--debug] [--quiet]",
		Short:   "install MyDecisive Cluster",
		Long:    "install MyDecisive Cluster",
		Example: `  mdai install --kubecontext kind-mdai-local # install on kind cluster mdai-local
  mdai install --debug                   # install in debug mode
  mdai install --quiet                   # install in quiet mode
  mdai install --confirm                 # install, with confirmation`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if !flags.confirm {
				kubeconfig := ctx.Value(mdaitypes.Kubeconfig{}).(string)
				kubecontext := ctx.Value(mdaitypes.Kubecontext{}).(string)
				if err := huh.NewConfirm().
					Title("Install MDAI into this cluster?").
					Description(fmt.Sprintf("kubeconfig: %s\nkubecontext: %s\n", kubeconfig, kubecontext)).
					Affirmative("Yes!").
					Negative("No.").
					Value(&flags.confirm).Run(); err != nil {
					return fmt.Errorf("install failed: %w", err)
				}
			}
			if !flags.confirm {
				return errors.New("aborting installation")
			}
			logger := log.New(os.Stderr)
			if flags.debug {
				logger.SetLevel(log.DebugLevel)
			}
			ctx = log.WithContext(ctx, logger)
			return mdaiInstall(ctx)
		},
	}
	cmd.Flags().BoolVar(&flags.debug, "debug", false, "debug mode")
	cmd.Flags().BoolVar(&flags.quiet, "quiet", false, "quiet mode")
	cmd.Flags().BoolVar(&flags.confirm, "confirm", false, "confirm installation")

	cmd.MarkFlagsMutuallyExclusive("debug", "quiet")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func mdaiInstall(ctx context.Context) error {
	spinnerCtx, cancel := context.WithCancelCause(ctx)

	go func() {
		opts := []mdaihelm.ClientOption{mdaihelm.WithContext(ctx)}
		helmclient := mdaihelm.NewClient(opts...)
		cancel(helmclient.InstallChart("mdai-cluster"))
	}()
	if err := spinner.New().
		Title("installing MDAI Cluster 🐙").
		//	TitleStyle(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00020A", Dark: "#D3D3D3"})).
		//		Style(lipgloss.NewStyle().PaddingLeft(1).Foreground(purple)).
		// Action(func() { log.FromContext(ctx).Print("NOOP") }).
		Context(spinnerCtx).
		Run(); err != nil {
		return fmt.Errorf("failed to install cluster: %w", err)
	}

	if spinnerCtx.Err() != nil && !errors.Is(context.Cause(spinnerCtx), context.Canceled) {
		fmt.Println(lipgloss.NewStyle().PaddingLeft(1).Foreground(red).Render(DisabledString) + " installing MDAI Cluster 🐙")
		return fmt.Errorf("failed to install cluster: %w", context.Cause(spinnerCtx))
	}

	fmt.Println(lipgloss.NewStyle().PaddingLeft(1).Foreground(green).Render(EnabledString) + " installing MDAI Cluster 🐙")

	manifest, _ := embedFS.ReadFile("templates/mdai-operator.yaml")
	if err := operator.Install(ctx, manifest); err != nil {
		return fmt.Errorf("failed to apply mdai operator manifest: %w", err)
	}

	fmt.Println(" 🍻 You're ready to go")
	return nil
}
