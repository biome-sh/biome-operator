# Overview

This repo is community distro of Chef Habitat Application Automation.

It was generated from original using small tool [ForkMan](https://github.com/jsirex/forkman).

## Backporting changes

There is `patch-project.sh` example. It reworks project and updates branch `forkman-raw`. It tracks upstream changes.

Here is how can you proceed:

1. Clone this project
1. Add `habitat` repo as remote: `git remote add habitat https://github.com/habitat-sh/habitat-operator.git`
1. Fetch latest changes: `git fetch habitat master`
1. Make sure you have local branch up to date with `origin/forkman-raw`. Script will use it as base for new changes.
1. Use `patch-project.sh` script with `habitat-sh/habitat.yml` config to refactor original repo:
``` shell
./patch-project.sh habitat-sh/habitat.yaml ../biome-operator habitat/master`
```
1. Do not try to modify this branch, because new `forkman` run override your change with dictionary. Keep branch clean
1. Push back `forkman-raw` branch (or create PR)
1. If after `forkman` project requires additional manual fixes create another temporary branch from `master`, merge there `forkman-raw` and add patches after. It will presrve your changes later

It is highly recommended to not change code by hand because it requires effort for further support. If it possible to patch with `forkman` - use forkman

# Upstream README

[![Build Status](https://circleci.com/gh/habitat-sh/habitat-operator.svg?style=svg)](https://circleci.com/gh/habitat-sh/habitat-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/habitat-sh/habitat-operator)](https://goreportcard.com/report/github.com/habitat-sh/habitat-operator)

# habitat-operator

This project is currently unstable - breaking changes may still land in the future.

## Overview

The Habitat operator is a Kubernetes controller designed to solve running and auto-managing Habitat Services on Kubernetes. It does this by making use of [`Custom Resource Definitions`][crd].

To learn more about Habitat, please visit the [Habitat website](https://www.habitat.sh/).

For a more detailed description of the Habitat type have a look [here](https://github.com/habitat-sh/habitat-operator/blob/master/pkg/apis/habitat/v1beta1/types.go).

## Prerequisites

- Habitat `>= 0.52.0`
- Kubernetes cluster with version `1.9.x`, `1.10.x` or `1.11.x`
- Kubectl version `1.10.x` or `1.11.x`

## Installing

Make sure you have [golang compiler](https://golang.org/dl/) installed. Follow the installation instructions on download page to learn about `GOPATH`. Then run following command:

    go get -u github.com/habitat-sh/habitat-operator/cmd/habitat-operator

This will put the built binary in `$GOPATH/bin`, make sure this is in your `PATH`, so you can access the binary from anywhere.

## Building manually from source directory

Clone the code locally:

    go get -u github.com/habitat-sh/habitat-operator/cmd/habitat-operator
    cd ${GOPATH:-$HOME/go}/src/github.com/habitat-sh/habitat-operator


Then build it:

    make build

This command will create a `habitat-operator` binary in the source directory. Copy this file somewhere to your `PATH`.

## Usage

### Running outside of a Kubernetes cluster

Start the Habitat operator by running:

    habitat-operator --kubeconfig ~/.kube/config

### Running inside a Kubernetes cluster

#### Building image from source

First build the image:

    make image

This will produce a `habitat/habitat-operator` image, which can then be deployed to your cluster.

The name of the generated docker image can be changed with an `IMAGE` variable, for example `make image IMAGE=mycorp/my-habitat-operator`. If the `habitat-operator` name is fine, then a `REPO` variable can be used like `make image REPO=mycorp` to generate the `mycorp/habitat-operator` image. Use the `TAG` variable to change the tag to something else (the default value is taken from `git describe --tags --always`) and a `HUB` variable to avoid using the default docker hub.

#### Using release image

Habitat operator images are located [here](https://hub.docker.com/r/habitat/habitat-operator/), they are tagged with the release version.

#### Deploying Habitat operator

##### Cluster with RBAC enabled

Make sure to give Habitat operator the correct permissions, so it's able to create and monitor certain resources. To do it, use the manifest files located under the examples directory:

    kubectl create -f examples/rbac

For more information see [the README file in RBAC example](examples/rbac/)

##### Cluster with RBAC disabled

To deploy the operator inside the Kubernetes cluster use the Deployment manifest file located under the examples directory:

    kubectl create -f examples/habitat-operator-deployment.yml

### Deploying an example

To create an example service run:

    kubectl create -f examples/standalone/habitat.yml

This will create a single-pod deployment of an `nginx` Habitat service.

More examples are located in the [example directory](examples/).

## Contributing

### Dependency management

This project uses [go dep](https://github.com/golang/dep/) `>= v0.4.1` for dependency management.

If you add, remove or change an import, run:

    dep ensure

### Testing

To run unit tests locally, run:

    make test

Clean up after the tests with:

    make clean-test

Our current setup does not allow e2e tests to run locally. It is best run on a [CI setup with Google Cloud](/doc/ci-gcp-setup.md).

### Code generation

If you change one of the types in `pkg/apis/habitat/v1beta1/types.go`, run the code generation script with:

    make codegen

[crd]: https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/
