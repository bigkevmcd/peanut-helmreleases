package helm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIdentifyUpgrades_invalid_chart(t *testing.T) {
}

func TestIdentifyUpgrades(t *testing.T) {
	upgradesTests := []struct {
		name     string
		dir      string
		pipeline HelmReleasePipeline

		want []ChartUpgrade
	}{
		{
			name: "single upgradable chart", dir: "testdata/example-charts",
			pipeline: HelmReleasePipeline{
				Name: "testing",
				Environments: []HelmReleaseEnvironment{
					{
						Name: "dev",
						Charts: []HelmReleaseChart{
							{
								Name:    "test-service",
								Version: "1.0.1",
								Source: helmv2.CrossNamespaceObjectReference{
									Kind:      "HelmRepository",
									Name:      "testing",
									Namespace: "testing",
								},
							},
						},
					},
				},
			},
			want: []ChartUpgrade{
				{
					Current: HelmReleaseChart{
						Name:    "test-service",
						Version: "1.0.1",
						Source:  helmv2.CrossNamespaceObjectReference{Kind: "HelmRepository", Name: "testing", Namespace: "testing"},
					},
					Available: HelmReleaseChart{
						Name:    "test-service",
						Version: "1.1.2",
						Source:  helmv2.CrossNamespaceObjectReference{Kind: "HelmRepository", Name: "testing", Namespace: "testing"},
					},
				},
			},
		},
	}

	for _, tt := range upgradesTests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.FileServer(http.Dir(tt.dir)))
			defer testServer.Close()

			hr := &sourcev1.HelmRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testing",
					Namespace: "testing",
				},
				Status: sourcev1.HelmRepositoryStatus{
					URL: testServer.URL + "/index.yaml",
				},
			}

			upgrades, err := IdentifyUpgrades(context.TODO(), tt.pipeline, newFakeClient(t, hr))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, upgrades); diff != "" {
				t.Fatalf("failed to identify upgrades: %s\n", diff)
			}
		})
	}
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := helmv2.AddToScheme(scheme); err != nil {
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
