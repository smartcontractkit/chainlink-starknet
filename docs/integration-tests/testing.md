## Integration tests usage

Setup k8s context, if you don't have k8s, spin up a local cluster using [this](../kubernetes.md) guide

### Run tests using ephemeral envs

```
make e2e_test
```

### Run tests on a standalone local env

1. Spin up an env, for example, see yaml file for more options with a stark-devnet/pathfinder real node

```
envcli new -p ops/chainlink-starknet.yaml
```

2. Check created file in a previous command output, example `Environment setup and written to file environmentFile=chainlink-stark-k42hp.yaml`
3. Run the tests

```
ENVIRONMENT_FILE="$(pwd)/chainlink-stark-k42hp.yaml" KEEP_ENVIRONMENTS="Always" make e2e_test
```

4. Check the env file or connect command logs for a forwarded `local_ports` and try it in the browser
5. Destroy the env

```
envcli rm -e chainlink-stark-b7mt9.yaml
```

### Interact with an env using other scripts

1. Spin up an env, for example, see yaml file for more options with a stark-devnet/pathfinder real node

```
envcli new -p ops/chainlink-starknet.yaml
```

2. Check created file in a previous command output, example `Environment setup and written to file environmentFile=chainlink-stark-mx7rg.yaml`
3. Connect to your env

```
envcli connect -e ${your_env_file_yaml}
```

4. Check the env file or connect command logs for a forwarded `local_ports` and try it in the browser
5. Interact using other scripts
6. Destroy the env

```
envcli rm -e chainlink-stark-b7mt9.yaml
```
