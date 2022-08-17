package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	namespace string
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
		run("create registry", "k3d", "registry", "create", "registry.localhost", "--port", "12345")
		run("create k8s cluster", "k3d", "cluster", "create", "local", "--registry-use", "k3d-registry.localhost:12345")
		run("switch k8s context", "kubectl", "config", "use-context", "k3d-local")
	// build and upload image to local registry
	case "build":
		context := "../../../../chainlink" // TODO: make this an arg
		run("build image", "docker", "build", "-f", context+"/core/chainlink.Dockerfile", context, "-t", "chainlink:local")
		run("tag image", "docker", "tag", "chainlink:local", "localhost:12345/chainlink:local")
		run("push image", "docker", "push", "localhost:12345/chainlink:local")
	// run ginkgo commands to spin up environment
	case "run":
		os.Chdir("../../../") // move to repo root
		run("start environment", "ginkgo", "-r", "--focus", "@ocr", "integration-tests/smoke", "--", "--chainlink-image", "k3d-registry.localhost:12345/chainlink", "--chainlink-version", "local", "--keep-alive")
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

func run(name string, f string, args ...string) {
	fmt.Printf("\n-- %s --\n", strings.ToUpper(name))
	cmd := exec.Command(f, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()

	// stream output to cmd line
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		p := make([]byte, 100)
		for {
			n, err := stdout.Read(p)
			if err == io.EOF {
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
			if err == io.EOF {
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