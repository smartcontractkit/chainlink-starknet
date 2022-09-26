# Simple Local Env CLI

## Prerequisites
```
k3d - local kubernetes cluster
docker - build local containers
kubectl - interact with kubernetes cluster
```

from the root of the repo, install ginkgo
```
make install
```

make sure gauntlet and contract artifacts are properly compiled
```
yarn
make build-ts-contracts
```

## Commands
```bash

# starts up a k3d environment with a local registry
go run main.go create

# builds a local chainlink image
# note: assumes a folder structure like Documents/chainlink + Documents/chainlink-starknet
go run main.go build

# runs the ginkgo testing environment using the locally compiled image
# note: the namespace is in the first few lines of logging
go run main.go run

# removes the pods that were spun up during the `run` command
go run main.go stop <namespace>

# removes the entire k3d environment & registry
go run main.go delete
```

## Useful Commands & Tools
```bash
# pull logs from chainlink node 
# example: kubectl logs --namespace chainlink-smoke-ocr-starknet-ci-4553f chainlink-0-d79496974-kzczg -c node
# container is not needed if inspecting pod with single container (like starknet-devnet)
kubectl logs --namespace <namespace> <pod> -c <container>
```

Useful kubernetes cluster explorer - [Lens](https://k8slens.dev/) or [k9s](https://k9scli.io/)(lite weight, CLI based)
