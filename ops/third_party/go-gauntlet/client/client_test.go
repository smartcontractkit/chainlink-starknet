package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostExecuteWithResponseReturnsReport(t *testing.T) {
	t.Parallel()

	reportID := "test-report-id"
	output := map[string]any{"contractAddress": "0x123"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/execute" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		out := any(output)
		_ = json.NewEncoder(w).Encode(Report{Id: reportID, Output: &out})
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
		t.Fatal("expected report output")
	}
}

func TestPostReportsWithResponseReturnsReports(t *testing.T) {
	t.Parallel()

	reportID := "test-report-id"
	output := map[string]any{"contractAddress": "0x123"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/reports" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		out := any(output)
		_ = json.NewEncoder(w).Encode(map[string]Report{
			reportID: {Id: reportID, Output: &out},
		})
	}))
	defer server.Close()

	client, err := NewClientWithResponses(server.URL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}

	response, err := client.PostReportsWithResponse(context.Background(), PostReportsJSONRequestBody{
		Ids: []string{reportID},
	})
	if err != nil {
		t.Fatalf("PostReportsWithResponse: %v", err)
	}

	if response.JSON200 == nil {
		t.Fatal("expected reports map")
	}

	report, ok := (*response.JSON200)[reportID]
	if !ok || report.Output == nil {
		t.Fatal("expected report with output")
	}
}
