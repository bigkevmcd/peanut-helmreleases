package main

import (
	"context"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	runclient "github.com/fluxcd/pkg/runtime/client"
	"github.com/gitops-tools/apps-scanner/pkg/pipelines"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/bigkevmcd/peanut-helmpipelines/pkg/helm"
)

var (
	scheme            = runtime.NewScheme()
	kubeclientOptions = &runclient.Options{}
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(helmv2.AddToScheme(scheme))
}

func main() {
	cfg, err := config.GetConfig()
	cobra.CheckErr(err)

	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	cobra.CheckErr(err)

	rootCmd := newRootCmd(cl)
	cobra.CheckErr(rootCmd.Execute())
}

func newRootCmd(cl client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm-pipelines",
		Short: "List pipelines in the cluster",
		RunE:  listPipelines(cl),
	}

	kubeclientOptions.BindFlags(cmd.PersistentFlags())

	return cmd
}

func listPipelines(cl client.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		fmt.Println("Starting to scan for helm releases")
		helmReleaseList := &helmv2.HelmReleaseList{}
		err := cl.List(context.Background(), helmReleaseList, client.HasLabels([]string{pipelines.PipelineNameLabel}))
		if err != nil {
			return fmt.Errorf("failed to list helm releases: %w", err)
		}
		fmt.Printf("found %d helm releases\n", len(helmReleaseList.Items))

		helmPipelines, err := helm.ParseHelmReleasePipelines(helmReleaseList.Items)
		if err != nil {
			return fmt.Errorf("failed to discover pipelines: %w", err)
		}

		for _, v := range helmPipelines {
			for _, env := range v.Environments {
				fmt.Printf("pipeline: %s stage: %s: %v\n", v.Name, env.Name, env.Charts)
			}
		}
		return nil
	}
}
