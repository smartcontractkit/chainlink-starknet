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

type Report struct {
	Id     string `json:"id"`
	Output *any   `json:"output,omitempty"`
}

type PostExecuteParams struct{}

type PostExecuteResponse struct {
	JSON200 *Report
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
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal execute request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/execute", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build execute request: %w", err)
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
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read execute response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	report := &Report{}
	if len(respBody) > 0 {
		if err = json.Unmarshal(respBody, report); err != nil {
			return nil, fmt.Errorf("decode execute response: %w", err)
		}
	}

	return &PostExecuteResponse{JSON200: report}, nil
}
