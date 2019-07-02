# Runtime binding

This demonstrates how to run two Biome Services with a [binding](https://www.habitat.sh/docs/run-packages-binding/) between them.

## Workflow

After the Biome operator is up and running, execute the following command from the root of this repository:

```
kubectl create -f examples/bind/biome.yml
```

This will deploy two `Biome`s, a simple HTTP server written in Go that will be bound to a Redis database. The Go server will display the port number the database listens on.

The web app is listening on port `30001`. When running on minikube, its IP can
be retrieved with `minikube ip`.
