package kustomizations

import (
	"context"
	"testing"

	"github.com/bigkevmcd/peanut-helmpipelines/test"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestKustomizationPipelines(t *testing.T) {
	pipelinesTests := []struct {
		name           string
		kustomizations []kustomizev1.Kustomization
		objs           []runtime.Object
		want           []KustomizationPipeline
	}{
		{
			name:           "no kustomization",
			kustomizations: []kustomizev1.Kustomization{},
			want:           []KustomizationPipeline{},
		},
		{
			name: "kustomizations without pipelines",
			kustomizations: []kustomizev1.Kustomization{
				test.NewKustomization(test.Named("testing1", "test-ns")),
				test.NewKustomization(test.Named("testing2", "test-ns"))},
			want: []KustomizationPipeline{},
		},
		{
			name: "single kustomization in stage",
			kustomizations: []kustomizev1.Kustomization{
				test.NewKustomization(test.InPipeline("demo-pipeline", "staging", ""), test.Source("test-repo", "test-ns"))},
			objs: []runtime.Object{
				newGitRepository("test-repo", "test-ns", "https://github.com/example/example.git", &sourcev1.GitRepositoryRef{
					Branch: "main",
				}),
			},
			want: []KustomizationPipeline{
				{
					Name: "demo-pipeline",
					Environments: []KustomizationEnvironment{
						{
							Name: "staging",
							Kustomizations: []EnvironmentKustomization{
								{
									Path: "./testing",
									Reference: &sourcev1.GitRepositoryRef{
										Branch: "main",
									},
									URL: "https://github.com/example/example.git",
									Source: kustomizev1.CrossNamespaceSourceReference{
										Kind:      "GitRepository",
										Name:      "test-repo",
										Namespace: "test-ns",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "kustomizations in two stages of the same pipeline",
			kustomizations: []kustomizev1.Kustomization{
				test.NewKustomization(test.InPipeline("demo-pipeline", "staging", ""), test.Source("test-repo", "test-ns")),
				test.NewKustomization(test.Named("production-deploys", "production"),
					test.InPipeline("demo-pipeline", "production", "staging"), test.Source("test-repo", "test-ns")),
			},
			objs: []runtime.Object{
				newGitRepository("test-repo", "test-ns", "https://github.com/example/example.git", &sourcev1.GitRepositoryRef{
					Branch: "testing",
				}),
			},
			want: []KustomizationPipeline{
				{
					Name: "demo-pipeline",
					Environments: []KustomizationEnvironment{
						{
							Name: "staging",
							Kustomizations: []EnvironmentKustomization{
								{
									Reference: &sourcev1.GitRepositoryRef{Branch: "testing"},
									Path:      "./testing",
									URL:       "https://github.com/example/example.git",
									Source: kustomizev1.CrossNamespaceSourceReference{
										Kind: "GitRepository", Name: "test-repo", Namespace: "test-ns",
									},
								},
							},
						},
						{
							Name: "production",
							Kustomizations: []EnvironmentKustomization{
								{
									Reference: &sourcev1.GitRepositoryRef{Branch: "testing"},
									Path:      "./testing",
									URL:       "https://github.com/example/example.git",
									Source: kustomizev1.CrossNamespaceSourceReference{
										Kind: "GitRepository", Name: "test-repo", Namespace: "test-ns",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "kustomizations in the same stage of the same pipeline",
			kustomizations: []kustomizev1.Kustomization{
				test.NewKustomization(test.InPipeline("demo-pipeline", "staging", ""),
					test.Source("test-repo", "test-ns"),
					test.Path("./files1"),
				),
				test.NewKustomization(test.Named("staging-deploys", "staging"),
					test.InPipeline("demo-pipeline", "staging", ""),
					test.Source("test-repo", "test-ns"),
					test.Path("./files2"),
				),
			},
			objs: []runtime.Object{
				newGitRepository("test-repo", "test-ns", "https://github.com/example/example.git", &sourcev1.GitRepositoryRef{
					Branch: "testing",
				}),
			},
			want: []KustomizationPipeline{
				{
					Name: "demo-pipeline",
					Environments: []KustomizationEnvironment{
						{
							Name: "staging",
							Kustomizations: []EnvironmentKustomization{
								{
									Reference: &sourcev1.GitRepositoryRef{Branch: "testing"},
									URL:       "https://github.com/example/example.git",
									Path:      "./files1",
									Source: kustomizev1.CrossNamespaceSourceReference{
										Kind: "GitRepository", Name: "test-repo", Namespace: "test-ns",
									},
								},
								{
									Reference: &sourcev1.GitRepositoryRef{Branch: "testing"},
									URL:       "https://github.com/example/example.git",
									Path:      "./files2",
									Source: kustomizev1.CrossNamespaceSourceReference{
										Kind: "GitRepository", Name: "test-repo", Namespace: "test-ns",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "kustomization with missing source",
			kustomizations: []kustomizev1.Kustomization{
				test.NewKustomization(test.InPipeline("demo-pipeline", "staging", ""), test.Source("test-repo", "test-ns"))},
			want: []KustomizationPipeline{
				{
					Name: "demo-pipeline",
					Environments: []KustomizationEnvironment{
						{
							Name: "staging",
							Kustomizations: []EnvironmentKustomization{
								{
									Path:      "./testing",
									Reference: nil,
									Source: kustomizev1.CrossNamespaceSourceReference{
										Kind:      "GitRepository",
										Name:      "test-repo",
										Namespace: "test-ns",
									},
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
			l := &kustomizev1.KustomizationList{
				Items: tt.kustomizations,
			}

			objs := []runtime.Object{}
			for i := range tt.kustomizations {
				objs = append(objs, &tt.kustomizations[i])
			}
			if tt.objs != nil {
				objs = append(objs, tt.objs...)
			}

			cl := newFakeClient(t, objs...)

			ps, err := ParseKustomizationPipelines(context.TODO(), cl, l)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, ps); diff != "" {
				t.Fatalf("failed to parse pipelines:\n%s", diff)
			}
		})
	}
}

func sourceRef(kind, namespace, name string) kustomizev1.CrossNamespaceSourceReference {
	return kustomizev1.CrossNamespaceSourceReference{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
	}
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := kustomizev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := sourcev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
}

func newGitRepository(name, ns, sourceURL string, ref *sourcev1.GitRepositoryRef) *sourcev1.GitRepository {
	return &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: sourcev1.GitRepositorySpec{
			URL:       sourceURL,
			Reference: ref,
		},
	}
}
