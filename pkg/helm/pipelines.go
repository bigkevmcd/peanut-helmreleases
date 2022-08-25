package helm

import (
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/gitops-tools/pkg/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gitops-tools/apps-scanner/pkg/pipelines"
)

// HelmReleasePipelines provides a mapping of Helm charts in environments to
// their pipelines.
type HelmReleasePipeline struct {
	Name         string
	Environments []HelmReleaseEnvironment
}

// HelmReleaseEnvironment represents the charts in a specific staged of a
// pipeline.
type HelmReleaseEnvironment struct {
	Name              string
	Charts            []HelmReleaseChart
	ChartHelmReleases map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference
}

// HelmReleaseChart is the specific version of the chart in a HelmRelease.
type HelmReleaseChart struct {
	Name    string
	Version string
	Source  helmv2.CrossNamespaceObjectReference
}

// ParseHelmReleasePipelines parses the pipelines and the versions of the charts
// used by the HelmReleases in each stage in each pipeline.
func ParseHelmReleasePipelines(hl *helmv2.HelmReleaseList) ([]HelmReleasePipeline, error) {
	p := pipelines.NewParser()
	if err := p.Add(hl); err != nil {
		return nil, fmt.Errorf("failed to parse HelmReleases: %w", err)
	}
	ps, err := p.Pipelines()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate pipelines: %w", err)
	}

	charts := parsePipelineCharts(hl.Items)
	parsed := []HelmReleasePipeline{}
	chartHelmReleases := map[HelmReleaseChart]sets.Set[helmv2.CrossNamespaceObjectReference]{}
	for _, pipeline := range ps {
		envsToCharts := map[string]sets.Set[HelmReleaseChart]{}
		for _, c := range charts[pipeline.Name] {
			envCharts := envsToCharts[c.environment]
			hrc := HelmReleaseChart{Name: c.chart, Version: c.version, Source: c.source}
			if envCharts == nil {
				envCharts = sets.New[HelmReleaseChart]()
			}
			helmReleases := chartHelmReleases[hrc]
			if helmReleases == nil {
				helmReleases = sets.New[helmv2.CrossNamespaceObjectReference]()
			}
			envCharts.Insert(hrc)
			helmReleases.Insert(c.helmRelease)
			envsToCharts[c.environment] = envCharts
			chartHelmReleases[hrc] = helmReleases
		}

		hrp := HelmReleasePipeline{Name: pipeline.Name, Environments: []HelmReleaseEnvironment{}}
		for _, envName := range pipeline.Environments {
			hrp.Environments = append(hrp.Environments,
				HelmReleaseEnvironment{Name: envName,
					Charts:            envsToCharts[envName].List(),
					ChartHelmReleases: unpackChartReleases(chartHelmReleases),
				})
		}
		parsed = append(parsed, hrp)
	}

	return parsed, nil
}

func unpackChartReleases(packed map[HelmReleaseChart]sets.Set[helmv2.CrossNamespaceObjectReference]) map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference {
	unpacked := map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference{}
	for k, v := range packed {
		unpacked[k] = v.List()
	}
	return unpacked
}

type pipelineChart struct {
	pipeline    string
	environment string
	chart       string
	version     string
	source      helmv2.CrossNamespaceObjectReference
	helmRelease helmv2.CrossNamespaceObjectReference
}

func parsePipelineCharts(releases []helmv2.HelmRelease) map[string][]pipelineChart {
	discovered := map[string][]pipelineChart{}

	for _, hr := range releases {
		lbls := hr.GetLabels()
		pipeline := lbls[pipelines.PipelineNameLabel]
		env := lbls[pipelines.PipelineEnvironmentLabel]
		// We can't place HelmReleases into any env if they are not labelled.
		if pipeline == "" || env == "" {
			continue
		}
		chart, version := hr.Spec.Chart.Spec.Chart, hr.Spec.Chart.Spec.Version
		pc := discovered[pipeline]
		if pc == nil {
			pc = []pipelineChart{}
		}

		pc = append(pc, pipelineChart{
			pipeline: pipeline, environment: env,
			chart: chart, version: version,
			source:      hr.Spec.Chart.Spec.SourceRef,
			helmRelease: objectReferenceFromObject(&hr),
		})
		discovered[pipeline] = pc
	}

	return discovered
}

func objectReferenceFromObject(obj client.Object) helmv2.CrossNamespaceObjectReference {
	apiVersion, kind := obj.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
	return helmv2.CrossNamespaceObjectReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(),
	}
}
