package cryptographermain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Here simple form at API Link
const (
	defaultBaseURL           = "https://api-etcetcetc"
	defaultRetry             = 3
	defaultTimeoutBetweenReq = 1 * time.Minute
)

var (
	// ErrNoResponse - error, if answer not bin delivery
	ErrNoResponse = errors.New("no response was received")

	// ErrNotConfigured - error, if not bin delivery http client
	ErrNotConfigured = errors.New("the client is not configured")

	// ErrWrongMethod - error, if method for request incorrect
	ErrWrongMethod = errors.New("wrong method")
)

// Client - structure client, where http.Client, baseURL - adress for client
// Retry - numbers of attempts, timeoutBetween - timeout of attempts
type Client struct {
	httpClient        *http.Client
	baseURL           string
	retry             uint
	timeoutBetweenReq time.Duration
}

// Options - type func, that change client settings
type Options func(*Client)

// NewClient - creating a client with default settings
func NewClient() *Client {
	client := newDefaultClient()
	return client
}

func newDefaultClient() *Client {
	client := http.DefaultClient
	return &Client{httpClient: client}
}

// NewClientWithOptions - creating a client with the addition of functions that change the client settings to the arguments,
// for example, WithBaseURL, WithTimeoutRequest, WithMaxRetry.
func NewClientWithOptions(opts ...Options) *Client {
	client := newDefaultClient()
	for _, o := range opts {
		o(client)
	}
	return client
}

func (c *Client) do(method string, body interface{}) (*http.Response, error) {
	if c.retry == 0 {
		c.retry = defaultRetry
	}
	if c.timeoutBetweenReq == 0 {
		c.timeoutBetweenReq = defaultTimeoutBetweenReq
	}
	for ; c.retry > 0; c.retry-- {
		req, err := c.request(method, body)
		if err != nil {
			return nil, err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			fmt.Println(err)
			time.Sleep(c.timeoutBetweenReq)
			continue
		}
		err = c.cheсkStatus(resp)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return nil, ErrNoResponse
}

func (c *Client) cheсkStatus(resp *http.Response) error {
	if resp.StatusCode <= http.StatusOK && resp.StatusCode > http.StatusMultipleChoices {
		return fmt.Errorf("unexpected response status: %v", resp.StatusCode)
	}
	return nil
}

func (c *Client) request(method string, body interface{}) (*http.Request, error) {
	if c.httpClient == nil {
		return nil, ErrNotConfigured
	}
	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}
	switch method {
	case http.MethodGet:
		req, err := http.NewRequest(method, c.baseURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	case http.MethodPost:
		buffer := new(bytes.Buffer)
		if body != nil {
			if err := json.NewEncoder(buffer).Encode(body); err != nil {
				return nil, err
			}
		}
		req, err := http.NewRequest(method, c.baseURL, buffer)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	default:
		return nil, errors.New("wrong method")
	}
}

// WithBaseURL - option to add a new address for the request
func WithBaseURL(baseURL string) Options {
	return func(client *Client) {
		client.baseURL = baseURL
	}
}

// WithTimeoutRequest - option to add new request timeout
func WithTimeoutRequest(timeoutReq time.Duration) Options {
	return func(client *Client) {
		client.httpClient.Timeout = timeoutReq
	}
}

// WithMaxRetry - option to add a new number of attempts for the request
func WithMaxRetry(retry uint) Options {
	return func(client *Client) {
		client.retry = retry
	}
}

// WithTimeoutBetweenReq - option to add a new timeout between retries
func WithTimeoutBetweenReq(timeoutBetweenReq time.Duration) Options {
	return func(client *Client) {
		client.timeoutBetweenReq = timeoutBetweenReq
	}
}
