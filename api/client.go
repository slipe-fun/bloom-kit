package api

import (
	"context"
	"fmt"
	"sync"

	"resty.dev/v3"
)

type ApiError struct {
	Code    string
	Message string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("api error [%s]: %s", e.Code, e.Message)
}

type Client struct {
	httpClient *resty.Client
	baseURL    string
	token      string
	tokenMu    sync.RWMutex
}

func NewClient(baseURL string) *Client {
	restyClient := resty.New()

	restyClient.SetBaseURL(baseURL)

	client := &Client{
		httpClient: restyClient,
		baseURL:    baseURL,
	}

	restyClient.AddRequestMiddleware(func(cli *resty.Client, req *resty.Request) error {
		token := client.getToken()
		if token != "" {
			req.SetAuthToken(token)
		}
		return nil
	})

	return client
}

func (c *Client) SetToken(token string) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.token = token
}

func (c *Client) getToken() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token
}

func Send[Req any, Resp any](ctx context.Context, c *Client, method, path string, body *Req) (*Resp, error) {
	req := c.httpClient.R().SetContext(ctx)

	if body != nil {
		req.SetBody(body)
	}

	var respData Resp
	var apiError ApiError

	resp, err := req.
		SetResult(&respData).
		SetResultError(&apiError).
		Execute(method, path)

	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}

	if resp.IsStatusFailure() {
		if apiError.Code != "" {
			return nil, &apiError
		}
		return nil, fmt.Errorf("http error %d: %s", resp.StatusCode(), resp.String())
	}

	return &respData, nil
}
