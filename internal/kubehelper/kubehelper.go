package kubehelper

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	opentelemetry "github.com/decisiveai/opentelemetry-operator/apis/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	namespace    = "mdai"
	operatorName = "mydecisiveengine-sample-1"
)

var mdaiOperator = mydecisivev1.MyDecisiveEngine{
	ObjectMeta: metav1.ObjectMeta{
		Name:      operatorName,
		Namespace: namespace,
	},
}

type Helper struct {
	kubeconfig             string
	kubecontext            string
	restConfig             *rest.Config
	apiConfig              *api.Config
	apiExtensionsClientset *apiextensionsclient.Clientset
	k8sClient              client.Client
	clientset              *kubernetes.Clientset
}

type HelperOption func(*Helper)

func WithContext(ctx context.Context) HelperOption {
	return func(helper *Helper) {
		if kubeconfig, ok := ctx.Value(mdaitypes.Kubeconfig{}).(string); ok {
			helper.kubeconfig = kubeconfig
		}
		if kubecontext, ok := ctx.Value(mdaitypes.Kubecontext{}).(string); ok {
			helper.kubecontext = kubecontext
		}
	}
}

func New(options ...HelperOption) (*Helper, error) {
	helper := new(Helper)
	for _, option := range options {
		option(helper)
	}
	log.SetLogger(zap.New())
	apiConfig, err := clientcmd.LoadFromFile(helper.kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*apiConfig,
		&clientcmd.ConfigOverrides{
			CurrentContext: helper.kubecontext,
		})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create rest config: %w", err)
	}

	s := scheme.Scheme
	if err := mydecisivev1.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("failed to add mydecisivev1 scheme: %w", err)
	}
	if err := opentelemetry.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("failed to add opentelemetry scheme: %w", err)
	}

	k8sClient, err := client.New(restConfig, client.Options{Scheme: s})
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clientset: %w", err)
	}
	apiExtensionsClientset, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create api extensions client: %w", err)
	}

	helper.restConfig = restConfig
	helper.apiConfig = apiConfig
	helper.apiExtensionsClientset = apiExtensionsClientset
	helper.k8sClient = k8sClient
	helper.clientset = clientset

	return helper, nil
}

func (helper *Helper) GetOperator(ctx context.Context) (*mydecisivev1.MyDecisiveEngine, error) {
	list := mydecisivev1.MyDecisiveEngineList{}
	if err := helper.k8sClient.List(
		ctx,
		&list,
		&client.ListOptions{
			Namespace: namespace,
		},
	); err != nil {
		return nil, fmt.Errorf("failed to get operator list: %w", err)
	}
	if len(list.Items) > 1 {
		operatorNames := make([]string, len(list.Items))
		for i, item := range list.Items {
			operatorNames[i] = item.GetName()
		}
		return nil, fmt.Errorf("more than one mydecisivev1.MyDecisiveEngine found [%s]", strings.Join(operatorNames, ", "))
	}

	return &list.Items[0], nil
}

func (helper *Helper) GetOTELOperator(ctx context.Context) (*opentelemetry.OpenTelemetryCollector, error) {
	get := opentelemetry.OpenTelemetryCollector{}
	if err := helper.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      "gateway",
	}, &get); err != nil {
		return nil, fmt.Errorf("failed to get opentelemetry operator: %w", err)
	}
	return &get, nil
}

func (helper *Helper) GetTelemetryFiltering(ctx context.Context) (*mydecisivev1.TelemetryFilterConfig, error) {
	operator, err := helper.GetOperator(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mdai operator: %w", err)
	}
	return operator.Spec.TelemetryModule.Collectors[0].TelemetryFiltering, nil
}

func (helper *Helper) Patch(ctx context.Context, patchType types.PatchType, patch []byte) error {
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := helper.k8sClient.Patch(
			ctx,
			&mdaiOperator,
			client.RawPatch(patchType, patch),
		)
		return err
	}); err != nil {
		return fmt.Errorf("failed to patch mdai operator: %w", err)
	}
	return nil
}

func (helper *Helper) Apply(ctx context.Context, manifest []byte, gvk *schema.GroupVersionKind) error {
	obj, err := getObject(manifest)
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}

	get := unstructured.Unstructured{}
	get.SetGroupVersionKind(*gvk)
	if err := helper.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, &get); err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("failed to get manifest: %w", err)
		}
		if err := helper.k8sClient.Create(ctx, obj); err != nil {
			return fmt.Errorf("failed to create manifest: %w", err)
		}
		return nil
	}
	obj.SetResourceVersion(get.GetResourceVersion())
	if err := helper.k8sClient.Update(ctx, obj); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

func (helper *Helper) DeleteCRD(ctx context.Context, crd string) error {
	if err := helper.apiExtensionsClientset.ApiextensionsV1().CustomResourceDefinitions().Delete(
		ctx,
		crd,
		metav1.DeleteOptions{},
	); err != nil {
		return fmt.Errorf("failed to delete crd: %w", err)
	}
	return nil
}

func (helper *Helper) GetDeployment(ctx context.Context, deployment, namespace string) (*appsv1.Deployment, error) {
	return helper.clientset.AppsV1().Deployments(namespace).Get(ctx, deployment, metav1.GetOptions{})
}

func (helper *Helper) GetPodByLabel(ctx context.Context, namespace, labelSelector string) (*corev1.PodList, error) {
	return helper.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
}

func getObject(manifest []byte) (*unstructured.Unstructured, error) {
	var decodedObj map[string]interface{}
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 1024) //nolint: mnd
	if err := decoder.Decode(&decodedObj); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: decodedObj}, nil
}
