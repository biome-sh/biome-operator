# Configuration via environment variables

This example demonstrates how environment variables can be used for initial service configuration.

## Workflow

After the Biome operator is up and running, execute the following command from the root of this repository:

    kubectl create -f examples/env-vars/biome.yml

By default, Redis listens on port 6379, but we change this to 6999 by passing
the `HAB_REDIS` environment variable to the supervisor (the supervisor
automatically loads environment variables named `HAB_PACKAGENAME`, as explained
[here](https://www.habitat.sh/docs/using-biome/#config-updates)).

When you inspect the logs of the pod that was just created (with `kubectl logs
$podname`) you should see something like:

    redis.foobar(O): 42:M 07 Feb 15:34:08.765 * The server is now ready to accept connections on port 6999

## Environment Variables

You can use the full range of options defined by the Downward API. For an
overview, please refer to [this
guide](https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/).
