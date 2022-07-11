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

	"github.com/dontpanicdao/caigo"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seed = 0 keys for starknet-devnet
var privateKeys0Seed []string = []string{
	"0xe3e70682c2094cac629f6fbed82c07cd",
	"0xf728b4fa42485e3a0a5d2f346baa9455",
	"0xeb1167b367a9c3787c65c1e582e2e662",
	"0xf7c1bd874da5e709d4713d60c8a70639",
	"0xe443df789558867f5ba91faf7a024204",
	"0x23a7711a8133287637ebdcd9e87a1613",
	"0x1846d424c17c627923c6612f48268673",
	"0xfcbd04c340212ef7cca5a5a19e4d6e3c",
	"0xb4862b21fb97d43588561712e8e5216a",
	"0x259f4329e6f4590b9a164106cf6a659e",
}

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
		t.Logf("API server ready at %s\n", url)
		break
	}
	require.True(t, ready)
	return url
}

func TestKeys(t *testing.T, count int) map[string]keys.Key {
	keyMap := map[string]keys.Key{}

	require.True(t, len(privateKeys0Seed) >= count, "requested more keys than available")
	for i, k := range privateKeys0Seed {
		// max number of keys to generate
		if i >= count {
			break
		}

		keyBytes, err := caigo.HexToBytes(k)
		require.NoError(t, err)
		raw := keys.Raw(keyBytes)

		keyMap[raw.Key().PublicKeyStr()] = raw.Key()

	}
	return keyMap
}

func mustRandomPort() string {
	r, err := rand.Int(rand.Reader, big.NewInt(65535-1023))
	if err != nil {
		panic(fmt.Errorf("unexpected error generating random port: %w", err))
	}
	return strconv.Itoa(int(r.Int64() + 1024))
}
