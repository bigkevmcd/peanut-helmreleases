package helm

import (
	"testing"

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
