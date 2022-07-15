package helm

// Promotion is a calculated upgrade for an environment.
type Promotion struct {
	Environment string
	From        HelmReleaseChart
	To          HelmReleaseChart
}

// CalculatePromotions calculates a set of Promotions based on the differences
// between environments.
//
// A promotion is not necessarily a newer version, only a directly immediate
// environment has the same chart with a different version.
func CalculatePromotions(pipeline HelmReleasePipeline) []Promotion {
	pairs := calculatePromotionPairs(pipeline)
	promotions := []Promotion{}
	for _, pair := range pairs {
		for _, v := range pair.toCharts {
			if upgrade := findChart(v, pair.fromCharts); upgrade != nil {
				promotions = append(promotions, Promotion{Environment: pair.to, From: v, To: *upgrade})
			}
		}
	}

	return promotions
}

// find a matching chart in the provided list, ignoring the version.
func findChart(chart HelmReleaseChart, charts []HelmReleaseChart) *HelmReleaseChart {
	for _, c := range charts {
		if (c.Name == chart.Name && c.Source == chart.Source) && c.Version != chart.Version {
			return &c
		}
	}

	return nil
}

type promotionPair struct {
	from       string
	fromCharts []HelmReleaseChart

	to       string
	toCharts []HelmReleaseChart
}

func calculatePromotionPairs(p HelmReleasePipeline) []promotionPair {
	pairs := []promotionPair{}
	for i := range p.Environments {
		if i < len(p.Environments)-1 {
			pairs = append(pairs,
				promotionPair{
					from: p.Environments[i].Name, fromCharts: p.Environments[i].Charts,
					to: p.Environments[i+1].Name, toCharts: p.Environments[i+1].Charts})
		}
	}
	return pairs
}
