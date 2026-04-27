package agent

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"github.com/xiehqing/infra/pkg/httpx"
	"io"
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

func WithHeader(key, value string) Option {
	return func(c *Client) {
		if key == "" {
			return
		}
		c.headers[key] = value
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

// Execute agent执行
func (c *Client) Execute(req *AppExecuteRequest) (string, error) {
	var out Response[string]
	headers := make(map[string]string, len(c.headers))
	for key, value := range c.headers {
		headers[key] = value
	}
	opts := httpx.NewOptions(
		httpx.WithPath("/api/v1/app_execute"),
		httpx.WithHeaders(headers),
		httpx.WithBody(req),
		httpx.WithMethodPost(),
		httpx.WithPrintLog(true),
	)
	err := c.client.DoWithPtr(opts, &out)
	if err != nil {
		return "", errors.WithMessage(err, "agent.sdk: execute failed")
	}
	if out.Code != 0 {
		return "", errors.New(out.Message)
	}
	return out.Data, nil
}

// Stream agent执行
func (c *Client) Stream(req *AppStreamRequest, handler func(chunk SseMessage) error) error {
	req.Stream = true
	if req.MessageType == "" {
		req.MessageType = MessageFormatBase
	}
	headers := make(map[string]string, len(c.headers))
	for key, value := range c.headers {
		headers[key] = value
	}
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "text/event-stream"
	opts := httpx.NewOptions(
		httpx.WithPath("/api/v1/app_execute"),
		httpx.WithHeaders(headers),
		httpx.WithBody(req),
		httpx.WithMethodPost(),
		httpx.WithPrintLog(true),
	)
	response, err := c.client.DoRaw(opts)
	if err != nil {
		return errors.WithMessage(err, "agent.sdk: stream do failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		payload, _ := io.ReadAll(response.Body)
		return errors.New(strings.TrimSpace(string(payload)))
	}
	return consumeSSEMessage(response.Body, handler)
}

// consumeSSEMessage 消费sse消息
func consumeSSEMessage(reader io.Reader, handler func(chunk SseMessage) error) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var event SseMessage
	var dataLines []string

	flush := func() error {
		if len(dataLines) == 0 {
			event = SseMessage{}
			return nil
		}
		event.Data = strings.Join(dataLines, "\n")
		defer func() {
			event = SseMessage{}
			dataLines = nil
		}()
		return handler(event)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "event:"):
			event.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "id:"):
			event.ID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		case strings.HasPrefix(line, "data:"):
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("sdk.stream: read sse stream: %w", err)
	}
	return flush()
}
