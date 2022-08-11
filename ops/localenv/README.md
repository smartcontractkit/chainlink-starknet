# Simple Local Env CLI

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
