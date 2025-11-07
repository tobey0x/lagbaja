package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/tobey0x/lagbaja/internal/models"
	apperrors "github.com/tobey0x/lagbaja/pkg/errors"
	"google.golang.org/api/option"
)

type FlashcardService struct {
	pdfService *PDFService
	client     *genai.Client
	model      *genai.GenerativeModel
}

func NewFlashcardService(pdfService *PDFService, apiKey string) *FlashcardService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error creating Gemini client: %v", err)
		return nil
	}

	model := client.GenerativeModel("gemini-2.0-flash-lite")
	
	return &FlashcardService{
		pdfService: pdfService,
		client:     client,
		model:      model,
	}
}


func (s *FlashcardService) GenerateFromURL(url string) (*models.FlashcardSet, error) {
	// Download PDF
	pdfData, err := s.pdfService.DownloadPDF(url)
	if err != nil {
		return nil, err
	}

	// Validate PDF
	if err := s.pdfService.ValidatePDF(pdfData); err != nil {
		return nil, err
	}

	// Extract text
	text, err := s.pdfService.ExtractText(pdfData)
	if err != nil {
		return nil, err
	}

	// Generate flashcards from text
	return s.generateFlashcards(text, url)
}

func (s *FlashcardService) GenerateFromPDFData(pdfData []byte) (*models.FlashcardSet, error) {
	// Validate PDF
	if err := s.pdfService.ValidatePDF(pdfData); err != nil {
		return nil, err
	}

	// Extract text
	text, err := s.pdfService.ExtractText(pdfData)
	if err != nil {
		return nil, err
	}

	// Generate flashcards from text
	return s.generateFlashcards(text, "uploaded_pdf")
}

func (s *FlashcardService) GenerateFromText(text string) (*models.FlashcardSet, error) {
	return s.generateFlashcards(text, "user_input")
}

func (s *FlashcardService) generateFlashcards(text, source string) (*models.FlashcardSet, error) {
	log.Printf("Generating flashcards from text (length: %d)", len(text))

	if len(strings.TrimSpace(text)) == 0 {
		return nil, apperrors.NewAppError(
			models.InvalidParams,
			"no content to generate flashcards from",
			nil,
		)
	}

	prompt := fmt.Sprintf(`Create a set of 5-10 high-quality flashcards from the following text. 
Each flashcard should:
- Have a clear, specific question
- Include a concise but comprehensive answer
- Be categorized with an appropriate topic
- Cover key concepts, definitions, and applications

Text to process:
%s

Format each flashcard as:
Q: [Question]
A: [Answer]
T: [Topic]

`, text)

	ctx := context.Background()
	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, apperrors.NewAppError(
			models.InternalError,
			"failed to generate flashcards",
			err,
		)
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, apperrors.NewAppError(
			models.InternalError,
			"no response generated from AI model",
			nil,
		)
	}

	// Parse Gemini's response into flashcards
	var flashcards []models.Flashcard
	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	cards := strings.Split(responseText, "\n\n")

	for _, card := range cards {
		if !strings.Contains(card, "Q:") || !strings.Contains(card, "A:") {
			continue
		}

		lines := strings.Split(card, "\n")
		var question, answer, topic string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Q:") {
				question = strings.TrimSpace(strings.TrimPrefix(line, "Q:"))
			} else if strings.HasPrefix(line, "A:") {
				answer = strings.TrimSpace(strings.TrimPrefix(line, "A:"))
			} else if strings.HasPrefix(line, "T:") {
				topic = strings.TrimSpace(strings.TrimPrefix(line, "T:"))
			}
		}

		if question != "" && answer != "" {
			if topic == "" {
				topic = "Concept"
			}
			flashcards = append(flashcards, models.Flashcard{
				Question: question,
				Answer:   answer,
				Topic:    topic,
			})
		}
	}

	// Ensure we have at least one flashcard
	if len(flashcards) == 0 {
		return nil, apperrors.NewAppError(
			models.InternalError,
			"could not generate meaningful flashcards from the content",
			nil,
		)
	}

	return &models.FlashcardSet{
		Title:      s.generateTitle(source),
		Source:     source,
		Flashcards: flashcards,
		TotalCards: len(flashcards),
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *FlashcardService) generateTitle(source string) string {
	if strings.HasPrefix(source, "http") {
		return "Flashcards from PDF"
	}
	return "Study Flashcards"
}

func (s *FlashcardService) truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func (s *FlashcardService) FormatAsText(set *models.FlashcardSet) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s\n\n", set.Title))
	builder.WriteString(fmt.Sprintf("Generated %d flashcards from: %s\n\n", set.TotalCards, set.Source))

	for i, card := range set.Flashcards {
		builder.WriteString(fmt.Sprintf("**Card %d**", i+1))
		if card.Topic != "" {
			builder.WriteString(fmt.Sprintf(" (%s)", card.Topic))
		}
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("Q: %s\n", card.Question))
		builder.WriteString(fmt.Sprintf("A: %s\n\n", card.Answer))
	}

	return builder.String()
}

func (s *FlashcardService) ExtractPDFURL(text string) string {
	words := strings.Fields(text)
	for _, word := range words {
		cleaned := strings.TrimRight(word, ".,;")
		if strings.HasPrefix(cleaned, "http") && strings.Contains(strings.ToLower(cleaned), ".pdf") {
			return cleaned
		}
	}
	return ""
}