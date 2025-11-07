package service

import (
	"strings"
	"testing"

	"github.com/tobey0x/lagbaja/internal/models"
)

func TestFlashcardService_ExtractPDFURL(t *testing.T) {
	pdfService := NewPDFService()
	service := NewFlashcardService(pdfService, "test-api-key")

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Extract PDF URL from text",
			text:     "Please analyze this document: https://example.com/document.pdf and generate flashcards.",
			expected: "https://example.com/document.pdf",
		},
		{
			name:     "Extract PDF URL with trailing comma",
			text:     "Check out https://example.com/file.pdf, it's interesting.",
			expected: "https://example.com/file.pdf",
		},
		{
			name:     "No PDF URL in text",
			text:     "This is just plain text without any URLs.",
			expected: "",
		},
		{
			name:     "URL without .pdf extension",
			text:     "Visit https://example.com for more info.",
			expected: "",
		},
		{
			name:     "Multiple URLs, return first PDF",
			text:     "See https://example.com/first.pdf and https://example.com/second.pdf",
			expected: "https://example.com/first.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ExtractPDFURL(tt.text)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFlashcardService_GenerateTitle(t *testing.T) {
	pdfService := NewPDFService()
	service := NewFlashcardService(pdfService, "test-api-key")

	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "HTTP URL source",
			source:   "https://example.com/document.pdf",
			expected: "Flashcards from PDF",
		},
		{
			name:     "HTTPS URL source",
			source:   "https://example.com/file.pdf",
			expected: "Flashcards from PDF",
		},
		{
			name:     "User input source",
			source:   "user_input",
			expected: "Study Flashcards",
		},
		{
			name:     "Plain text source",
			source:   "manual entry",
			expected: "Study Flashcards",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.generateTitle(tt.source)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFlashcardService_Truncate(t *testing.T) {
	pdfService := NewPDFService()
	service := NewFlashcardService(pdfService, "test-api-key")

	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{
			name:     "Text shorter than max length",
			text:     "Short text",
			maxLen:   50,
			expected: "Short text",
		},
		{
			name:     "Text equal to max length",
			text:     "Exactly fifty characters in this text right here!",
			maxLen:   50,
			expected: "Exactly fifty characters in this text right here!",
		},
		{
			name:     "Text longer than max length",
			text:     "This is a very long text that needs to be truncated because it exceeds the maximum length.",
			maxLen:   20,
			expected: "This is a very long ...",
		},
		{
			name:     "Empty text",
			text:     "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.truncate(tt.text, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFlashcardService_FormatAsText(t *testing.T) {
	pdfService := NewPDFService()
	service := NewFlashcardService(pdfService, "test-api-key")

	flashcardSet := &models.FlashcardSet{
		Title:      "Test Flashcards",
		Source:     "test_source",
		TotalCards: 2,
		Flashcards: []models.Flashcard{
			{
				Question: "What is 2+2?",
				Answer:   "4",
				Topic:    "Math",
			},
			{
				Question: "What is the capital of France?",
				Answer:   "Paris",
				Topic:    "Geography",
			},
		},
	}

	result := service.FormatAsText(flashcardSet)

	// Check that result contains expected elements
	if !strings.Contains(result, "Test Flashcards") {
		t.Error("Result should contain title")
	}

	if !strings.Contains(result, "Generated 2 flashcards") {
		t.Error("Result should contain total cards count")
	}

	if !strings.Contains(result, "What is 2+2?") {
		t.Error("Result should contain first question")
	}

	if !strings.Contains(result, "What is the capital of France?") {
		t.Error("Result should contain second question")
	}

	if !strings.Contains(result, "Math") {
		t.Error("Result should contain topics")
	}

	if !strings.Contains(result, "Card 1") && !strings.Contains(result, "Card 2") {
		t.Error("Result should contain card numbers")
	}
}

func TestPDFService_ValidatePDF(t *testing.T) {
	service := NewPDFService()

	tests := []struct {
		name        string
		data        []byte
		shouldError bool
	}{
		{
			name:        "Valid PDF header",
			data:        []byte("%PDF-1.4\nrest of pdf content"),
			shouldError: false,
		},
		{
			name:        "Invalid PDF - too small",
			data:        []byte("abc"),
			shouldError: true,
		},
		{
			name:        "Invalid PDF - wrong header",
			data:        []byte("This is not a PDF file"),
			shouldError: true,
		},
		{
			name:        "Empty data",
			data:        []byte{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePDF(tt.data)
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
