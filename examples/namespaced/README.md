# Namespaced Biome example

This demonstrates how to deploy service in a [Kubernetes namespace](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/).

## Workflow

Simply run:

 `kubectl create -f examples/namespaced/biome.yml`.

Note that any Secrets used by a `Biome` must be in the same namespace as the `Biome` itself.
