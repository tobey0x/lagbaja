package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tobey0x/lagbaja/internal/config"
	"github.com/tobey0x/lagbaja/internal/handler"
	"github.com/tobey0x/lagbaja/internal/service"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize services
	pdfService := service.NewPDFService()
	flashcardService := service.NewFlashcardService(pdfService, cfg.APIKey)

	// Initialize handler
	a2aHandler := handler.NewA2AHandler(flashcardService)

	// Setup routes
	mux := http.NewServeMux()
	mux.Handle("/a2a", a2aHandler)
	mux.HandleFunc("/health", healthCheckHandler)
	mux.HandleFunc("/upload", uploadHandler(flashcardService))

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting A2A Flashcard Generator on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"flashcard-generator"}`))
}

func uploadHandler(flashcardService *service.FlashcardService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form (max 10MB)
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		// Get the file from the form
		file, header, err := r.FormFile("pdf")
		if err != nil {
			http.Error(w, "Failed to get PDF file from form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		log.Printf("Received uploaded file: %s (%d bytes)", header.Filename, header.Size)

		// Read file content
		pdfData := make([]byte, header.Size)
		_, err = file.Read(pdfData)
		if err != nil {
			http.Error(w, "Failed to read PDF file", http.StatusInternalServerError)
			return
		}

		// Generate flashcards from PDF
		flashcards, err := flashcardService.GenerateFromPDFData(pdfData)
		if err != nil {
			log.Printf("Error generating flashcards: %v", err)
			http.Error(w, fmt.Sprintf("Failed to generate flashcards: %v", err), http.StatusInternalServerError)
			return
		}

		// Return flashcards as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(flashcards)
	}
}