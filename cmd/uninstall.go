package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kubehelper"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/spf13/cobra"
)

func NewUninstallCommand() *cobra.Command {
	flags := uninstallFlags{}
	cmd := &cobra.Command{
		GroupID: "installation",
		Use:     "uninstall",
		Short:   "uninstall MyDecisive Cluster",
		Long:    "uninstall MyDecisive Cluster",
		Example: `  mdai uninstall --kubecontext kind-mdai-local # uninstall from kind cluster mdai-local
  mdai uninstall --debug                   # uninstall in debug mode
  mdai uninstall --quiet                   # uninstall in quiet mode
  mdai uninstall --confirm                 # uninstall, with confirmation`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			if !flags.confirm {
				kubeconfig := ctx.Value(mdaitypes.Kubeconfig{}).(string)
				kubecontext := ctx.Value(mdaitypes.Kubecontext{}).(string)
				if err := huh.NewConfirm().
					Title("Uninstall MDAI from this cluster?").
					Description(fmt.Sprintf("kubeconfig: %s\nkubecontext: %s\n", kubeconfig, kubecontext)).
					Negative("No!").
					Affirmative("Yes.").
					Value(&flags.confirm).Run(); err != nil {
					return fmt.Errorf("uninstall failed: %w", err)
				}
			}
			if !flags.confirm {
				return errors.New("aborting uninstallation")
			}
			logger := log.New(os.Stderr)
			if flags.debug {
				logger.SetLevel(log.DebugLevel)
			}
			ctx = log.WithContext(ctx, logger)
			return mdaiUninstall(ctx)
		},
	}
	cmd.Flags().BoolVar(&flags.debug, "debug", false, "debug mode")
	cmd.Flags().BoolVar(&flags.quiet, "quiet", false, "quiet mode")
	cmd.Flags().BoolVar(&flags.confirm, "confirm", false, "confirm uninstallation")

	cmd.MarkFlagsMutuallyExclusive("debug", "quiet")

	return cmd
}

func mdaiUninstall(ctx context.Context) error {
	spinnerCtx, cancel := context.WithCancelCause(ctx)

	go func() {
		opts := []mdaihelm.ClientOption{mdaihelm.WithContext(ctx)}
		helmclient := mdaihelm.NewClient(opts...)
		cancel(helmclient.UninstallChart("mdai-cluster"))
	}()

	if err := spinner.New().Title("uninstalling MDAI Cluster 🐙").
		// TitleStyle(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00020A", Dark: "#D3D3D3"})).
		// Style(lipgloss.NewStyle().PaddingLeft(1).Foreground(purple)).
		Context(spinnerCtx).
		Run(); err != nil {
		return fmt.Errorf("failed to install cluster: %w", err)
	}

	if spinnerCtx.Err() != nil && !errors.Is(context.Cause(spinnerCtx), context.Canceled) {
		fmt.Println(lipgloss.NewStyle().PaddingLeft(1).Foreground(red).Render(DisabledString) + " uninstalling MDAI Cluster 🐙")
		return fmt.Errorf("failed to install cluster: %w", context.Cause(spinnerCtx))
	}

	fmt.Println(lipgloss.NewStyle().PaddingLeft(1).Foreground(green).Render(EnabledString) + " uninstalling MDAI Cluster 🐙")
	helper, err := kubehelper.New(kubehelper.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to initialize kubehelper: %w", err)
	}

	for _, crd := range customResourceDefinitions() {
		if err = helper.DeleteCRD(ctx, crd); err != nil {
			fmt.Println("\tCRD " + crd + " not found, skipping deletion.")
			continue
		}
		fmt.Println("\tCRD " + crd + " deleted successfully.")
	}
	fmt.Println(" 🙁 Sad to see you go.")
	return nil
}
