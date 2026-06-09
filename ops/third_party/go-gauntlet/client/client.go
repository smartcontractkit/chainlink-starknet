package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultPollInterval = 2 * time.Second
	defaultPollTimeout  = 5 * time.Minute
)

type Provider struct {
	Name  string                  `json:"name"`
	Type  string                  `json:"type"`
	Input map[string]*interface{} `json:"input,omitempty"`
}

type Datasource struct{}

type Config struct {
	Providers   []Provider   `json:"providers"`
	Datasources []Datasource `json:"datasources"`
}

type Operation struct {
	Args *any   `json:"args,omitempty"`
	Name string `json:"name"`
}

type PostExecuteJSONRequestBody struct {
	Config    *Config   `json:"config"`
	Operation Operation `json:"operation"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Report struct {
	Id     string `json:"id"`
	Output *any   `json:"output,omitempty"`
	Error  *Error `json:"error,omitempty"`
}

type PostReportsJSONRequestBody struct {
	Ids []string `json:"ids"`
}

type PostExecuteParams struct{}

type PostExecuteResponse struct {
	JSON200 *Report
}

type PostReportsResponse struct {
	JSON200 *map[string]Report
}

type RequestEditorFn func(*http.Request) error

type ClientWithResponses struct {
	baseURL    string
	httpClient *http.Client
}

func NewClientWithResponses(server string) (*ClientWithResponses, error) {
	baseURL := strings.TrimRight(server, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("server URL is empty")
	}

	return &ClientWithResponses{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}, nil
}

func (c *ClientWithResponses) PostExecuteWithResponse(ctx context.Context, _ *PostExecuteParams, body PostExecuteJSONRequestBody, reqEditors ...RequestEditorFn) (*PostExecuteResponse, error) {
	report, err := c.postJSON(ctx, http.MethodPost, "/execute", body, reqEditors...)
	if err != nil {
		return nil, err
	}

	if report == nil || report.Id == "" {
		return &PostExecuteResponse{JSON200: report}, nil
	}

	if report.Output != nil || report.Error != nil {
		return &PostExecuteResponse{JSON200: report}, nil
	}

	polled, err := c.pollReport(ctx, report.Id, reqEditors...)
	if err != nil {
		return nil, err
	}

	return &PostExecuteResponse{JSON200: polled}, nil
}

func (c *ClientWithResponses) PostReportsWithResponse(ctx context.Context, body PostReportsJSONRequestBody, reqEditors ...RequestEditorFn) (*PostReportsResponse, error) {
	reports, err := c.fetchReports(ctx, body.Ids, reqEditors...)
	if err != nil {
		return nil, err
	}

	return &PostReportsResponse{JSON200: reports}, nil
}

func (c *ClientWithResponses) pollReport(ctx context.Context, reportID string, reqEditors ...RequestEditorFn) (*Report, error) {
	deadline := time.Now().Add(defaultPollTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		reports, err := c.fetchReports(ctx, []string{reportID}, reqEditors...)
		if err != nil {
			return nil, err
		}

		if reports == nil {
			time.Sleep(defaultPollInterval)
			continue
		}

		report, ok := (*reports)[reportID]
		if !ok {
			time.Sleep(defaultPollInterval)
			continue
		}

		if report.Error != nil {
			return nil, fmt.Errorf("gauntlet++ report %s failed: %s (%s)", reportID, report.Error.Message, report.Error.Code)
		}

		if report.Output != nil {
			return &report, nil
		}

		time.Sleep(defaultPollInterval)
	}

	return nil, fmt.Errorf("timed out waiting for gauntlet++ report %s after %s", reportID, defaultPollTimeout)
}

func (c *ClientWithResponses) fetchReports(ctx context.Context, ids []string, reqEditors ...RequestEditorFn) (*map[string]Report, error) {
	body := PostReportsJSONRequestBody{Ids: ids}
	respBody, err := c.doRequest(ctx, http.MethodPost, "/reports", body, reqEditors...)
	if err != nil {
		return nil, err
	}

	reports := make(map[string]Report)
	if len(respBody) == 0 {
		return &reports, nil
	}

	if err = json.Unmarshal(respBody, &reports); err != nil {
		return nil, fmt.Errorf("decode reports response: %w", err)
	}

	return &reports, nil
}

func (c *ClientWithResponses) postJSON(ctx context.Context, method, path string, body any, reqEditors ...RequestEditorFn) (*Report, error) {
	respBody, err := c.doRequest(ctx, method, path, body, reqEditors...)
	if err != nil {
		return nil, err
	}

	report := &Report{}
	if len(respBody) > 0 {
		if err = json.Unmarshal(respBody, report); err != nil {
			return nil, fmt.Errorf("decode execute response: %w", err)
		}
	}

	return report, nil
}

func (c *ClientWithResponses) doRequest(ctx context.Context, method, path string, body any, reqEditors ...RequestEditorFn) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	for _, editor := range reqEditors {
		if editor == nil {
			continue
		}
		if err = editor(req); err != nil {
			return nil, fmt.Errorf("apply request editor: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
