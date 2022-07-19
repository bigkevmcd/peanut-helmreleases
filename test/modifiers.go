package test

import (
	"github.com/gitops-tools/apps-scanner/pkg/pipelines"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InPipeline is an option for resources that applies the correct labels to
// indicate that the resource is in a pipeline.
func InPipeline(name, env, after string) func(client.Object) {
	return func(hr client.Object) {
		lbls := hr.GetLabels()
		if lbls == nil {
			lbls = map[string]string{}
		}
		lbls[pipelines.PipelineNameLabel] = name
		lbls[pipelines.PipelineEnvironmentLabel] = env
		lbls[pipelines.PipelineEnvironmentAfterLabel] = after
		hr.SetLabels(lbls)
	}
}

// Named is an option that sets the name on created resources.
func Named(name, namespace string) func(client.Object) {
	return func(hr client.Object) {
		hr.SetName(name)
		hr.SetNamespace(namespace)
	}
}
