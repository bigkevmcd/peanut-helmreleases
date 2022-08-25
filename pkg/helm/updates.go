package helm

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sort"

	"github.com/Masterminds/semver"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var chartGetters = getter.Providers{
	getter.Provider{
		Schemes: []string{"http", "https"},
		New:     getter.NewHTTPGetter,
	},
}

// ChartUpgrade represents an available upgrade, in terms of current, and new
// chart versions.
type ChartUpgrade struct {
	Current   HelmReleaseChart
	Available HelmReleaseChart
}

// IdentifyUpgrades looks for upgradable charts in a pipeline.
//
// An upgradable chart has a newer version.
func IdentifyUpgrades(ctx context.Context, p HelmReleasePipeline, c client.Client) ([]ChartUpgrade, error) {
	upgrades := []ChartUpgrade{}
	for _, env := range p.Environments {
		for _, chart := range env.Charts {
			// TODO: cache this
			index, err := getChartIndex(ctx, chart, c)
			if err != nil {
				// TODO wrap?
				return nil, err
			}
			newer, err := findNewerVersionOfChart(chart, index)
			if err != nil {
				// TODO wrap?
				return nil, err
			}
			if newer != nil {
				upgrades = append(upgrades, ChartUpgrade{
					Current:   chart,
					Available: *newer,
				})
			}
		}
	}

	return upgrades, nil
}

func getChartIndex(ctx context.Context, chart HelmReleaseChart, c client.Client) (*repo.IndexFile, error) {
	// TODO: abandon if the reference is not to a HelmRepository
	hr := &sourcev1.HelmRepository{}
	if err := c.Get(ctx, types.NamespacedName{Name: chart.Source.Name, Namespace: chart.Source.Namespace}, hr); err != nil {
		return nil, err
	}

	// TODO: what if the URL == "" ?
	u, err := url.Parse(hr.Status.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL %q: %w", hr.Status.URL, err)
	}

	getter, err := chartGetters.ByScheme(u.Scheme)
	if err != nil {
		return nil, fmt.Errorf("no provider for scheme %q: %w", u.Scheme, err)
	}

	res, err := getter.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error fetching index file: %w", err)
	}

	b, err := io.ReadAll(res)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	indexFile := &repo.IndexFile{}
	if err := yaml.Unmarshal(b, indexFile); err != nil {
		return nil, fmt.Errorf("error unmarshaling chart response: %w", err)
	}

	if indexFile.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}
	indexFile.SortEntries()

	return indexFile, nil
}

func findNewerVersionOfChart(chart HelmReleaseChart, index *repo.IndexFile) (*HelmReleaseChart, error) {
	parsed, err := semver.NewVersion(chart.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version %q for chart %q", chart.Version, chart.Name)
	}
	newerVersions := []*semver.Version{}
	for _, v := range index.Entries[chart.Name] {
		entryVersion, err := semver.NewVersion(v.Metadata.Version)
		if err != nil {
			// This shouldn't happen
			return nil, err
		}
		if entryVersion.GreaterThan(parsed) {
			newerVersions = append(newerVersions, entryVersion)
		}
	}

	if len(newerVersions) == 0 {
		return nil, nil
	}
	sort.Sort(sort.Reverse(semver.Collection(newerVersions)))

	return &HelmReleaseChart{
		Name:    chart.Name,
		Version: newerVersions[0].String(),
		Source:  chart.Source,
	}, nil
}
