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

type AppSession struct {
	AppID            int64   `json:"appId"`
	ID               string  `json:"id"`
	ParentSessionID  string  `json:"parentSessionId"`
	Title            string  `json:"title"`
	MessageCount     int64   `json:"messageCount"`
	PromptTokens     int64   `json:"promptTokens"`
	CompletionTokens int64   `json:"completionTokens"`
	SummaryMessageID string  `json:"summaryMessageId"`
	Cost             float64 `json:"cost"`
	Todos            []Todo  `json:"todos"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

type Todo struct {
	Content    string `json:"content"`
	Status     string `json:"status"`
	ActiveForm string `json:"active_form"`
}

type StandardMessage struct {
	ID        string         `json:"id"`
	Role      string         `json:"role"`
	Steps     []StandardStep `json:"steps"`
	Timestamp int64          `json:"timestamp"`
	Time      string         `json:"time"`
}

type StandardStep struct {
	Type StandardStepType `json:"type"`
	Data any              `json:"data"`
}

type StandardStepType string

const (
	Thought          StandardStepType = "thought"
	Chat             StandardStepType = "chat"
	ToolCallReq      StandardStepType = "tool_call_req"
	ToolCallResp     StandardStepType = "tool_call_resp"
	Error            StandardStepType = "error"
	FinishReasonType StandardStepType = "finish_reason"
	TokenUsage       StandardStepType = "token_usage"
)

type FinishReason string

const (
	FinishReasonEndTurn   FinishReason = "end_turn"
	FinishReasonMaxTokens FinishReason = "max_tokens"
	FinishReasonToolUse   FinishReason = "tool_use"
	FinishReasonCanceled  FinishReason = "canceled"
	FinishReasonError     FinishReason = "error"

	// Should never happen
	FinishReasonUnknown FinishReason = "unknown"
)

// DataMessage 数据消息
type DataMessage struct {
	ID               string            `json:"id"`
	Role             string            `json:"role"`
	SessionID        string            `json:"session_id"`
	Parts            []ContentPartData `json:"parts"`
	Model            string            `json:"model"`
	Provider         string            `json:"provider"`
	CreatedAt        string            `json:"createdAt"`
	UpdatedAt        string            `json:"updatedAt"`
	IsSummaryMessage bool              `json:"is_summary_message"`
	Files            []File            `json:"files,omitempty"`
}

type File struct {
	ID        string
	SessionID string
	Path      string
	Content   string
	Version   int64
	CreatedAt int64
	UpdatedAt int64
}

type ContentPartType string

const (
	ReasoningType  ContentPartType = "reasoning"
	TextType       ContentPartType = "text"
	ImageURLType   ContentPartType = "image_url"
	BinaryType     ContentPartType = "binary"
	ToolCallType   ContentPartType = "tool_call"
	ToolResultType ContentPartType = "tool_result"
	FinishType     ContentPartType = "finish"
)

type ContentPartData struct {
	Type ContentPartType `json:"type"`
	Data ContentPart     `json:"data"`
}

type ContentPart struct {
	// TextType
	Text string `json:"text,omitempty"`
	// ReasoningType
	Thinking         string                      `json:"thinking,omitempty"`
	Signature        string                      `json:"signature,omitempty"`
	ThoughtSignature string                      `json:"thought_signature,omitempty"` // Used for google
	ToolID           string                      `json:"tool_id,omitempty"`           // Used for openrouter google models
	ResponsesData    *ResponsesReasoningMetadata `json:"responses_data,omitempty"`
	StartedAt        int64                       `json:"started_at,omitempty,omitempty"`
	FinishedAt       int64                       `json:"finished_at,omitempty,omitempty"`
	// ToolCall ToolResult
	ID               string `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	Input            string `json:"input,omitempty"`
	ProviderExecuted bool   `json:"provider_executed,omitempty"`
	Finished         bool   `json:"finished,omitempty"`
	ToolCallID       string `json:"tool_call_id,omitempty"`
	Content          string `json:"content,omitempty"`
	Data             string `json:"data,omitempty"`
	MIMEType         string `json:"mime_type,omitempty"`
	Metadata         string `json:"metadata,omitempty"`
	IsError          bool   `json:"is_error,omitempty"`
	// finish
	Reason  FinishReason `json:"reason,omitempty"`
	Time    int64        `json:"time,omitempty"`
	Message string       `json:"message,omitempty"`
	Details string       `json:"details,omitempty"`
	// image url
	URL string `json:"url,omitempty"`
}

// ResponsesReasoningMetadata represents reasoning metadata for OpenAI Responses API.
type ResponsesReasoningMetadata struct {
	ItemID           string   `json:"item_id"`
	EncryptedContent *string  `json:"encrypted_content"`
	Summary          []string `json:"summary"`
}
