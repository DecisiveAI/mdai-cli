package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	"github.com/decisiveai/mdai-cli/internal/kind"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/decisiveai/mdai-cli/internal/viewport"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

//go:embed templates/*
var embedFS embed.FS

var installationType string

var installCmd = &cobra.Command{
	GroupID: "installation",
	Use:     "install [--cluster-name CLUSTER-NAME] [--debug] [--quiet]",
	Short:   "install MyDecisive Cluster",
	Long:    "install MyDecisive Cluster",
	Example: `  mdai install --cluster-name mdai-local # install locally on kind cluster mdai-local
  mdai install --debug                   # install in debug mode
  mdai install --quiet                   # install in quiet mode`,
	PreRun: func(_ *cobra.Command, _ []string) {
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
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		messages := make(chan string)
		debug := make(chan string)
		errs := make(chan error)
		done := make(chan bool)
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

		go func() error {
			switch installationType {
			case "kind":
				task <- "creating kubernetes cluster via kind"
				kindclient := kind.NewClient(messages, debug, errs, clusterName)
				if _, err := kindclient.Install(); err != nil {
					errs <- fmt.Errorf("failed to create kubernetes cluster: %w", err)
					return fmt.Errorf("failed to create kubernetes cluster: %w", err)
				}
			}

			tmpfile, err := os.CreateTemp(os.TempDir(), "mdai-cli")
			if err != nil {
				errs <- fmt.Errorf("failed to create temp dir: %w", err)
				return fmt.Errorf("failed to create temp dir: %w", err)
			}
			defer os.Remove(tmpfile.Name())
			helmclient := mdaihelm.NewClient(messages, debug, errs, tmpfile.Name())
			task <- "adding helm repos"
			if err := helmclient.AddRepos(); err != nil {
				errs <- fmt.Errorf("failed to add helm repos: %w", err)
				return fmt.Errorf("failed to add helm repos: %w", err)
			}
			for _, helmchart := range mdaiHelmcharts {
				task <- "installing helm chart " + helmchart
				if err := helmclient.InstallChart(helmchart); err != nil {
					errs <- fmt.Errorf("failed to install helm chart %s: %w", helmchart, err)
					return fmt.Errorf("failed to install helm chart %s: %w", helmchart, err)
				}
			}

			cfg, err := config.GetConfig()
			if err != nil {
				errs <- fmt.Errorf("failed to get kubernetes config: %w", err)
				return fmt.Errorf("failed to get kubernetes config: %w", err)
			}

			dynamicClient, err := dynamic.NewForConfig(cfg)
			if err != nil {
				errs <- fmt.Errorf("failed to create dynamic client: %w", err)
				return fmt.Errorf("failed to create dynamic client: %w", err)
			}

			gvr := schema.GroupVersionResource{
				Group:    mdaitypes.MDAIOperatorGroup,
				Version:  mdaitypes.MDAIOperatorVersion,
				Resource: mdaitypes.MDAIOperatorResource,
			}

			obj := &unstructured.Unstructured{}
			decoder := scheme.Codecs.UniversalDecoder()
			manifest, _ := embedFS.ReadFile("templates/mdai-operator.yaml")
			_, _, err = decoder.Decode(manifest, nil, obj)
			if err != nil {
				errs <- fmt.Errorf("failed to decode mdai-operator manifest: %w", err)
				return fmt.Errorf("failed to decode mdai-operator manifest: %w", err)
			}

			mdaiOperator, err := dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Get(
				context.TODO(),
				obj.GetName(),
				metav1.GetOptions{},
			)
			if err != nil && err.Error() != fmt.Sprintf(`%s.%s "%s" not found`, mdaitypes.MDAIOperatorResource, mdaitypes.MDAIOperatorGroup, obj.GetName()) {
				errs <- fmt.Errorf("failed to get mdai-operator: %w", err)
				return fmt.Errorf("failed to get mdai-operator: %w", err)
			}

			if mdaiOperator == nil {
				task <- "applying mdai-operator manifest"
				if _, err = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(
					context.TODO(),
					obj,
					metav1.CreateOptions{},
				); err != nil {
					errs <- fmt.Errorf("failed to apply mdai-operator manifest: %w", err)
					return fmt.Errorf("failed to apply mdai-operator manifest: %w", err)
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
					return fmt.Errorf("failed to update mdai-operator manifest: %w", err)
				}
				messages <- "mdai-operator manifest updated successfully"
			}
			messages <- "installation completed successfully"
			done <- true
			return nil
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

func init() {
	rootCmd.AddCommand(installCmd)
	// installCommand.Flags().Bool("aws", false, "aws installation type")
	// installCommand.Flags().Bool("local", false, "local installation type")
	installCmd.Flags().String("cluster-name", "mdai-local", "kubernetes cluster name")
	installCmd.Flags().Bool("debug", false, "debug mode")
	installCmd.Flags().Bool("quiet", false, "quiet mode")
	installCmd.DisableFlagsInUseLine = true
}
