package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bookshop/api/internal/pkg/ratelimit"
	"github.com/bookshop/api/pkg/logger"
)

// APIClient represents a client for working with external APIs
type APIClient struct {
	client      *http.Client
	baseURL     string
	rateLimiter *ratelimit.RateLimiter
	logger      logger.Logger
}

// APIClientConfig contains parameters for creating an API client
type APIClientConfig struct {
	BaseURL      string
	Timeout      time.Duration
	RateLimit    int  // Requests per second
	RetryCount   int  // Number of retry attempts
	RetryEnabled bool // Enable automatic request retry
}

// DefaultConfig returns the default configuration
func DefaultConfig() APIClientConfig {
	return APIClientConfig{
		Timeout:      10 * time.Second,
		RateLimit:    5, // 5 requests per second
		RetryCount:   3,
		RetryEnabled: true,
	}
}

// NewAPIClient creates a new client for external APIs
func NewAPIClient(config APIClientConfig, logger logger.Logger) *APIClient {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &APIClient{
		client:      client,
		baseURL:     config.BaseURL,
		rateLimiter: ratelimit.NewRateLimiter(config.RateLimit),
		logger:      logger,
	}
}

// Get performs a GET request to the API with rate limiting
func (c *APIClient) Get(ctx context.Context, path string) (*http.Response, error) {
	url := c.baseURL + path
	var response *http.Response
	var err error

	// Process request through rate limiter
	limiterErr := c.rateLimiter.Process(func() {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if reqErr != nil {
			err = fmt.Errorf("error creating request: %w", reqErr)
			return
		}

		// Set headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "BookshopAPI Client")

		// Execute request
		response, err = c.client.Do(req)
		if err != nil {
			err = fmt.Errorf("error making request: %w", err)
			return
		}
	})

	if limiterErr != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", limiterErr)
	}

	return response, err
}

// GetJSON performs a GET request and decodes the JSON response
func (c *APIClient) GetJSON(ctx context.Context, path string, result interface{}) error {
	response, err := c.Get(ctx, path)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check status code
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("API returned non-OK status: %d, body: %s",
			response.StatusCode, string(body))
	}

	// Decode JSON response
	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return fmt.Errorf("error decoding JSON: %w", err)
	}

	return nil
}

// Post performs a POST request to the API with rate limiting
func (c *APIClient) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path
	var response *http.Response
	var err error

	// Process request through rate limiter
	limiterErr := c.rateLimiter.Process(func() {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
		if reqErr != nil {
			err = fmt.Errorf("error creating request: %w", reqErr)
			return
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "BookshopAPI Client")

		// Execute request
		response, err = c.client.Do(req)
		if err != nil {
			err = fmt.Errorf("error making request: %w", err)
			return
		}
	})

	if limiterErr != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", limiterErr)
	}

	return response, err
}

// PostJSON performs a POST request with JSON data and decodes the JSON response
func (c *APIClient) PostJSON(ctx context.Context, path string,
	requestBody interface{}, responseResult interface{}) error {

	// Encode request body to JSON
	var bodyReader io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("error encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	response, err := c.Post(ctx, path, bodyReader)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check status code
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("API returned non-success status: %d, body: %s",
			response.StatusCode, string(body))
	}

	// If response is expected and status is not 204 No Content
	if responseResult != nil && response.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(response.Body).Decode(responseResult); err != nil {
			return fmt.Errorf("error decoding JSON response: %w", err)
		}
	}

	return nil
}

// Close releases resources used by the client
func (c *APIClient) Close() {
	c.rateLimiter.Stop()
}

// Helper function to read response as string
func readResponseBody(resp *http.Response) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("nil response")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(body), nil
}
