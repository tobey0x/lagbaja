package models


type Message struct {
	Kind      string        `json:"kind"`
	Role      string        `json:"role"`
	Parts     []MessagePart `json:"parts"`
	MessageID string        `json:"messageId"`
	TaskID    string        `json:"taskId,omitempty"`
	Metadata  interface{}   `json:"metadata,omitempty"`
}

type MessagePart struct {
	Kind string      `json:"kind"`
	Text string      `json:"text,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type TaskResult struct {
	ID        string     `json:"id"`
	ContextID string     `json:"contextId"`
	Status    Status     `json:"status"`
	Artifacts []Artifact `json:"artifacts"`
	History   []Message  `json:"history"`
	Kind      string     `json:"kind"`
}

type Status struct {
	State     string  `json:"state"`
	Timestamp string  `json:"timestamp"`
	Message   Message `json:"message"`
}

type Artifact struct {
	ArtifactID string        `json:"artifactId"`
	Name       string        `json:"name"`
	Parts      []MessagePart `json:"parts"`
}

type Configuration struct {
	Blocking bool `json:"blocking"`
}

// Task states
const (
	StateCompleted = "completed"
	StateRunning   = "running"
	StateFailed    = "failed"
)

// Message roles
const (
	RoleUser  = "user"
	RoleAgent = "agent"
)

// Part kinds
const (
	KindText    = "text"
	KindData    = "data"
	KindMessage = "message"
)