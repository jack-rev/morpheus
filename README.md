# morpheus
K8s multi-pod log tailing. Inspired by the stern project.

## Getting started

Ensure that you have `go` version 1.19 or higher installed. To build the `morpheus` executable, simply run `go build`, then run `./morpheus --namespace [namespace]`. Morpheus picks up
existing pods in the namespace specified, and listens for new pods to be created. If no namespace is specified, then `default` will be used
