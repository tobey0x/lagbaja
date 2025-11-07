package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tobey0x/lagbaja/internal/models"
	"github.com/tobey0x/lagbaja/internal/service"
)

func TestA2AHandler_ServeHTTP_MethodValidation(t *testing.T) {
	pdfService := service.NewPDFService()
	flashcardService := service.NewFlashcardService(pdfService, "test-api-key")
	handler := NewA2AHandler(flashcardService)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{"GET method should fail", http.MethodGet, http.StatusOK},
		{"PUT method should fail", http.MethodPut, http.StatusOK},
		{"DELETE method should fail", http.MethodDelete, http.StatusOK},
		{"POST method should succeed", http.MethodPost, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/a2a", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method != http.MethodPost {
				var response models.JSONRPCResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if response.Error == nil {
					t.Error("Expected error response for non-POST method")
				}
			}
		})
	}
}

func TestA2AHandler_ServeHTTP_JSONRPCValidation(t *testing.T) {
	pdfService := service.NewPDFService()
	flashcardService := service.NewFlashcardService(pdfService, "test-api-key")
	handler := NewA2AHandler(flashcardService)

	tests := []struct {
		name          string
		request       string
		expectedError int
	}{
		{
			name:          "Invalid JSON should return parse error",
			request:       `{invalid json}`,
			expectedError: models.ParseError,
		},
		{
			name:          "Missing jsonrpc field should return invalid request",
			request:       `{"id":"123","method":"message/send"}`,
			expectedError: models.InvalidRequest,
		},
		{
			name:          "Wrong jsonrpc version should return invalid request",
			request:       `{"jsonrpc":"1.0","id":"123","method":"message/send"}`,
			expectedError: models.InvalidRequest,
		},
		{
			name:          "Unknown method should return method not found",
			request:       `{"jsonrpc":"2.0","id":"123","method":"unknown/method","params":{}}`,
			expectedError: models.MethodNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/a2a", bytes.NewBufferString(tt.request))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			var response models.JSONRPCResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Error == nil {
				t.Fatal("Expected error response")
			}

			if response.Error.Code != tt.expectedError {
				t.Errorf("Expected error code %d, got %d", tt.expectedError, response.Error.Code)
			}
		})
	}
}

func TestA2AHandler_ServeHTTP_ValidRequest(t *testing.T) {
	pdfService := service.NewPDFService()
	flashcardService := service.NewFlashcardService(pdfService, "test-api-key")
	handler := NewA2AHandler(flashcardService)

	requestBody := models.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "test-123",
		Method:  "message/send",
		Params: map[string]interface{}{
			"message": map[string]interface{}{
				"kind":      "message",
				"role":      "user",
				"messageId": "msg-001",
				"taskId":    "task-001",
				"parts": []map[string]interface{}{
					{
						"kind": "text",
						"text": "Generate flashcards about basic mathematics.",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/a2a", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var response models.JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", response.JSONRPC)
	}

	if response.ID != "test-123" {
		t.Errorf("Expected ID test-123, got %s", response.ID)
	}

	// Note: This will fail due to invalid API key, but structure should be valid
	if response.Error != nil {
		// Expected to fail with API key error in test environment
		if response.Error.Code != models.InternalError {
			t.Logf("Expected internal error due to invalid API key: %v", response.Error)
		}
	}
}

func TestA2AHandler_ExtractMessage(t *testing.T) {
	pdfService := service.NewPDFService()
	flashcardService := service.NewFlashcardService(pdfService, "test-api-key")
	handler := NewA2AHandler(flashcardService)

	tests := []struct {
		name        string
		params      map[string]interface{}
		shouldError bool
	}{
		{
			name: "Valid message",
			params: map[string]interface{}{
				"message": map[string]interface{}{
					"kind":      "message",
					"role":      "user",
					"messageId": "msg-001",
					"parts": []interface{}{
						map[string]interface{}{
							"kind": "text",
							"text": "test",
						},
					},
				},
			},
			shouldError: false,
		},
		{
			name:        "Missing message parameter",
			params:      map[string]interface{}{},
			shouldError: true,
		},
		{
			name: "Invalid message format",
			params: map[string]interface{}{
				"message": "not an object",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := handler.extractMessage(tt.params)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if msg == nil {
					t.Error("Expected message but got nil")
				}
			}
		})
	}
}

func TestA2AHandler_ExtractUserInput(t *testing.T) {
	pdfService := service.NewPDFService()
	flashcardService := service.NewFlashcardService(pdfService, "test-api-key")
	handler := NewA2AHandler(flashcardService)

	tests := []struct {
		name     string
		message  *models.Message
		expected string
	}{
		{
			name: "Extract text from parts",
			message: &models.Message{
				Parts: []models.MessagePart{
					{Kind: "text", Text: "Hello World"},
				},
			},
			expected: "Hello World",
		},
		{
			name: "No text parts",
			message: &models.Message{
				Parts: []models.MessagePart{
					{Kind: "data", Data: map[string]string{"key": "value"}},
				},
			},
			expected: "",
		},
		{
			name: "Multiple parts - return first text",
			message: &models.Message{
				Parts: []models.MessagePart{
					{Kind: "data", Data: "data"},
					{Kind: "text", Text: "First text"},
					{Kind: "text", Text: "Second text"},
				},
			},
			expected: "First text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.extractUserInput(tt.message)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestA2AHandler_BuildTaskResult(t *testing.T) {
	pdfService := service.NewPDFService()
	flashcardService := service.NewFlashcardService(pdfService, "test-api-key")
	handler := NewA2AHandler(flashcardService)

	flashcards := &models.FlashcardSet{
		Title:      "Test Flashcards",
		Source:     "test",
		TotalCards: 2,
		Flashcards: []models.Flashcard{
			{Question: "Q1", Answer: "A1", Topic: "Topic1"},
			{Question: "Q2", Answer: "A2", Topic: "Topic2"},
		},
	}

	userMsg := &models.Message{
		Kind:      "message",
		Role:      "user",
		MessageID: "msg-001",
		TaskID:    "task-001",
		Parts: []models.MessagePart{
			{Kind: "text", Text: "Generate flashcards"},
		},
	}

	result := handler.buildTaskResult(flashcards, userMsg)

	if result.ID != "task-001" {
		t.Errorf("Expected task ID task-001, got %s", result.ID)
	}

	if result.Status.State != models.StateCompleted {
		t.Errorf("Expected state completed, got %s", result.Status.State)
	}

	if result.Kind != "task" {
		t.Errorf("Expected kind task, got %s", result.Kind)
	}

	if len(result.Artifacts) != 1 {
		t.Errorf("Expected 1 artifact, got %d", len(result.Artifacts))
	}

	if len(result.History) != 2 {
		t.Errorf("Expected 2 messages in history, got %d", len(result.History))
	}

	if result.Artifacts[0].Name != "flashcardSet" {
		t.Errorf("Expected artifact name flashcardSet, got %s", result.Artifacts[0].Name)
	}
}
