# Encrypted Biome example

By default supervisors will communicate with no encryption. This example demonstrates how to secure the communication.

## Workflow

The user needs to generate a [ring
key](https://www.habitat.sh/docs/run-packages-security/) using `bio ring key generate foobar`, and then base64
encode it with (on Linux) `bio ring export foobar | base64 -w 0` (please refer to
[this
document](https://kubernetes.io/docs/concepts/configuration/secret/#creating-a-secret-manually)
for platform-specific instructions on base64 encoding).

The encoded key can then be used as the value of the `ring-key` key in a Kubernetes
Secret.

The Secret's name must be the same as the key's filename, minus the
extension.

For example, for a key named `foobar`, the key filename might be something like
`foobar-20170824094632.sym.key`, and the corresponding Secret name
`foobar-20170824094632`.

The Secret's name must additionally be referenced in the `Biome` object's `ringSecretName` key.

After the Biome operator is up and running, execute the following command from the root of this repository:

```
kubectl create -f examples/encrypted/biome.yml
```

## Deletion

The Biome operator does not delete the Secret on Biome deletion. This is
because the user might want to re-use the secret across multiple
`Biome`s and `Biome` lifecycles.
