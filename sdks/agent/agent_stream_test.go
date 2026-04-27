package agent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Stream_ConsumesChunksBeforeEOF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}

		_, _ = fmt.Fprint(w, "event: message\ndata: first\n\n")
		flusher.Flush()
		time.Sleep(200 * time.Millisecond)
		_, _ = fmt.Fprint(w, "event: message\ndata: second\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := NewClient(server.URL)
	got := make(chan SseMessage, 2)
	done := make(chan error, 1)

	go func() {
		done <- client.Stream(&AppStreamRequest{
			AppExecuteRequest: AppExecuteRequest{
				AppID:      1,
				Prompt:     "test",
				DataType:   "ad",
				WorkingDir: `C:\work`,
			},
		}, func(chunk SseMessage) error {
			got <- chunk
			return nil
		})
	}()

	select {
	case chunk := <-got:
		if chunk.Data != "first" || chunk.Event != "message" {
			t.Fatalf("unexpected first chunk: %+v", chunk)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("first sse chunk was not delivered before stream finished")
	}

	if err := <-done; err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
}
