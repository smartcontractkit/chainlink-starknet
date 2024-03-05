package starknet

import (
	"context"
	"net/http"
	"testing"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewTestClient returns a client and a function to close a test server.
func NewTestClient(t *testing.T) *FeederClient {
	srv := NewTestServer()
	t.Cleanup(srv.Close)

	c := NewFeederClient(srv.URL).WithBackoff(NopBackoff).WithMaxRetries(0)
	c.client = &http.Client{
		Transport: &http.Transport{
			// On macOS tests often fail with the following error:
			//
			// "Get "http://127.0.0.1:xxxx/get_{feeder gateway method}?{arg}={value}": dial tcp 127.0.0.1:xxxx:
			//    connect: can't assign requested address"
			//
			// This error makes running local tests, in quick succession, difficult because we have to wait for the OS to release ports.
			// Sometimes the sync tests will hang because sync process will keep making requests if there was some error.
			// This problem is further exacerbated by having parallel tests.
			//
			// Increasing test client's idle conns allows for large concurrent requests to be made from a single test client.
			MaxIdleConnsPerHost: 1000,
		},
	}
	return c
}
func TestFeederClient(t *testing.T) {
	client := NewTestClient(t)
	tx, err := client.TransactionFailure(context.TODO(), &felt.Zero)
	require.NoError(t, err)

	// test server will return this for a transaction failure
	// so, the test server will never return a nonce error
	assert.Equal(t, tx.Code, "SOME_ERROR")
	assert.Equal(t, tx.ErrorMessage, "some error was encountered")
}
