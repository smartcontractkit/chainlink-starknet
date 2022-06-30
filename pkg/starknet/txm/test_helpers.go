package txm

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupLocalStarkNetNode sets up a local starknet node via cli, and returns the url
func SetupLocalStarkNetNode(t *testing.T) string {
	port := mustRandomPort()
	url := "http://127.0.0.1:" + port
	cmd := exec.Command("starknet-devnet",
		"--seed", "0", // use same seed for testing
		"--port", port,
	)
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		assert.NoError(t, cmd.Process.Kill())
		if err2 := cmd.Wait(); assert.Error(t, err2) {
			if !assert.Contains(t, err2.Error(), "signal: killed", cmd.ProcessState.String()) {
				t.Log("starknet-devnet stderr:", stdErr.String())
			}
		}
	})

	// Wait for api server to boot
	var ready bool
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)
		res, err := http.Get(url + "/is_alive")
		if err != nil || res.StatusCode != 200 {
			t.Logf("API server not ready yet (attempt %d)\n", i+1)
			continue
		}
		ready = true
		break
	}
	require.True(t, ready)
	return url
}

func mustRandomPort() string {
	r, err := rand.Int(rand.Reader, big.NewInt(65535-1023))
	if err != nil {
		panic(fmt.Errorf("unexpected error generating random port: %w", err))
	}
	return strconv.Itoa(int(r.Int64() + 1024))
}
