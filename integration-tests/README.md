# Local k8s run

Make sure to have `psql` installed locally. We use it to create a new database for each node.

Create a new network for containers (only needs to be done once). A custom network allows containers to DNS resolve each other using container names.

```
docker network create chainlink
```

Build a custom core image with starknet relayer bumped to some commit.

```
cd ../core
go get github.com/smartcontractkit/chainlink-starknet/relayer@<MY COMMIT HERE>
docker build . -t smartcontract/chainlink:starknet -f ./core/chainlink.Dockerfile
```

Compile contracts and gauntlet:

```
yarn build
cd contracts
scarb --profile release build
```

Run the tests!

```
cd integration-tests
go test -count 1 -v -timeout 30m --run OCRBasic ./smoke
```

Cleanup is broken right now, so use `something.down.sh` scripts to teardown everything afterwards.

# Old docs

For more information, see the [Chainlink Starknet Documentation | Integration Tests](../docs/integration-tests).
