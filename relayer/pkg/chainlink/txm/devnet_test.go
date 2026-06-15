package txm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchDevnetAccounts(t *testing.T) {
	t.Parallel()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rpc", r.URL.Path)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"private_key":"0x1","public_key":"0x2","address":"0x3"}]}`))
	}))
	defer mockServer.Close()

	accounts, err := FetchDevnetAccounts(mockServer.URL)
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	assert.Equal(t, "0x1", accounts[0].PrivateKey)
	assert.Equal(t, "0x2", accounts[0].PublicKey)
	assert.Equal(t, "0x3", accounts[0].Address)
}

func TestDevnetMint(t *testing.T) {
	t.Parallel()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rpc", r.URL.Path)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"new_balance":900000000000000000,"unit":"FRI","tx_hash":"0xabc"}}`))
	}))
	defer mockServer.Close()

	res, err := DevnetMint(mockServer.URL, "0x123", 900000000000000000, "FRI")
	require.NoError(t, err)
	assert.Contains(t, res, `"tx_hash":"0xabc"`)
}
