package markitdown

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/xiehqing/infra/pkg/httpx"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	client  *httpx.Client
	headers map[string]string
	timeout time.Duration
}

type Option func(*Client)

func WithHTTPClient(client *httpx.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.client = client
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.timeout = timeout
		}
	}
}

func WithTimeoutSecond(timeoutSecond int) Option {
	return func(c *Client) {
		if timeoutSecond > 0 {
			c.timeout = time.Duration(timeoutSecond) * time.Second
		}
	}
}

func WithHeader(key, value string) Option {
	return func(c *Client) {
		if key == "" {
			return
		}
		c.headers[key] = value
	}
}

func NewClient(baseURL string, opts ...Option) *Client {
	baseUrl := strings.TrimRight(baseURL, "/")
	client := &Client{
		baseURL: baseUrl,
		headers: map[string]string{},
	}
	for _, opt := range opts {
		opt(client)
	}
	client.client = httpx.NewClient(baseUrl, client.timeout)
	return client
}

// Convert 转换
func (c *Client) Convert(path string) (string, error) {
	if path == "" {
		return "", errors.New("sdk.markitdown: filepath is empty")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", errors.WithMessage(err, "sdk.markitdown: open file failed")
	}
	defer file.Close()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	defer writer.Close()
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return "", errors.WithMessage(err, "sdk.markitdown: create form file failed")
	}
	if _, err = io.Copy(part, file); err != nil {
		return "", errors.WithMessage(err, "sdk.markitdown: copy file failed")
	}
	headers := make(map[string]string, len(c.headers))
	for key, value := range c.headers {
		headers[key] = value
	}
	headers["Content-Type"] = writer.FormDataContentType()
	headers["Accept"] = "application/json, text/plain;q=0.9, */*;q=0.8"

	opts := httpx.NewOptions(
		httpx.WithPath("/convert"),
		httpx.WithHeaders(headers),
		httpx.WithBody(body.Bytes()),
		httpx.WithMethodPost(),
		httpx.WithPrintLog(true),
	)
	response, err := c.client.Do(opts)
	if err != nil {
		return "", errors.WithMessage(err, "sdk.markitdown: do request failed")
	}
	defer response.Body.Close()
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.WithMessage(err, "sdk.markitdown: read response body failed")
	}
	if response.StatusCode != http.StatusOK {
		return "", errors.Errorf("sdk.markitdown: response status code is %d, body: %s", response.StatusCode, strings.TrimSpace(string(respBody)))
	}
	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	if strings.Contains(contentType, "application/json") {
		result, ce := parseConvertJSON(respBody)
		if ce != nil {
			return "", errors.WithMessage(ce, "sdk.markitdown: parse json response failed")
		}
		return result, nil
	}
	return strings.TrimSpace(string(respBody)), nil
}

// 解析MarkItDown JSON响应
func parseConvertJSON(data []byte) (string, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", errors.WithMessage(err, "sdk.markitdown: parse markitdown response failed")
	}
	for _, key := range []string{"text_content", "textContent", "markdown", "content", "text", "result"} {
		value, ok := resp[key]
		if !ok {
			continue
		}
		if text, ok := value.(string); ok {
			return strings.TrimSpace(text), nil
		}
	}
	return "", errors.Errorf("sdk.markitdown: response content is empty: %s", strings.TrimSpace(string(data)))
}
