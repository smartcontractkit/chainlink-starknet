package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/smartcontractkit/chainlink-starknet/ops/utils"
)

// TODO: consider extracting entire file to `chainlink-relay/ops` or into a separate CLI tool
// this simply runs underlying functions but does not import them

// PODMAN:
// ~/.config/containers/registries.conf
// [[registry]]
// location = "localhost:12345"
// insecure = true

func main() {
	if len(os.Args) < 2 {
		panic("missing required command")
	}

	switch strings.ToLower(os.Args[1]) {
	// create k8s cluster + resources
	case "create":
		run("create registry", "k3d", "registry", "create", "registry.localhost", "--port", "127.0.0.1:12345")
		run("create k8s cluster", "k3d", "cluster", "create", "local", "--api-port", "127.0.0.1:12346", "--registry-use", "k3d-registry.localhost:12345")
		run("switch k8s context", "kubectl", "config", "use-context", "k3d-local")
	// build and upload image to local registry
	case "build":
		context := "../../../chainlink" // TODO: make this an arg
		run("build image", "docker", "build", "-f", context+"/core/chainlink.Dockerfile", context, "-t", "chainlink:local")
		run("tag image", "docker", "tag", "chainlink:local", "localhost:12345/chainlink:local")
		run("push image", "docker", "push", "localhost:12345/chainlink:local")
	// run ginkgo commands to spin up environment
	case "run":
		// move to repo root
		if err := os.Chdir(utils.IntegrationTestsRoot); err != nil {
			panic(err)
		}
		setEnvIfNotExists("CHAINLINK_IMAGE", "k3d-registry.localhost:12345/chainlink")
		setEnvIfNotExists("CHAINLINK_VERSION", "local")
		setEnvIfNotExists("KEEP_ENVIRONMENTS", "ALWAYS")
		setEnvIfNotExists("NODE_COUNT", "1")
		setEnvIfNotExists("TTL", "900h")
		run("start environment", "go", "test", "-count", "1", "-v", "-timeout", "30m", "--run", "^TestOCRBasic$", "./smoke")
	// stop k8s namespace from environment
	case "stop":
		if len(os.Args) < 3 {
			panic("missing namespace argument")
		}
		run("stopping environment", "kubectl", "delete", "namespaces", os.Args[2])
	// delete removes the k8s cluster
	case "delete":
		run("remove k8s cluster", "k3d", "cluster", "delete", "local")
		run("remove registry", "k3d", "registry", "delete", "k3d-registry.localhost")
	default:
		panic("unrecognized command")
	}
}

func setEnvIfNotExists(key, defaultValue string) {
	value := os.Getenv(key)
	if value == "" {
		os.Setenv(key, defaultValue)
		value = defaultValue
	}
	fmt.Printf("Using %s=%s\n", key, value)
}

func run(name string, f string, args ...string) {
	fmt.Printf("\n-- %s --\n", strings.ToUpper(name))
	cmd := exec.Command(f, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	// stream output to cmd line
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		p := make([]byte, 100)
		for {
			n, err := stdout.Read(p)
			if errors.Is(err, io.EOF) {
				wg.Done()
				break
			}
			fmt.Print(string(p[:n]))
		}
	}()
	go func() {
		p := make([]byte, 100)
		for {
			n, err := stderr.Read(p)
			if errors.Is(err, io.EOF) {
				wg.Done()
				break
			}
			fmt.Print(string(p[:n]))
		}
	}()

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	wg.Wait()
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}
