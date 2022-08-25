package kustomizations

import (
	"context"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/gitops-tools/apps-scanner/pkg/pipelines"
	"github.com/gitops-tools/pkg/sets"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KustomizationPipelines provides a mapping of Kustomizations in environments
// to their pipelines.
type KustomizationPipeline struct {
	Name         string
	Environments []KustomizationEnvironment
}

// KustomizationEnvironment represents the resources being deployed.
type KustomizationEnvironment struct {
	Name           string
	Kustomizations []EnvironmentKustomization
}

// EnvironmentKustomization is the source details for a Kustomization resource.
type EnvironmentKustomization struct {
	Path      string
	Reference *sourcev1.GitRepositoryRef
	URL       string
	Source    kustomizev1.CrossNamespaceSourceReference
}

// ParseKustomizationPipelines parses the pipelines and the versions of the
// GitRepository resources referenced by the Kustomizations in each stage of the
// pipeline.
func ParseKustomizationPipelines(ctx context.Context, cl client.Client, kusts ...*kustomizev1.Kustomization) ([]KustomizationPipeline, error) {
	p := pipelines.NewParser()
	if err := p.Add(kustomizationsToRuntimeObjects(kusts)); err != nil {
		return nil, fmt.Errorf("failed to parse Kustomizations: %w", err)
	}
	ps, err := p.Pipelines()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate pipelines: %w", err)
	}

	kustomizations := parsePipelineKustomizations(kusts)
	parsed := []KustomizationPipeline{}
	for _, pipeline := range ps {
		envsToKustomizations := map[string]sets.Set[EnvironmentKustomization]{}
		for _, k := range kustomizations[pipeline.Name] {
			envKustomizations := envsToKustomizations[k.environment]
			if envKustomizations == nil {
				envKustomizations = sets.New[EnvironmentKustomization]()
			}
			envKustomizations.Insert(EnvironmentKustomization{Source: k.source, Path: k.path})
			envsToKustomizations[k.environment] = envKustomizations
		}

		kp := KustomizationPipeline{Name: pipeline.Name, Environments: []KustomizationEnvironment{}}
		for _, envName := range pipeline.Environments {
			// TODO: Come up with a better sort
			kustomizations := envsToKustomizations[envName].SortedList(func(x, y EnvironmentKustomization) bool {
				return x.Path < y.Path
			})
			for i := range kustomizations {
				// TODO: Ignore if not GitRepository
				k := kustomizations[i]
				repo, err := loadGitRepository(ctx, cl, client.ObjectKey{Name: k.Source.Name, Namespace: k.Source.Namespace})
				if err != nil {
					return nil, fmt.Errorf("failed to load source %v: %w", k.Source, err)
				}
				if repo != nil {
					k.URL = repo.Spec.URL
					k.Reference = repo.Spec.Reference
					kustomizations[i] = k
				}
			}
			kp.Environments = append(kp.Environments,
				KustomizationEnvironment{Name: envName, Kustomizations: kustomizations})
		}
		parsed = append(parsed, kp)
	}

	return parsed, nil
}

func loadGitRepository(ctx context.Context, cl client.Client, o client.ObjectKey) (*sourcev1.GitRepository, error) {
	var repo sourcev1.GitRepository
	if err := cl.Get(ctx, o, &repo); err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	return &repo, nil
}

type pipelineKustomization struct {
	pipeline    string
	environment string
	source      kustomizev1.CrossNamespaceSourceReference
	path        string
}

// returns a map of pipeline -> pipelineKustomization
func parsePipelineKustomizations(kusts []*kustomizev1.Kustomization) map[string][]pipelineKustomization {
	discovered := map[string][]pipelineKustomization{}

	for _, k := range kusts {
		lbls := k.GetLabels()
		pipeline := lbls[pipelines.PipelineNameLabel]
		env := lbls[pipelines.PipelineEnvironmentLabel]
		// We can't place Kustomizations into any env if they are not labelled.
		if pipeline == "" || env == "" {
			continue
		}
		path, source := k.Spec.Path, k.Spec.SourceRef
		pc := discovered[pipeline]
		if pc == nil {
			pc = []pipelineKustomization{}
		}

		pc = append(pc, pipelineKustomization{
			pipeline: pipeline, environment: env,
			path:   path,
			source: source,
		})
		discovered[pipeline] = pc
	}

	return discovered
}

func kustomizationsToRuntimeObjects(kusts []*kustomizev1.Kustomization) []runtime.Object {
	objs := make([]runtime.Object, len(kusts))
	for i := range kusts {
		objs[i] = kusts[i]
	}

	return objs
}
