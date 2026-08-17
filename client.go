package odoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultTimeout    = 60 * time.Second
	DefaultMaxRetries = 2
	DefaultPageSize   = 200
	DefaultUserAgent  = "tomassicoffee-odoo-sync/go"
)

type ClientConfig struct {
	BaseURL    string
	Database   string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
	UserAgent  string
	HTTPClient *http.Client
}

type Client struct {
	endpointBaseURL string
	headers         map[string]string
	httpClient      *http.Client
	timeout         time.Duration
	maxRetries      int
}

func NewClientFromEnv() (*Client, error) {
	baseURL := firstEnv("ODOO_URL")
	apiKey := firstEnv("ODOO_API_KEY", "ODOO_PASSWORD", "ODOO_TOKEN")
	database := firstEnv("ODOO_DATABASE", "ODOO_DB")

	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("missing Odoo configuration: set ODOO_URL and ODOO_API_KEY")
	}

	timeoutMs, err := positiveIntEnv("ODOO_TIMEOUT_MS", int(DefaultTimeout/time.Millisecond))
	if err != nil {
		return nil, err
	}
	maxRetries, err := positiveIntEnv("ODOO_MAX_RETRIES", DefaultMaxRetries)
	if err != nil {
		return nil, err
	}

	return NewClient(ClientConfig{
		BaseURL:    baseURL,
		Database:   database,
		APIKey:     apiKey,
		Timeout:    time.Duration(timeoutMs) * time.Millisecond,
		MaxRetries: maxRetries,
		UserAgent:  envOr("ODOO_USER_AGENT", DefaultUserAgent),
	})
}

func NewClient(config ClientConfig) (*Client, error) {
	base, err := sanitizeBaseURL(config.BaseURL)
	if err != nil {
		return nil, err
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("missing Odoo API key")
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	maxRetries := config.MaxRetries
	if maxRetries < 0 {
		maxRetries = DefaultMaxRetries
	}
	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	headers := map[string]string{
		"Authorization": "bearer " + config.APIKey,
		"Content-Type":  "application/json; charset=utf-8",
		"User-Agent":    userAgent,
	}
	if config.Database != "" {
		headers["X-Odoo-Database"] = config.Database
	}

	return &Client{
		endpointBaseURL: base + "/json/2",
		headers:         headers,
		httpClient:      httpClient,
		timeout:         timeout,
		maxRetries:      maxRetries,
	}, nil
}

func newOdooHTTPError(model string, method Method, statusCode int, status string, body []byte) error {
	text := strings.TrimSpace(string(body))

	var payload OdooErrorPayload
	err := json.Unmarshal(body, &payload)
	if err == nil && (payload.Name != "" || payload.Message != "") {
		return &OdooHTTPError{
			Model:      model,
			Method:     method,
			StatusCode: statusCode,
			Status:     status,
			Payload:    &payload,
			RawBody:    text,
		}
	}

	if len(text) > 2000 {
		text = text[:2000] + "..."
	}

	return &OdooHTTPError{
		Model:      model,
		Method:     method,
		StatusCode: statusCode,
		Status:     status,
		RawBody:    text,
	}
}

func (c *Client) Call(ctx context.Context, model string, method Method, payload Payload, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode Odoo JSON-2 payload for %s.%s: %w", model, method, err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", c.endpointBaseURL, model, method)

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()
		req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			cancel()
			return fmt.Errorf("create Odoo JSON-2 request for %s.%s: %w", model, method, err)
		}
		for key, value := range c.headers {
			req.Header.Set(key, value)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries {
				sleepBackoff(attempt)
				continue
			}
			return fmt.Errorf("Odoo JSON-2 request failed for %s.%s: %w", model, method, err)
		}

		responseBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			if attempt < c.maxRetries {
				sleepBackoff(attempt)
				continue
			}
			return fmt.Errorf("read Odoo JSON-2 response for %s.%s: %w", model, method, readErr)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if attempt < c.maxRetries && retryableStatus(resp.StatusCode) {
				sleepBackoff(attempt)
				continue
			}
			fmtError := newOdooHTTPError(model, method, resp.StatusCode, resp.Status, []byte(responseBody))
			return fmt.Errorf("Odoo JSON-2 request failed for %w", fmtError)
		}

		if out == nil {
			return nil
		}
		if len(strings.TrimSpace(string(responseBody))) == 0 {
			return json.Unmarshal([]byte("null"), out)
		}

		var maybeError map[string]any
		if err := json.Unmarshal(responseBody, &maybeError); err == nil {
			if _, ok := maybeError["error"]; ok && len(maybeError) <= 2 {
				return fmt.Errorf("Odoo JSON-2 returned error for %s.%s: %s", model, method, string(responseBody))
			}
		}

		if err := json.Unmarshal(responseBody, out); err != nil {
			return fmt.Errorf("decode Odoo JSON-2 response for %s.%s: %w", model, method, err)
		}
		return nil
	}

	return fmt.Errorf("Odoo JSON-2 request failed for %s.%s", model, method)
}

func firstEnv(names ...string) string {
	for _, name := range names {
		value := strings.TrimSpace(os.Getenv(name))
		if value != "" {
			return value
		}
	}
	return ""
}

func envOr(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func positiveIntEnv(name string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}

func sanitizeBaseURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("ODOO_URL must be a valid http(s) URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("ODOO_URL must use http or https")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func retryableStatus(status int) bool {
	switch status {
	case 408, 425, 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

func sleepBackoff(attempt int) {
	time.Sleep(time.Duration(250*(1<<attempt)) * time.Millisecond)
}

func trimBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return "<empty response>"
	}
	if len(text) > 2000 {
		return text[:2000]
	}
	return text
}
