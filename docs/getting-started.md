# Getting Started

## Local setup

```bash
make install
```

## Go Modules Dependency

```mermaid
flowchart LR
  subgraph ops
  test-helpers
  k8s
  end

  subgraph relayer
  contract-bindings
  end

  subgraph e2e-tests
  smoke
  soak
  end

  contract-bindings --> e2e-tests
  test-helpers --> relayer
  k8s --> e2e-tests
```
