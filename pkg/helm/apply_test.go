package helm

import (
	"context"
	"testing"

	"github.com/bigkevmcd/peanut-helmpipelines/test"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestApplyPromotions(t *testing.T) {
	items := []helmv2.HelmRelease{
		test.NewHelmRelease(test.InPipeline("demo-pipeline", "staging", ""), test.Named("staging-deploy", "staging"), test.ChartVersion("redis", "1.0.12")),
		test.NewHelmRelease(test.InPipeline("demo-pipeline", "production", "staging"), test.Named("production-deploy", "production"), test.ChartVersion("redis", "1.0.9")),
	}
	pipelines, err := ParseHelmReleasePipelines(items)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(pipelines); l != 1 {
		t.Fatalf("got %d pipelines, want 1", l)
	}

	promotions := CalculatePromotions(pipelines[0])
	fc := newFakeClient(t, releasesToRuntimeObjects(items)...)

	if err := ApplyPromotions(context.TODO(), fc, promotions); err != nil {
		t.Fatal(err)
	}

	updated := helmv2.HelmRelease{}
	if err := fc.Get(context.TODO(), client.ObjectKeyFromObject(&items[1]), &updated); err != nil {
		t.Fatal(err)
	}

	want := helmv2.HelmChartTemplateSpec{
		Chart:   "redis",
		Version: "1.0.12",
		SourceRef: helmv2.CrossNamespaceObjectReference{
			Name:      "test-repository",
			Kind:      "HelmRepository",
			Namespace: "default",
		},
	}
	if diff := cmp.Diff(want, updated.Spec.Chart.Spec); diff != "" {
		t.Fatalf("failed to apply promotions:\n%s", diff)
	}
}
