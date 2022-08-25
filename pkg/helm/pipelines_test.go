package helm

import (
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/google/go-cmp/cmp"

	"github.com/bigkevmcd/peanut-helmpipelines/test"
)

func TestHelmChartPipelines(t *testing.T) {
	pipelinesTests := []struct {
		name  string
		items []helmv2.HelmRelease
		want  []HelmReleasePipeline
	}{
		{
			name:  "no helm releases",
			items: []helmv2.HelmRelease{},
			want:  []HelmReleasePipeline{},
		},
		{
			name:  "helm releases without pipelines",
			items: []helmv2.HelmRelease{test.NewHelmRelease(), test.NewHelmRelease()},
			want:  []HelmReleasePipeline{},
		},
		{
			name:  "single helm release in stage",
			items: []helmv2.HelmRelease{test.NewHelmRelease(test.InPipeline("demo-pipeline", "staging", ""))},
			want: []HelmReleasePipeline{
				{
					Name: "demo-pipeline",
					Environments: []HelmReleaseEnvironment{
						{
							Name: "staging",
							Charts: []HelmReleaseChart{
								{
									Name:    "redis",
									Version: "1.0.9",
									Source:  sourceRef("HelmRepository", "default", "test-repository"),
								},
							},
						},
					},
					ChartHelmReleases: map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference{
						{Name: "redis", Version: "1.0.9", Source: helmv2.CrossNamespaceObjectReference{Kind: "HelmRepository", Name: "test-repository", Namespace: "default"}}: {
							{Name: "test-release", Namespace: "default", Kind: "HelmRelease", APIVersion: "source.toolkit.fluxcd.io/v1beta2"}},
					},
				},
			},
		},
		{
			name: "helm releases in two stages of the same pipeline",
			items: []helmv2.HelmRelease{
				test.NewHelmRelease(test.InPipeline("demo-pipeline", "staging", ""), test.Named("staging-deploy", "staging")),
				test.NewHelmRelease(test.InPipeline("demo-pipeline", "production", "staging"), test.Named("production-deploy", "production")),
			},
			want: []HelmReleasePipeline{
				{
					Name: "demo-pipeline",
					Environments: []HelmReleaseEnvironment{
						{
							Name: "staging",
							Charts: []HelmReleaseChart{
								{
									Name:    "redis",
									Version: "1.0.9",
									Source:  sourceRef("HelmRepository", "default", "test-repository"),
								},
							},
						},
						{
							Name: "production",
							Charts: []HelmReleaseChart{
								{
									Name:    "redis",
									Version: "1.0.9",
									Source:  sourceRef("HelmRepository", "default", "test-repository"),
								},
							},
						},
					},
					ChartHelmReleases: map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference{
						{Name: "redis", Version: "1.0.9", Source: helmv2.CrossNamespaceObjectReference{Kind: "HelmRepository", Name: "test-repository", Namespace: "default"}}: {
							{Name: "staging-deploy", Namespace: "staging", Kind: "HelmRelease", APIVersion: "source.toolkit.fluxcd.io/v1beta2"},
							{Name: "production-deploy", Namespace: "production", Kind: "HelmRelease", APIVersion: "source.toolkit.fluxcd.io/v1beta2"}},
					},
				},
			},
		},
		{
			name: "helm releases in the same stage of the same pipeline",
			items: []helmv2.HelmRelease{
				test.NewHelmRelease(test.InPipeline("demo-pipeline", "staging", ""), test.Named("demo1", "test-ns1")),
				test.NewHelmRelease(test.InPipeline("demo-pipeline", "staging", ""), test.Named("demo2", "test-ns2")),
			},
			want: []HelmReleasePipeline{
				{
					Name: "demo-pipeline",
					Environments: []HelmReleaseEnvironment{
						{
							Name: "staging",
							Charts: []HelmReleaseChart{
								{
									Name:    "redis",
									Version: "1.0.9",
									Source:  sourceRef("HelmRepository", "default", "test-repository"),
								},
							},
						},
					},
					ChartHelmReleases: map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference{
						{Name: "redis", Version: "1.0.9", Source: helmv2.CrossNamespaceObjectReference{Kind: "HelmRepository", Name: "test-repository", Namespace: "default"}}: []helmv2.CrossNamespaceObjectReference{
							{Name: "demo1", Namespace: "test-ns1", Kind: "HelmRelease", APIVersion: "source.toolkit.fluxcd.io/v1beta2"},
							{Name: "demo2", Namespace: "test-ns2", Kind: "HelmRelease", APIVersion: "source.toolkit.fluxcd.io/v1beta2"}},
					},
				},
			},
		},
	}

	for _, tt := range pipelinesTests {
		t.Run(tt.name, func(t *testing.T) {
			ps, err := ParseHelmReleasePipelines(tt.items)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, ps); diff != "" {
				t.Fatalf("failed to parse pipelines:\n%s", diff)
			}
		})
	}
}

func sourceRef(kind, namespace, name string) helmv2.CrossNamespaceObjectReference {
	return helmv2.CrossNamespaceObjectReference{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
	}
}
