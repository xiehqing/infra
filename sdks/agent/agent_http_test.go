package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Execute_SendsJSONBodyAndHeaders(t *testing.T) {
	var gotMethod string
	var gotContentType string
	var gotHeader string
	var gotReq AppExecuteRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotHeader = r.Header.Get("X-Test-Header")
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":"done"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, WithHeader("X-Test-Header", "present"))
	resp, err := client.Execute(&AppExecuteRequest{
		AppID:      2,
		Prompt:     "你好",
		DataType:   "ad",
		WorkingDir: `C:\Users\17461\GolandProjects\app`,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if resp != "done" {
		t.Fatalf("unexpected response: %q", resp)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotContentType != "application/json" {
		t.Fatalf("unexpected content type: %s", gotContentType)
	}
	if gotHeader != "present" {
		t.Fatalf("expected custom header to be forwarded, got %q", gotHeader)
	}
	if gotReq.AppID != 2 || gotReq.DataType != "ad" || gotReq.WorkingDir != `C:\Users\17461\GolandProjects\app` || gotReq.Prompt != "你好" {
		t.Fatalf("unexpected request payload: %+v", gotReq)
	}
}
