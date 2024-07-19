package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kind"
	"github.com/decisiveai/mdai-cli/internal/viewport"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

//go:embed templates/*
var embedFS embed.FS

var installationType string

func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "installation",
		Use:     "install [--cluster-name CLUSTER-NAME] [--debug] [--quiet]",
		Short:   "install MyDecisive Cluster",
		Long:    "install MyDecisive Cluster",
		Example: `  mdai install --cluster-name mdai-local # install locally on kind cluster mdai-local
  mdai install --debug                   # install in debug mode
  mdai install --quiet                   # install in quiet mode`,
		Args: cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			/*
				aws, _ := cmd.Flags().GetBool("aws")
				local, _ := cmd.Flags().GetBool("local")
				if aws {
					installationType = "aws"
				}
				if local {
					installationType = "kind"
				}
			*/
			installationType = "kind"

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			messages := make(chan string)
			debug := make(chan string)
			errs := make(chan error)
			done := make(chan struct{})
			task := make(chan string)
			defer func() {
				close(messages)
				close(debug)
				close(errs)
				close(task)
				close(done)
			}()

			debugMode, _ := cmd.Flags().GetBool("debug")
			quietMode, _ := cmd.Flags().GetBool("quiet")
			clusterName, _ := cmd.Flags().GetString("cluster-name")

			/*
				if installationType == "" {
					s := huh.NewSelect[string]().
						Title("Installation Type").
						Options(
							huh.NewOption("Local Installation via kind", "kind"),
							huh.NewOption("AWS Installation via eks", "aws"),
						).
						Value(&installationType)

					huh.NewForm(huh.NewGroup(s)).Run()
				}
			*/

			go func() {
				switch installationType {
				case "kind":
					task <- "creating kubernetes cluster via kind"
					kindclient := kind.NewClient(messages, debug, errs, clusterName)
					if _, err := kindclient.Install(); err != nil {
						errs <- fmt.Errorf("failed to create kubernetes cluster: %w", err)
						return
					}
				}

				tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
				if err != nil {
					errs <- fmt.Errorf("failed to create temp dir: %w", err)
					return
				}
				defer os.Remove(tmpfile.Name())
				helmclient := mdaihelm.NewClient(messages, debug, errs, tmpfile.Name())
				task <- "adding helm repos"
				if err := helmclient.AddRepos(); err != nil {
					errs <- fmt.Errorf("failed to add helm repos: %w", err)
					return
				}
				for _, helmchart := range mdaiHelmcharts {
					task <- "installing helm chart " + helmchart
					if err := helmclient.InstallChart(helmchart); err != nil {
						errs <- fmt.Errorf("failed to install helm chart %s: %w", helmchart, err)
						return
					}
				}

				cfg, err := config.GetConfig()
				if err != nil {
					errs <- fmt.Errorf("failed to get kubernetes config: %w", err)
					return
				}

				dynamicClient, err := dynamic.NewForConfig(cfg)
				if err != nil {
					errs <- fmt.Errorf("failed to create dynamic client: %w", err)
					return
				}

				obj := &unstructured.Unstructured{}
				decoder := scheme.Codecs.UniversalDecoder()
				manifest, _ := embedFS.ReadFile("templates/mdai-operator.yaml")
				if _, _, err = decoder.Decode(manifest, nil, obj); err != nil {
					errs <- fmt.Errorf("failed to decode mdai-operator manifest: %w", err)
					return
				}

				mdaiOperator, err := dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Get(
					context.TODO(),
					obj.GetName(),
					metav1.GetOptions{},
				)
				// if err != nil && err.Error() != fmt.Sprintf(`%s.%s "%s" not found`, mdaitypes.MDAIOperatorResource, mdaitypes.MDAIOperatorGroup, obj.GetName()) {
				if err != nil {
					if !k8serrors.IsNotFound(err) {
						errs <- fmt.Errorf("failed to get mdai-operator: %w", err)
						return
					}
					task <- "applying mdai-operator manifest"
					if _, err = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(
						context.TODO(),
						obj,
						metav1.CreateOptions{},
					); err != nil {
						errs <- fmt.Errorf("failed to apply mdai-operator manifest: %w", err)
						return
					}
					messages <- "mdai-operator manifest applied successfully"
				} else {
					task <- "updating mdai-operator manifest"
					obj.SetResourceVersion(mdaiOperator.GetResourceVersion())
					if _, err = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Update(
						context.TODO(),
						obj,
						metav1.UpdateOptions{},
					); err != nil {
						errs <- fmt.Errorf("failed to update mdai-operator manifest: %w", err)
						return
					}
					messages <- "mdai-operator manifest updated successfully"
				}

				messages <- "installation completed successfully"
				done <- struct{}{}
			}()

			p := tea.NewProgram(
				viewport.InitialModel(
					messages,
					debug,
					errs,
					done,
					task,
					debugMode,
					quietMode,
				),
			)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("failed to run program: %w", err)
			}

			return nil
		},
	}
	// cmd.Flags().Bool("aws", false, "aws installation type")
	// cmd.Flags().Bool("local", false, "local installation type")
	cmd.Flags().String("cluster-name", "mdai-local", "kubernetes cluster name")
	cmd.Flags().Bool("debug", false, "debug mode")
	cmd.Flags().Bool("quiet", false, "quiet mode")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
