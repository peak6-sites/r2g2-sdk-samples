package main

// Assistant represents an AI assistant
type Assistant struct {
	ID      string `json:"id"`
	StoreID string `json:"storeId"`
}

// GetAssistantRequest represents a request to get an assistant
type GetAssistantRequest struct {
	ID string `json:"id"`
}

// UploadFileRequest represents a request to upload a file
type UploadFileRequest struct {
	StoreID  string `json:"storeId"`
	Filename string `json:"filename"`
}

// UploadFileResponse represents a response from the upload file endpoint
type UploadFileResponse struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// CreateThreadAndRunRequest represents a request to create a thread and run
type CreateThreadAndRunRequest struct {
	AssistantID string        `json:"assistantId"`
	Thread      ThreadRequest `json:"thread"`
}

// ThreadRequest represents a request to create a thread
type ThreadRequest struct {
	Messages []MessageRequest `json:"messages"`
}

// MessageRequest represents a request to create a message
type MessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Run represents an assistant run
type Run struct {
	ID            string `json:"id"`
	ThreadID      string `json:"threadId"`
	State         string `json:"state"`
	FailureReason string `json:"failureReason,omitempty"`
}

// GetRunRequest represents a request to get a run
type GetRunRequest struct {
	ID string `json:"id"`
}

// ListMessagesRequest represents a request to list messages
type ListMessagesRequest struct {
	ThreadID string `json:"threadId"`
	RunID    string `json:"runId"`
}

// ListMessagesResponse represents a response from the list messages endpoint
type ListMessagesResponse struct {
	Messages []Message `json:"messages"`
}

// Message represents a message in a thread
type Message struct {
	Content []MessageContent `json:"content"`
}

// MessageContent represents the content of a message
type MessageContent struct {
	Text string `json:"text,omitempty"`
}
