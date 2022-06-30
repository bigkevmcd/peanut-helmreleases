package test

import (
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/gitops-tools/apps-scanner/pkg/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewHelmRelease creates test HelmRelease resources.
func NewHelmRelease(opts ...func(*helmv2.HelmRelease)) helmv2.HelmRelease {
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

// InPipeline is an option for NewHelmRelease that applies the correct labels to
// indicate that the HelmRelease is in a pipeline.
func InPipeline(name, env, after string) func(*helmv2.HelmRelease) {
	return func(hr *helmv2.HelmRelease) {
		lbls := hr.GetLabels()
		if lbls == nil {
			lbls = map[string]string{}
		}
		lbls[pipelines.PipelineNameLabel] = name
		lbls[pipelines.PipelineEnvironmentLabel] = env
		lbls[pipelines.PipelineEnvironmentAfterLabel] = after
		hr.SetLabels(lbls)
	}
}
