package agent

type MessageFormatType string

const (
	MessageFormatBase     MessageFormatType = "base"
	MessageFormatStandard MessageFormatType = "standard"
	MessageFormatOrigin   MessageFormatType = "origin"
)

type AppExecuteRequest struct {
	AppID          int64    `json:"appId"`
	UserID         int64    `json:"userId,omitempty"`
	TenantID       int64    `json:"tenantId,omitempty"`
	DataType       string   `json:"dataType,omitempty"`
	WorkingDir     string   `json:"workingDir"`
	ReferenceFiles []string `json:"referenceFiles,omitempty"`
	Prompt         string   `json:"prompt,omitempty"`
}

type AppStreamRequest struct {
	AppExecuteRequest
	Stream      bool              `json:"stream,omitempty"`
	MessageType MessageFormatType `json:"messageType,omitempty"`
}

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type AgentResult struct {
	Reasoning string `json:"reasoning"`
	Content   string `json:"content"`
	Usage     Usage  `json:"usage"`
}

type Usage struct {
	InputTokens         int64 `json:"input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	TotalTokens         int64 `json:"total_tokens"`
	ReasoningTokens     int64 `json:"reasoning_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_tokens"`
	CacheReadTokens     int64 `json:"cache_read_tokens"`
}

type SseMessage struct {
	ID    string `json:"id,omitempty"`
	Event string `json:"event,omitempty"`
	Data  string `json:"data,omitempty"`
}
