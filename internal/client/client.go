package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client communicates with the FormFor REST API.
type Client struct {
	APIKey  string
	BaseURL string
	HTTP    *http.Client
}

// New creates a Client with sensible defaults.
func New(apiKey, baseURL string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateForm creates a new form with the given parameters.
func (c *Client) CreateForm(params CreateFormParams) (*Form, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	resp, err := c.do("POST", "/v1/forms", body)
	if err != nil {
		return nil, err
	}

	var form Form
	if err := json.Unmarshal(resp, &form); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &form, nil
}

// Ask is a shorthand that creates a yes/no approval form.
func (c *Client) Ask(question string, to string, opts AskOptions) (*Form, error) {
	params := CreateFormParams{
		Title:     question,
		Recipient: to,
		ExpiresIn: opts.Expires,
		Context:   opts.Context,
		Fields: []Field{
			{
				ID:       "approved",
				Type:     "boolean",
				Label:    question,
				Required: true,
			},
		},
	}
	return c.CreateForm(params)
}

// GetForm retrieves full form details by ID.
func (c *Client) GetForm(id string) (*FormDetail, error) {
	resp, err := c.do("GET", "/v1/forms/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}

	var detail FormDetail
	if err := json.Unmarshal(resp, &detail); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &detail, nil
}

// GetResponse retrieves the submitted response for a form.
func (c *Client) GetResponse(formID string) (*FormResponse, error) {
	resp, err := c.do("GET", "/v1/forms/"+url.PathEscape(formID)+"/response", nil)
	if err != nil {
		return nil, err
	}

	var fr FormResponse
	if err := json.Unmarshal(resp, &fr); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &fr, nil
}

// ListForms lists forms with optional filters.
func (c *Client) ListForms(opts ListOptions) (*ListResult, error) {
	q := url.Values{}
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}

	path := "/v1/forms"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result ListResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// CancelForm cancels a pending form.
func (c *Client) CancelForm(id string) error {
	_, err := c.do("POST", "/v1/forms/"+url.PathEscape(id)+"/cancel", nil)
	return err
}

// RemindForm sends a reminder for a pending form.
func (c *Client) RemindForm(id string) error {
	_, err := c.do("POST", "/v1/forms/"+url.PathEscape(id)+"/remind", nil)
	return err
}

// WaitForResponse polls every 2 seconds until the form has a response or the timeout elapses.
func (c *Client) WaitForResponse(formID string, timeout time.Duration) (*FormResponse, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		detail, err := c.GetForm(formID)
		if err != nil {
			return nil, fmt.Errorf("polling form status: %w", err)
		}

		switch detail.Status {
		case "completed":
			if detail.Response != nil {
				return detail.Response, nil
			}
			// Response embedded; fetch it explicitly.
			return c.GetResponse(formID)
		case "expired":
			return nil, fmt.Errorf("form %s expired before receiving a response", formID)
		case "cancelled":
			return nil, fmt.Errorf("form %s was cancelled", formID)
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for response to form %s", formID)
		}

		<-ticker.C
	}
}

// do executes an HTTP request against the API and returns the response body.
func (c *Client) do(method, path string, body []byte) ([]byte, error) {
	u := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "formfor-cli/0.1.0")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp apiErrorResponse
		if jsonErr := json.Unmarshal(data, &errResp); jsonErr == nil && errResp.Error.Message != "" {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Code:       errResp.Error.Code,
				Message:    errResp.Error.Message,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(data)),
		}
	}

	return data, nil
}
