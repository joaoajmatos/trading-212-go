// Package trading212 provides a Go client for the Trading 212 REST API.
//
// # Authentication
//
// Create a client with your API key obtained from Trading 212 Settings → API:
//
//	client := trading212.New("your-api-key")
//
// By default the client targets the live environment. Use [WithDemo] to target
// the demo (paper trading) environment:
//
//	client := trading212.New("your-api-key", trading212.WithDemo())
//
// # Error handling
//
// All methods return an error on failure. HTTP-level errors (status >= 400) are
// returned as *[Error], which carries the HTTP status code and the error message
// from the API response body. Use [errors.As] to inspect them:
//
//	_, err := client.GetAccountSummary(ctx)
//	var apiErr *trading212.Error
//	if errors.As(err, &apiErr) && apiErr.StatusCode == 429 {
//	    // rate limited
//	}
//
// Network-level failures (e.g. connection refused) are returned as standard
// net/http errors and are never wrapped in *Error.
//
// # Pagination
//
// Endpoints that return lists support cursor-based pagination via [Cursor].
// Use the cursor like a scanner:
//
//	cur := client.HistoryOrders(trading212.HistoryOrdersParams{Limit: 50})
//	for cur.Next(ctx) {
//	    order := cur.Item()
//	    _ = order
//	}
//	if err := cur.Err(); err != nil {
//	    log.Fatal(err)
//	}
package trading212

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// LiveBaseURL is the base URL for the live (real money) environment.
	LiveBaseURL = "https://live.trading212.com/api/v0"

	// DemoBaseURL is the base URL for the demo (paper trading) environment.
	DemoBaseURL = "https://demo.trading212.com/api/v0"

	defaultTimeout = 30 * time.Second
)

// Client is the Trading 212 API client. The zero value is not useful; create
// one with [New].
type Client struct {
	apiKey      string
	baseURL     string
	httpClient  *http.Client
	rateLimiter *endpointRateLimiter
}

// Option configures a [Client].
type Option func(*Client)

// WithDemo configures the client to use the demo (paper trading) environment.
func WithDemo() Option {
	return func(c *Client) {
		c.baseURL = DemoBaseURL
	}
}

// WithHTTPClient replaces the default HTTP client. Use this to add custom
// transports (logging, tracing, retries) or to set a custom timeout.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithBaseURL overrides the API base URL. Intended for testing and non-standard
// deployments; most callers should use [WithDemo] instead.
func WithBaseURL(u string) Option {
	return func(c *Client) {
		c.baseURL = u
	}
}

// WithTimeout sets the request timeout on the default HTTP client. Has no
// effect if [WithHTTPClient] is also provided.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// New creates a new Trading 212 API client.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: LiveBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// do executes an HTTP request against the Trading 212 API.
//
//   - method: HTTP verb (GET, POST, DELETE, …)
//   - path: path starting with /, e.g. "/equity/account/summary"
//   - body: request body marshalled to JSON; nil for requests with no body
//   - result: pointer to the type to unmarshal the response into; nil to discard
//
// Returns the raw *http.Response so callers can inspect headers (e.g. rate
// limits). The response body is fully consumed and closed before returning.
func (c *Client) do(ctx context.Context, method, path string, body, result any) (*http.Response, error) {
	if c.rateLimiter != nil {
		if err := c.rateLimiter.wait(ctx, method, path); err != nil {
			return nil, err
		}
	}

	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("trading212: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("trading212: build request: %w", err)
	}

	req.Header.Set("Authorization", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap
	if err != nil {
		return resp, fmt.Errorf("trading212: read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &Error{
			StatusCode: resp.StatusCode,
			RateLimit:  RateLimitFromResponse(resp),
		}
		if jsonErr := json.Unmarshal(respBody, apiErr); jsonErr != nil {
			apiErr.Message = string(respBody)
		}
		return resp, apiErr
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return resp, fmt.Errorf("trading212: unmarshal response: %w", err)
		}
	}

	return resp, nil
}
