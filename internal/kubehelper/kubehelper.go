package kubehelper

import (
	"bytes"
	"context"
	"fmt"

	mydecisivev1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	opentelemetry "github.com/decisiveai/opentelemetry-operator/apis/v1alpha1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
	config                 *rest.Config
	apiExtensionsClientset *apiextensionsclient.Clientset
	k8sClient              client.Client
}

func New() (*Helper, error) {
	log.SetLogger(zap.New())
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	s := scheme.Scheme
	if err := mydecisivev1.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("failed to add mydecisivev1 scheme: %w", err)
	}
	if err := opentelemetry.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("failed to add opentelemetry scheme: %w", err)
	}
	k8sClient, err := client.New(cfg, client.Options{Scheme: s})
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}
	apiExtensionsClientset, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create api extensions client: %w", err)
	}

	helper := Helper{
		config:                 cfg,
		apiExtensionsClientset: apiExtensionsClientset,
		k8sClient:              k8sClient,
	}

	return &helper, nil
}

func (helper *Helper) GetOperator(ctx context.Context) (*mydecisivev1.MyDecisiveEngine, error) {
	get := mydecisivev1.MyDecisiveEngine{}
	if err := helper.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      operatorName,
	}, &get); err != nil {
		return nil, fmt.Errorf("failed to get mdai operator: %w", err)
	}
	return &get, nil
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

func getObject(manifest []byte) (*unstructured.Unstructured, error) {
	var decodedObj map[string]interface{}
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 1024)
	if err := decoder.Decode(&decodedObj); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: decodedObj}, nil
}
