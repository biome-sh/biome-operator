# biome-operator

Installs [biome-operator](https://github.com/biome-sh/biome-operator) to manage Biome services in a Kubernetes cluster.

## Introduction

This chart bootstraps a [biome-operator](https://github.com/biome-sh/biome-operator) deployment in a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

Note: If you have Kubernetes and Helm installed, skip to [this section](#Installing the Chart).

See [Biome Operator's README](https://github.com/biome-sh/biome-operator/blob/master/README.md).

### RBAC
If role-based access control (RBAC) is enabled in your cluster, you may need to give Tiller (the server-side component of Helm) additional permissions. **If RBAC is not enabled, be sure to set `rbacEnable` to `false` when installing the chart.**

1. Create a ServiceAccount for Tiller in the `kube-system` namespace
```console
$ kubectl -n kube-system create sa tiller
```

2. Create a ClusterRoleBinding for Tiller

```console
$ kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
```

3. Install Tiller, specifying the new ServiceAccount

```console
$ helm init --service-account tiller
```

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm repo add biome https://biome-sh.github.io/biome-operator/helm/charts/stable/
$ helm install --name my-release biomesh/biome-operator
```

The command deploys biome-operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the biome-operator chart and their default values.

Parameter | Description | Default
--- | --- | ---
`image.repository` | Image | `docker.io/biomesh/biome-operator`
`image.tag` | Image tag | The latest release tag (e.g `v0.8.1`)
`image.pullPolicy` | Image pull policy | `IfNotPresent`
`nodeSelector` | Node labels for pod assignment | `{}`
`rbacEnable` | If true, create & use RBAC resources | `true`
`resources` | Pod resource requests & limits | `{}`
`operatorNamespaced` | If this operator should run scoped to Single namespace | `false`
`namespace` | Namespace this operator should run inside | `biome-operator`

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install --name my-release biomesh/biome-operator --set sendAnalytics=true
```

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install --name my-release biomesh/biome-operator -f values.yaml
```

> **Tip**: You can use the default [values.yaml](values.yaml)

## Hosting

The Helm chart repository is managed on the [`gh-pages` branch](https://github.com/biome-sh/biome-operator/tree/gh-pages) and instructions for repository management can be found [here](https://github.com/biome-sh/biome-operator/tree/gh-pages/helm/charts).
