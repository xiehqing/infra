package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/xiehqing/infra/pkg/logs"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	client  *http.Client
	BaseURL string
}

func (c *Client) Client() *http.Client {
	return c.client
}

// NewClient creates a new http client with the given base url and timeout.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		client: &http.Client{
			Timeout: timeout,
		},
		BaseURL: baseURL,
	}
}

// NewTransportClient  creates a new http client with the given base url, transport and timeout.
func NewTransportClient(baseURL string, transport http.RoundTripper, timeout time.Duration) *Client {
	return &Client{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		BaseURL: baseURL,
	}
}

// NewDefaultClient creates a new http client with the given base url and default timeout.
func NewDefaultClient(baseURL string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 100 * time.Second,
		},
		BaseURL: baseURL,
	}
}

// buildRequest builds a new http request with the given options.
func (c *Client) buildRequest(options *RequestOption) (*http.Request, error) {
	var body io.Reader
	if options.Body != nil {
		// 判断options.Body是否是[]byte
		if _, ok := options.Body.([]byte); ok {
			body = bytes.NewBuffer(options.Body.([]byte))
		} else {
			jsonData, err := json.Marshal(options.Body)
			if err != nil {
				return nil, fmt.Errorf("httpx.buildRequest: marshal body error: %v", err)
			}
			body = bytes.NewBuffer(jsonData)
		}
	}
	// 处理查询参数
	reqURL := c.BaseURL + options.Path
	if len(options.Query) > 0 {
		params := url.Values{}
		for key, value := range options.Query {
			params.Add(key, value)
		}
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}
	req, err := http.NewRequest(options.Method.String(), reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("httpx.buildRequest: create request error: %v", err)
	}
	// 设置请求头
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}
	if options.Body != nil && strings.TrimSpace(req.Header.Get("Content-Type")) == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// Do send a new http request with the given options.
func (c *Client) Do(options *RequestOption) (*http.Response, error) {
	requestTime := time.Now()
	request, err := c.buildRequest(options)
	if options.PrintLog {
		r := &RequestLog{
			Timestamp: requestTime.Format("2006-01-02 15:04:05.000"),
			Method:    options.Method.String(),
			URL:       request.URL.String(),
			Headers:   options.Headers,
			Body:      options.Body,
			RequestID: options.RequestID,
		}
		// 判断options.Body是否是[]byte
		if _, ok := options.Body.([]byte); ok {
			r.Body = string(options.Body.([]byte))
		} else {
			jsonData, err := json.Marshal(options.Body)
			if err != nil {
				logs.Errorf("httpx.Do: marshal request body error: %v", err)
			} else {
				r.Body = string(jsonData)
			}
		}
		LogRequestJSON(r, options.Sensitive)
	}
	if err != nil {
		return nil, err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, errors.WithMessage(err, "httpx.Do: do request error")
	}
	defer response.Body.Close()

	// 读取响应体内容到缓冲区
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "httpx.Do: read response body error")
	}
	// 创建可以重复读取的响应
	bufferedResp := &BufferedResponse{
		Response: &http.Response{
			Status:           response.Status,
			StatusCode:       response.StatusCode,
			Proto:            response.Proto,
			ProtoMajor:       response.ProtoMajor,
			ProtoMinor:       response.ProtoMinor,
			Header:           response.Header,
			ContentLength:    int64(len(bodyBytes)),
			TransferEncoding: response.TransferEncoding,
			Close:            response.Close,
			Uncompressed:     response.Uncompressed,
			Trailer:          response.Trailer,
			Request:          response.Request,
			TLS:              response.TLS,
		},
		bodyBuffer: bodyBytes,
	}

	// 设置可以重复读取的Body
	bufferedResp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	responseTime := time.Now()
	if options.PrintLog {
		r := &ResponseLog{
			Timestamp:  responseTime.Format("2006-01-02 15:04:05.000"),
			StatusCode: response.StatusCode,
			RequestID:  options.RequestID,
			DurationMs: responseTime.Sub(requestTime).Milliseconds(),
			Body:       string(bodyBytes),
		}
		if err != nil {
			r.Error = err.Error()
		}
		LogResponseJSON(r)
	}
	return bufferedResp.Response, nil
}

// DoRaw sends a new http request with the given options and returns the original response body for streaming scenarios.
func (c *Client) DoRaw(options *RequestOption) (*http.Response, error) {
	requestTime := time.Now()
	request, err := c.buildRequest(options)
	if options.PrintLog {
		r := &RequestLog{
			Timestamp: requestTime.Format("2006-01-02 15:04:05.000"),
			Method:    options.Method.String(),
			URL:       request.URL.String(),
			Headers:   options.Headers,
			Body:      options.Body,
			RequestID: options.RequestID,
		}
		if _, ok := options.Body.([]byte); ok {
			r.Body = string(options.Body.([]byte))
		} else {
			jsonData, marshalErr := json.Marshal(options.Body)
			if marshalErr != nil {
				logs.Errorf("httpx.DoRaw: marshal request body error: %v", marshalErr)
			} else {
				r.Body = string(jsonData)
			}
		}
		LogRequestJSON(r, options.Sensitive)
	}
	if err != nil {
		return nil, err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, errors.WithMessage(err, "httpx.DoRaw: do request error")
	}
	if options.PrintLog {
		responseTime := time.Now()
		r := &ResponseLog{
			Timestamp:  responseTime.Format("2006-01-02 15:04:05.000"),
			StatusCode: response.StatusCode,
			RequestID:  options.RequestID,
			DurationMs: responseTime.Sub(requestTime).Milliseconds(),
		}
		LogResponseJSON(r)
	}
	return response, nil
}

// DoWithPtr sends a new http request with the given options and unmarshal the response body to the given interface.
func (c *Client) DoWithPtr(options *RequestOption, resp interface{}) error {
	response, err := c.Do(options)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	// 读取响应体内容到缓冲区
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.WithMessage(err, "httpx.DoWithPtr: read response body error")
	}
	return json.Unmarshal(bodyBytes, resp)
}
