package wire

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var defaultHTTPClient = &http.Client{Timeout: 10 * time.Second}

// Client holds API endpoint configuration.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client // optional, uses default if nil
}

// Do makes an API request and decodes the response into response.
func (c Client) Do(method, path string, body []byte, response any) error {
	url := c.resolveURL(path)

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = defaultHTTPClient
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= http.StatusBadRequest {
		apiErr, _ := decodeInto(data, nil)
		if apiErr != nil {
			return errors.New(apiErr.Message)
		}
		return fmt.Errorf("server returned %s", res.Status)
	}

	if response == nil || len(data) == 0 {
		return nil
	}

	apiErr, err := decodeInto(data, response)
	if apiErr != nil {
		return errors.New(apiErr.Message)
	}
	return err
}

// Get issues a GET request.
func (c Client) Get(path string, response any) error {
	return c.Do(http.MethodGet, path, nil, response)
}

// Post issues a POST request.
func (c Client) Post(path string, body []byte, response any) error {
	return c.Do(http.MethodPost, path, body, response)
}

// Put issues a PUT request.
func (c Client) Put(path string, body []byte, response any) error {
	return c.Do(http.MethodPut, path, body, response)
}

// Delete issues a DELETE request.
func (c Client) Delete(path string, response any) error {
	return c.Do(http.MethodDelete, path, nil, response)
}

func (c Client) resolveURL(path string) string {
	base := strings.TrimRight(c.BaseURL, "/")
	if path == "" {
		return base
	}
	cleanPath := strings.TrimLeft(path, "/")
	if base == "" {
		return "/" + cleanPath
	}
	return base + "/" + cleanPath
}
