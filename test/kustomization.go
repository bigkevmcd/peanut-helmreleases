package test

import (
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewKustomization creates test Kustomization resources.
func NewKustomization(opts ...func(client.Object)) kustomizev1.Kustomization {
	hr := kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-kustomization",
			Namespace: "default",
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{Duration: time.Minute},
			Path:     "./testing",
		},
	}
	for _, o := range opts {
		o(&hr)
	}
	return hr
}

// Source sets the source on a Kustomization.
func Source(name, namespace string) func(client.Object) {
	return func(k client.Object) {
		kz := k.(*kustomizev1.Kustomization)
		kz.Spec.SourceRef = kustomizev1.CrossNamespaceSourceReference{
			Kind:      "GitRepository",
			Name:      name,
			Namespace: namespace,
		}
	}
}

// Path sets the path on a Kustomization.
func Path(path string) func(client.Object) {
	return func(k client.Object) {
		kz := k.(*kustomizev1.Kustomization)
		kz.Spec.Path = path
	}
}
