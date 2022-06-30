# peanut-pipelines

Delivery pipeline discovery from Flux HelmReleases.

This is **experimental** code.

# Getting started

## Install the base Flux components:

```shell
$ flux install --components source-controller,helm-controller
```

## Install some example resources

```shell
$ kubectl create -f examples/example-helm-releases.yaml
```

## Build and run

### Command-line tool

```shell
$ go build ./cmd/helm-pipelines
$ ./helm-pipelines
Starting to scan for helm releases
found 2 helm releases
pipeline: demo-pipeline stage: staging: [{podinfo 6.1.6 HelmRepository/default/podinfo-repo}]
pipeline: demo-pipeline stage: production: [{podinfo 6.1.5 HelmRepository/default/podinfo-repo}]
```

This discovered a pipeline called `demo-pipeline` with two environments `staging` and `production` one with v6.1.6 of podinfo and the other with v6.1.5 of podinfo.

### gRPC Server

```shell
$ go build ./cmd/peanut-pipelines
$ ./peanut-pipelines
2022/06/28 07:46:20 Listening at :8080
```

In a separate terminal...

```shell
$ grpcurl -plaintext localhost:8080 pipelines.v1.PipelinesService/ListPipelines
{
  "results": [
    {
      "name": "demo-pipeline",
      "environments": [
        {
          "name": "staging",
          "charts": [
            {
              "name": "podinfo",
              "version": "6.1.6",
              "source": "HelmRepository/default/podinfo-repo"
            }
          ]
        },
        {
          "name": "production",
          "charts": [
            {
              "name": "podinfo",
              "version": "6.1.5",
              "source": "HelmRepository/default/podinfo-repo"
            }
          ]
        }
      ]
    }
  ]
}
```
