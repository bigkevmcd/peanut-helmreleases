package server

import (
	"context"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/gitops-tools/apps-scanner/pkg/pipelines"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/bigkevmcd/peanut-helmpipelines/pkg/helm"
	pipelinesv1 "github.com/bigkevmcd/peanut-helmpipelines/pkg/protos/pipelines/v1"
)

var _ pipelinesv1.PipelinesServiceServer = (*pipelinesGRPCServer)(nil)

type pipelinesGRPCServer struct {
	pipelinesv1.UnimplementedPipelinesServiceServer
	logr.Logger
	client.Client
}

// NewPipelinesServer creates a new server.
func NewPipelinesServer(l logr.Logger, c client.Client) *pipelinesGRPCServer {
	return &pipelinesGRPCServer{Logger: l, Client: c}
}

func (s *pipelinesGRPCServer) ListPipelines(ctx context.Context, in *pipelinesv1.ListPipelinesRequest) (*pipelinesv1.ListPipelinesResponse, error) {
	helmReleaseList := &helmv2.HelmReleaseList{}
	err := s.Client.List(ctx, helmReleaseList, client.HasLabels([]string{pipelines.PipelineNameLabel}))
	if err != nil {
		return nil, fmt.Errorf("failed to list helm releases: %w", err)
	}
	// fmt.Printf("found %d helm releases\n", len(helmReleaseList.Items))

	helmPipelines, err := helm.ParseHelmReleasePipelines(helmReleaseList)
	if err != nil {
		fmt.Errorf("failed to discover pipelines: %w", err)
	}

	return &pipelinesv1.ListPipelinesResponse{Results: pipelinesToResponse(helmPipelines)}, nil
}

func pipelinesToResponse(hp []helm.HelmReleasePipeline) []*pipelinesv1.Pipeline {
	result := []*pipelinesv1.Pipeline{}
	for _, v := range hp {
		result = append(result, &pipelinesv1.Pipeline{
			Name:         v.Name,
			Environments: envsToResponseEnvironments(v.Environments),
		})
	}
	return result
}

func envsToResponseEnvironments(envs []helm.HelmReleaseEnvironment) []*pipelinesv1.Pipeline_Environment {
	result := []*pipelinesv1.Pipeline_Environment{}
	for _, ev := range envs {
		pe := &pipelinesv1.Pipeline_Environment{Name: ev.Name}
		for _, c := range ev.Charts {
			pe.Charts = append(pe.Charts, &pipelinesv1.Pipeline_Environment_HelmChart{
				Name:    c.Name,
				Version: c.Version,
				Source:  referenceToSource(c.Source),
			})
		}
		result = append(result, pe)
	}
	return result
}

func referenceToSource(r helmv2.CrossNamespaceObjectReference) *pipelinesv1.CrossNamespaceObjectReference {
	return &pipelinesv1.CrossNamespaceObjectReference{
		Kind:      r.Kind,
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}
