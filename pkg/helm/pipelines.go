package helm

import (
	"fmt"
	"sort"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"k8s.io/apimachinery/pkg/util/sets"

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
	Name   string
	Charts []HelmReleaseChart
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
	for _, pipeline := range ps {
		envsToCharts := map[string]helmReleaseChartSet{}
		for _, c := range charts[pipeline.Name] {
			envCharts := envsToCharts[c.environment]
			if envCharts == nil {
				envCharts = newHelmReleaseCharts()
			}
			envCharts.insert(HelmReleaseChart{Name: c.chart, Version: c.version, Source: c.source})
			envsToCharts[c.environment] = envCharts
		}

		hrp := HelmReleasePipeline{Name: pipeline.Name, Environments: []HelmReleaseEnvironment{}}
		for _, envName := range pipeline.Environments {
			hrp.Environments = append(hrp.Environments,
				HelmReleaseEnvironment{Name: envName, Charts: envsToCharts[envName].List()})
		}
		parsed = append(parsed, hrp)
	}

	return parsed, nil
}

type pipelineChart struct {
	pipeline    string
	environment string
	chart       string
	version     string
	source      helmv2.CrossNamespaceObjectReference
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
			source: hr.Spec.Chart.Spec.SourceRef,
		})
		discovered[pipeline] = pc
	}

	return discovered
}

type helmReleaseChartSet map[HelmReleaseChart]sets.Empty

// newHelmReleaseCharts creates and returns a new set of helmReleaseChartSet.
func newHelmReleaseCharts(items ...HelmReleaseChart) helmReleaseChartSet {
	ss := helmReleaseChartSet{}
	return ss.insert(items...)
}

func (s helmReleaseChartSet) insert(items ...HelmReleaseChart) helmReleaseChartSet {
	for _, item := range items {
		s[item] = sets.Empty{}
	}
	return s
}

// List returns the contents as a sorted slice.
// WARNING: This is suboptimal as it's stringifying on each comparison, there
// aren't expected to be a huge number of helmReleaseChartSet.
func (s helmReleaseChartSet) List() []HelmReleaseChart {
	if len(s) == 0 {
		return nil
	}
	res := []HelmReleaseChart{}
	for key := range s {
		res = append(res, key)
	}
	sortString := func(r HelmReleaseChart) string {
		return fmt.Sprintf("name=%q version=%q source=%q", r.Name, r.Version, r.Source)
	}
	sort.Slice(res, func(i, j int) bool {
		return sortString(res[i]) < sortString(res[j])
	})
	return res
}
