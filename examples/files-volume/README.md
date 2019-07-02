# Runtime binding

This demonstrates how to provide the contents of the Biome Service files directory via a Kubernetes Secret.


## Workflow

After the Biome operator is up and running, execute the following command from the root of this repository:

```
kubectl create -f examples/files-volume/biome.yml
```

This will deploy redis and also put a file containing a password into `hab/svc/redis/files/pwfile`.

