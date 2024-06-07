package oteloperator

import (
	"context"
	"fmt"

	opentelemetry "github.com/decisiveai/opentelemetry-operator/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func GetConfig() string {
	group := "opentelemetry.io"
	version := "v1alpha1"
	kind := "OpenTelemetryCollector"
	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	s := scheme.Scheme
	opentelemetry.AddToScheme(s)
	list := opentelemetry.OpenTelemetryCollectorList{}
	list.SetGroupVersionKind(gvk)
	cfg := config.GetConfigOrDie()
	k8sClient, _ := client.New(cfg, client.Options{Scheme: s})
	if err := k8sClient.List(context.TODO(), &list); err != nil {
		fmt.Printf("error: %+v\n", err)
		return ""
	}
	return list.Items[0].Spec.Config
}
