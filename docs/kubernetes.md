# Kubernetes

We run our software in Kubernetes.

### Local k3d setup

1. `make install`
2. (Optional) Install `Lens` from [here](https://k8slens.dev/) or use `k9s` as a low resource consumption alternative from [here](https://k9scli.io/topics/install/)
   or from source [here](https://github.com/smartcontractkit/helmenv)
3. Setup your docker resources, 6vCPU/10Gb RAM are enough for most CL related tasks
4. `k3d cluster create local`
5. Check your contexts with `kubectl config get-contexts`
6. Switch context `kubectl config use-context k3d-local`
7. Run any tests, use a guide [here](integration-tests/README.md)
8. Stop the cluster

```
k3d cluster stop local
```
