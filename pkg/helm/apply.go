package helm

import (
	"context"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyPromotions updates the promotions to the cluster.
//
// For each promotion, the "To" version will be applied to the HelmReleases
// identified as requiring update for that version.
func ApplyPromotions(ctx context.Context, cl client.Client, proms []Promotion) error {
	for _, promotion := range proms {
		for _, cno := range promotion.PromotedReleases {
			hr := &helmv2.HelmRelease{}
			if err := cl.Get(ctx, keyFromCrossNamespaceObject(cno), hr); err != nil {
				// TODO: better error?
				return err
			}
			hr.Spec.Chart.Spec.Version = promotion.To.Version
			if err := cl.Update(ctx, hr); err != nil {
				return err
			}
		}
	}

	return nil
}

func keyFromCrossNamespaceObject(obj helmv2.CrossNamespaceObjectReference) client.ObjectKey {
	return client.ObjectKey{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
}
