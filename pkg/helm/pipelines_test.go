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
				},
			},
		},
		{
			name: "helm releases in two stages of the same pipeline",
			items: []helmv2.HelmRelease{
				test.NewHelmRelease(test.InPipeline("demo-pipeline", "staging", "")),
				test.NewHelmRelease(test.InPipeline("demo-pipeline", "production", "staging")),
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
				},
			},
		},
		{
			name: "helm releases in the same stage of the same pipeline",
			items: []helmv2.HelmRelease{
				test.NewHelmRelease(test.InPipeline("demo-pipeline", "staging", "")),
				test.NewHelmRelease(test.InPipeline("demo-pipeline", "staging", "")),
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
				},
			},
		},
	}

	for _, tt := range pipelinesTests {
		t.Run(tt.name, func(t *testing.T) {

			l := &helmv2.HelmReleaseList{
				Items: tt.items,
			}

			ps, err := ParseHelmReleasePipelines(l)
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
