package agent

import (
	"github.com/xiehqing/infra/pkg/jsonx"
	"testing"
)

func TestClient_Execute(t *testing.T) {
	client := NewClient("http://127.0.0.1:7799")
	execute, err := client.Execute(&AppExecuteRequest{
		AppID:      1,
		Prompt:     "你好",
		DataType:   "ad",
		WorkingDir: "C:\\Users\\17461\\GolandProjects\\app",
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(execute)
}

func TestClient_Stream(t *testing.T) {
	client := NewClient("http://127.0.0.1:7799")
	sr := &AppStreamRequest{}
	sr.AppID = 1
	sr.Prompt = "全程使用中文直接帮我生成一篇800字作文"
	sr.DataType = "ad"
	sr.WorkingDir = "C:\\Users\\17461\\GolandProjects\\app"
	sr.MessageType = MessageFormatStandard
	err := client.Stream(sr, func(chunk SseMessage) error {
		t.Log(chunk)
		return nil
	})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestClient_AppSession(t *testing.T) {
	client := NewClient("http://127.0.0.1:7799", WithPrintLog(false))
	messages, err := client.AppSessions(1, 0, "", "")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(jsonx.ToJsonIgnoreError(messages))
}

func TestClient_StandardMessages(t *testing.T) {
	client := NewClient("http://127.0.0.1:7799", WithPrintLog(false))
	messages, err := client.StandardMessages("9ec96ef4-3e81-482f-9742-5e2cf79faeaa")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(jsonx.ToJsonIgnoreError(messages))
}

func TestClient_Messages(t *testing.T) {
	client := NewClient("http://127.0.0.1:7799", WithPrintLog(false))
	messages, err := client.Messages("9ec96ef4-3e81-482f-9742-5e2cf79faeaa")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(jsonx.ToJsonIgnoreError(messages))
}
