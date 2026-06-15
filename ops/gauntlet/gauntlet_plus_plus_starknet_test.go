package gauntlet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	g "github.com/smartcontractkit/gauntlet-plus-plus/sdks/go-gauntlet/client"
)

func newTestStarknetGauntletPlusPlus(t *testing.T, serverURL string) *StarknetGauntletPlusPlus {
	t.Helper()

	client, err := g.NewClientWithResponses(serverURL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}

	providers := []g.Provider{}
	return &StarknetGauntletPlusPlus{
		client:    client,
		providers: &providers,
	}
}

func TestExecuteReturnsReportPollsUntilOutput(t *testing.T) {
	t.Parallel()

	var pollCount atomic.Int32
	reportID := "test-report-id"
	output := map[string]any{"contractAddress": "0x123"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/execute":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(g.Report{Id: reportID})
		case "/reports":
			count := pollCount.Add(1)
			w.Header().Set("Content-Type", "application/json")
			if count < 2 {
				_ = json.NewEncoder(w).Encode(map[string]g.Report{
					reportID: {Id: reportID},
				})
				return
			}
			out := any(output)
			_ = json.NewEncoder(w).Encode(map[string]g.Report{
				reportID: {Id: reportID, Output: &out},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	sgpp := newTestStarknetGauntletPlusPlus(t, server.URL)
	args := map[string]any{}
	report, err := sgpp.executeReturnsReport(&Request{
		Command: "starknet/chain/open-zeppelin:deploy",
		Input:   args,
	})
	if err != nil {
		t.Fatalf("executeReturnsReport: %v", err)
	}

	if report.Output == nil {
		t.Fatal("expected polled report output")
	}

	outputMap, ok := (*report.Output).(map[string]any)
	if !ok {
		t.Fatalf("expected map output, got %T", *report.Output)
	}

	if outputMap["contractAddress"] != "0x123" {
		t.Fatalf("unexpected contractAddress: %v", outputMap["contractAddress"])
	}

	if pollCount.Load() < 2 {
		t.Fatalf("expected at least 2 poll attempts, got %d", pollCount.Load())
	}
}

func TestExecuteReturnsReportReturnsImmediateReportError(t *testing.T) {
	t.Parallel()

	reportID := "failed-report-id"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/execute" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(g.Report{
			Id: reportID,
			Error: &g.Error{
				Code:    "OPERATION_ERROR",
				Message: "Invalid block ID",
			},
		})
	}))
	defer server.Close()

	sgpp := newTestStarknetGauntletPlusPlus(t, server.URL)
	args := map[string]any{}
	_, err := sgpp.executeReturnsReport(&Request{
		Command: "starknet/chain/open-zeppelin:declare",
		Input:   args,
	})
	if err == nil {
		t.Fatal("expected error from immediate failed report")
	}
}

func TestExecuteReturnsReportReturnsPolledReportError(t *testing.T) {
	t.Parallel()

	reportID := "failed-report-id"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/execute":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(g.Report{Id: reportID})
		case "/reports":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]g.Report{
				reportID: {
					Id: reportID,
					Error: &g.Error{
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

	sgpp := newTestStarknetGauntletPlusPlus(t, server.URL)
	args := map[string]any{}
	_, err := sgpp.executeReturnsReport(&Request{
		Command: "starknet/chain/open-zeppelin:declare",
		Input:   args,
	})
	if err == nil {
		t.Fatal("expected error from failed report")
	}
}

func TestPollGauntletReportRespectsContextCancellation(t *testing.T) {
	t.Parallel()

	reportID := "test-report-id"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/reports" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]g.Report{
			reportID: {Id: reportID},
		})
	}))
	defer server.Close()

	sgpp := newTestStarknetGauntletPlusPlus(t, server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := sgpp.pollGauntletReport(ctx, reportID)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
