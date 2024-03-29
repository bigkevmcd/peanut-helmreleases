package helm

import (
	"testing"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/google/go-cmp/cmp"
)

func TestCalculatePromotions(t *testing.T) {
	promotionTests := []struct {
		name     string
		pipeline HelmReleasePipeline
		want     []Promotion
	}{
		{
			name: "single environment - no upgrades",
			pipeline: HelmReleasePipeline{
				Name: "demo-pipeline",
				Environments: []HelmReleaseEnvironment{
					{
						Name: "staging",
						Charts: []HelmReleaseChart{
							{
								Name:    "redis",
								Version: "1.0.12",
								Source:  sourceRef("HelmRepository", "default", "test-repository"),
							},
						},
					},
				},
				ChartHelmReleases: map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference{},
			},
			want: []Promotion{},
		},
		{
			name: "single upgrade between two environments",
			pipeline: HelmReleasePipeline{
				Name: "demo-pipeline",
				Environments: []HelmReleaseEnvironment{
					{
						Name: "staging",
						Charts: []HelmReleaseChart{
							{
								Name:    "redis",
								Version: "1.0.12",
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
				ChartHelmReleases: map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference{
					HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.9",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					}: []helmv2.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "redis-production", Namespace: "testing"},
					},
				},
			},
			want: []Promotion{
				{
					Environment: "production",
					From: HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.9",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					To: HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.12",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					PromotedReleases: []v2beta1.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "redis-production", Namespace: "testing"},
					},
				},
			},
		},
		{
			name: "charts in one environment, not the other",
			pipeline: HelmReleasePipeline{
				Name: "demo-pipeline",
				Environments: []HelmReleaseEnvironment{
					{
						Name: "staging",
						Charts: []HelmReleaseChart{
							{
								Name:    "redis",
								Version: "1.0.12",
								Source:  sourceRef("HelmRepository", "default", "test-repository"),
							},
							{
								Name:    "postgresql",
								Version: "13.0.1",
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
				ChartHelmReleases: map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference{
					HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.9",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					}: []helmv2.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "redis-production", Namespace: "testing"},
					},
				},
			},
			want: []Promotion{
				{
					Environment: "production",
					From: HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.9",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					To: HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.12",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					PromotedReleases: []v2beta1.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "redis-production", Namespace: "testing"}},
				},
			},
		},
		{
			name: "multiple upgrades between two environments",
			pipeline: HelmReleasePipeline{
				Name: "demo-pipeline",
				Environments: []HelmReleaseEnvironment{
					{
						Name: "staging",
						Charts: []HelmReleaseChart{
							{
								Name:    "redis",
								Version: "1.0.12",
								Source:  sourceRef("HelmRepository", "default", "test-repository"),
							},
							{
								Name:    "postgresql",
								Version: "13.0.1",
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
							{
								Name:    "postgresql",
								Version: "13.0.0",
								Source:  sourceRef("HelmRepository", "default", "test-repository"),
							},
						},
					},
				},
				ChartHelmReleases: map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference{
					{
						Name:    "redis",
						Version: "1.0.9",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					}: []helmv2.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "redis-production", Namespace: "default"},
					},
					{
						Name:    "postgresql",
						Version: "13.0.0",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					}: []helmv2.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "postgres-production", Namespace: "default"},
					},
				},
			},
			want: []Promotion{
				{
					Environment: "production",
					From: HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.9",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					To: HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.12",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					PromotedReleases: []v2beta1.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "redis-production", Namespace: "default"},
					},
				},
				{
					Environment: "production",
					From: HelmReleaseChart{
						Name:    "postgresql",
						Version: "13.0.0",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					To: HelmReleaseChart{
						Name:    "postgresql",
						Version: "13.0.1",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					PromotedReleases: []v2beta1.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "postgres-production", Namespace: "default"},
					},
				},
			},
		},
		{
			name: "promotions ignores identical charts",
			pipeline: HelmReleasePipeline{
				Name: "demo-pipeline",
				Environments: []HelmReleaseEnvironment{
					{
						Name: "staging",
						Charts: []HelmReleaseChart{
							{
								Name:    "redis",
								Version: "1.0.12",
								Source:  sourceRef("HelmRepository", "default", "test-repository"),
							},
							{
								Name:    "postgresql",
								Version: "13.0.1",
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
							{
								Name:    "postgresql",
								Version: "13.0.1",
								Source:  sourceRef("HelmRepository", "default", "test-repository"),
							},
						},
					},
				},
				ChartHelmReleases: map[HelmReleaseChart][]helmv2.CrossNamespaceObjectReference{
					HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.9",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					}: []helmv2.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "redis-production", Namespace: "testing"},
					},
				},
			},
			want: []Promotion{
				{
					Environment: "production",
					From: HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.9",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					To: HelmReleaseChart{
						Name:    "redis",
						Version: "1.0.12",
						Source:  sourceRef("HelmRepository", "default", "test-repository"),
					},
					PromotedReleases: []v2beta1.CrossNamespaceObjectReference{
						{Kind: "HelmRelease", Name: "redis-production", Namespace: "testing"},
					},
				},
			},
		},
	}

	for _, tt := range promotionTests {
		t.Run(tt.name, func(t *testing.T) {
			promotions := CalculatePromotions(tt.pipeline)

			if diff := cmp.Diff(tt.want, promotions); diff != "" {
				t.Fatalf("failed to calculate promotions:\n%s", diff)
			}
		})
	}
}
