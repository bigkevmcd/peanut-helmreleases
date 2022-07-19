package test

import (
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewHelmRelease creates test HelmRelease resources.
func NewHelmRelease(opts ...func(client.Object)) helmv2.HelmRelease {
	hr := helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-release",
			Namespace: "default",
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{Duration: time.Minute},
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   "redis",
					Version: "1.0.9",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Name:      "test-repository",
						Kind:      "HelmRepository",
						Namespace: "default",
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(&hr)
	}
	return hr
}
