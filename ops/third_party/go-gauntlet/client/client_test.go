package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestPostExecuteWithResponsePollsUntilOutput(t *testing.T) {
	t.Parallel()

	var pollCount atomic.Int32
	reportID := "test-report-id"
	output := map[string]any{"contractAddress": "0x123"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/execute":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Report{Id: reportID})
		case "/reports":
			count := pollCount.Add(1)
			w.Header().Set("Content-Type", "application/json")
			if count < 2 {
				_ = json.NewEncoder(w).Encode(map[string]Report{
					reportID: {Id: reportID},
				})
				return
			}
			out := any(output)
			_ = json.NewEncoder(w).Encode(map[string]Report{
				reportID: {Id: reportID, Output: &out},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClientWithResponses(server.URL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}

	args := any(map[string]any{})
	response, err := client.PostExecuteWithResponse(context.Background(), &PostExecuteParams{}, PostExecuteJSONRequestBody{
		Config: &Config{},
		Operation: Operation{
			Args: &args,
			Name: "starknet/chain/open-zeppelin:deploy",
		},
	})
	if err != nil {
		t.Fatalf("PostExecuteWithResponse: %v", err)
	}

	if response.JSON200 == nil || response.JSON200.Output == nil {
		t.Fatal("expected polled report output")
	}

	outputMap, ok := (*response.JSON200.Output).(map[string]any)
	if !ok {
		t.Fatalf("expected map output, got %T", *response.JSON200.Output)
	}

	if outputMap["contractAddress"] != "0x123" {
		t.Fatalf("unexpected contractAddress: %v", outputMap["contractAddress"])
	}

	if pollCount.Load() < 2 {
		t.Fatalf("expected at least 2 poll attempts, got %d", pollCount.Load())
	}
}

func TestPostExecuteWithResponseReturnsReportError(t *testing.T) {
	t.Parallel()

	reportID := "failed-report-id"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/execute":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Report{Id: reportID})
		case "/reports":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]Report{
				reportID: {
					Id: reportID,
					Error: &Error{
						Code:    "EXECUTION_FAILED",
						Message: "declare failed",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClientWithResponses(server.URL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}

	args := any(map[string]any{})
	_, err = client.PostExecuteWithResponse(context.Background(), &PostExecuteParams{}, PostExecuteJSONRequestBody{
		Config: &Config{},
		Operation: Operation{
			Args: &args,
			Name: "starknet/chain/open-zeppelin:declare",
		},
	})
	if err == nil {
		t.Fatal("expected error from failed report")
	}
}
