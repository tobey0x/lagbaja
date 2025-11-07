package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// "strings"
	"time"

	"github.com/tobey0x/lagbaja/internal/models"
	"github.com/tobey0x/lagbaja/internal/service"
	apperrors "github.com/tobey0x/lagbaja/pkg/errors"

	"github.com/google/uuid"
)

type A2AHandler struct {
	flashcardService *service.FlashcardService
}

func NewA2AHandler(flashcardService *service.FlashcardService) *A2AHandler {
	return &A2AHandler{
		flashcardService: flashcardService,
	}
}

func (h *A2AHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "unknown", models.InvalidRequest, "Only POST method is supported", "")
		return
	}

	var req models.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "", models.ParseError, "Parse error", err.Error())
		return
	}

	if req.JSONRPC != "2.0" {
		h.sendError(w, req.ID, models.InvalidRequest, "Invalid Request", "jsonrpc must be 2.0")
		return
	}

	switch req.Method {
	case "message/send":
		h.handleMessageSend(w, req)
	default:
		h.sendError(w, req.ID, models.MethodNotFound, "Method not found", fmt.Sprintf("Method %s not supported", req.Method))
	}
}

func (h *A2AHandler) handleMessageSend(w http.ResponseWriter, req models.JSONRPCRequest) {
	// Extract and validate message
	msg, err := h.extractMessage(req.Params)
	if err != nil {
		appErr := err.(*apperrors.AppError)
		h.sendError(w, req.ID, appErr.Code, appErr.Message, appErr.Error())
		return
	}

	// Extract user input
	userInput := h.extractUserInput(msg)
	if userInput == "" {
		h.sendError(w, req.ID, models.InvalidParams, "Invalid params", "No text content found in message")
		return
	}

	// Process request
	result, err := h.processRequest(userInput, msg)
	if err != nil {
		appErr := err.(*apperrors.AppError)
		h.sendError(w, req.ID, appErr.Code, appErr.Message, appErr.Error())
		return
	}

	h.sendSuccess(w, req.ID, result)
}

func (h *A2AHandler) extractMessage(params map[string]interface{}) (*models.Message, error) {
	messageData, ok := params["message"].(map[string]interface{})
	if !ok {
		return nil, apperrors.NewAppError(
			models.InvalidParams,
			"Missing or invalid 'message' parameter",
			nil,
		)
	}

	messageJSON, _ := json.Marshal(messageData)
	var msg models.Message
	if err := json.Unmarshal(messageJSON, &msg); err != nil {
		return nil, apperrors.NewAppError(
			models.InvalidParams,
			"Invalid message format",
			err,
		)
	}

	return &msg, nil
}

func (h *A2AHandler) extractUserInput(msg *models.Message) string {
	for _, part := range msg.Parts {
		if part.Kind == models.KindText {
			return part.Text
		}
	}
	return ""
}

func (h *A2AHandler) processRequest(input string, userMsg *models.Message) (*models.TaskResult, error) {
	var flashcards *models.FlashcardSet
	var err error

	// Check if input contains PDF URL
	if pdfURL := h.flashcardService.ExtractPDFURL(input); pdfURL != "" {
		log.Printf("Processing PDF from URL: %s", pdfURL)
		flashcards, err = h.flashcardService.GenerateFromURL(pdfURL)
	} else {
		// Check if message contains a data part with PDF content
		pdfData := h.extractPDFData(userMsg)
		if pdfData != nil {
			log.Printf("Processing uploaded PDF (%d bytes)", len(pdfData))
			flashcards, err = h.flashcardService.GenerateFromPDFData(pdfData)
		} else {
			log.Printf("Generating flashcards from text input")
			flashcards, err = h.flashcardService.GenerateFromText(input)
		}
	}

	if err != nil {
		return nil, err
	}

	// Build response
	return h.buildTaskResult(flashcards, userMsg), nil
}

func (h *A2AHandler) extractPDFData(msg *models.Message) []byte {
	for _, part := range msg.Parts {
		if part.Kind == models.KindData {
			// Check if data is a base64 encoded PDF or raw bytes
			if dataMap, ok := part.Data.(map[string]interface{}); ok {
				// Handle base64 encoded PDF
				if contentType, ok := dataMap["contentType"].(string); ok {
					if contentType == "application/pdf" {
						if base64Data, ok := dataMap["data"].(string); ok {
							// Decode base64 to bytes
							// For now, return the string as bytes
							// In production, properly decode base64
							return []byte(base64Data)
						}
					}
				}
				// Handle direct byte array
				if pdfBytes, ok := dataMap["pdf"].([]byte); ok {
					return pdfBytes
				}
			}
			// Handle if data is already bytes
			if bytes, ok := part.Data.([]byte); ok {
				return bytes
			}
		}
	}
	return nil
}

func (h *A2AHandler) buildTaskResult(flashcards *models.FlashcardSet, userMsg *models.Message) *models.TaskResult {
	// Generate IDs
	taskID := userMsg.TaskID
	if taskID == "" {
		taskID = fmt.Sprintf("task-%s", uuid.New().String()[:8])
	}

	responseText := h.flashcardService.FormatAsText(flashcards)
	responseMsg := models.Message{
		Kind:      models.KindMessage,
		Role:      models.RoleAgent,
		MessageID: fmt.Sprintf("msg-%s", uuid.New().String()[:8]),
		Parts: []models.MessagePart{
			{
				Kind: models.KindText,
				Text: responseText,
			},
		},
	}

	artifacts := []models.Artifact{
		{
			ArtifactID: fmt.Sprintf("artifact-%s", uuid.New().String()),
			Name:       "flashcardSet",
			Parts: []models.MessagePart{
				{
					Kind: models.KindData,
					Data: flashcards,
				},
			},
		},
	}

	return &models.TaskResult{
		ID:        taskID,
		ContextID: fmt.Sprintf("ctx-%s", uuid.New().String()),
		Status: models.Status{
			State:     models.StateCompleted,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Message:   responseMsg,
		},
		Artifacts: artifacts,
		History: []models.Message{
			*userMsg,
			responseMsg,
		},
		Kind: "task",
	}
}

func (h *A2AHandler) sendSuccess(w http.ResponseWriter, id string, result interface{}) {
	response := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	h.writeResponse(w, response, http.StatusOK)
}

func (h *A2AHandler) sendError(w http.ResponseWriter, id string, code int, message, data string) {
	response := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &models.RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	h.writeResponse(w, response, http.StatusOK)
}

func (h *A2AHandler) writeResponse(w http.ResponseWriter, response models.JSONRPCResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}