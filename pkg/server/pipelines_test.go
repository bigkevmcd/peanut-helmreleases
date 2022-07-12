package server

import (
	"context"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	pipelinesv1 "github.com/bigkevmcd/peanut-helmpipelines/pkg/protos/pipelines/v1"
	v1 "github.com/bigkevmcd/peanut-helmpipelines/pkg/protos/pipelines/v1"
	"github.com/bigkevmcd/peanut-helmpipelines/test"
)

func TestListPipelines(t *testing.T) {
	hr := test.NewHelmRelease(test.InPipeline("demo-pipeline", "staging", ""))
	fc := newFakeClient(t, &hr)
	srv := NewPipelinesServer(logr.Discard(), fc)
	in := &pipelinesv1.ListPipelinesRequest{}

	resp, err := srv.ListPipelines(context.TODO(), in)
	if err != nil {
		t.Fatal(err)
	}

	pipelines := resp.GetResults()
	want := []*pipelinesv1.Pipeline{
		{
			Name: "demo-pipeline",
			Environments: []*v1.Pipeline_Environment{
				{
					Name: "staging",
					Charts: []*pipelinesv1.Pipeline_Environment_HelmChart{
						{
							Name:    "redis",
							Version: "1.0.9",
							Source: &pipelinesv1.CrossNamespaceObjectReference{
								Kind:      "HelmRepository",
								Namespace: "default",
								Name:      "test-repository",
							},
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, pipelines,
		cmpopts.IgnoreUnexported(timestamppb.Timestamp{}),
		cmpopts.IgnoreUnexported(pipelinesv1.Pipeline_Environment{}),
		cmpopts.IgnoreUnexported(pipelinesv1.Pipeline_Environment_HelmChart{}),
		cmpopts.IgnoreUnexported(pipelinesv1.CrossNamespaceObjectReference{}),
		cmpopts.IgnoreUnexported(pipelinesv1.Pipeline{})); diff != "" {
		t.Fatalf("incorrect pipelines response:\n%s", diff)
	}
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := helmv2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
}
