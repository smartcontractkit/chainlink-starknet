package utils

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	// ProjectRoot Root folder of this project
	ProjectRoot = filepath.Join(filepath.Dir(b), "/../..")
	// SolanaTestsRoot path to starknet e2e tests
	IntegrationTestsRoot = filepath.Join(ProjectRoot, "integration-tests")
	// OpsRoot path to ops folder
	OpsRoot = filepath.Join(ProjectRoot, "ops")
)
