package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/tobey0x/lagbaja/internal/models"
	apperrors "github.com/tobey0x/lagbaja/pkg/errors"
)

type PDFService struct {
	httpClient *http.Client
}

func NewPDFService() *PDFService {
	return &PDFService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *PDFService) DownloadPDF(url string) ([]byte, error) {
	log.Printf("Downloading PDF from URL: %s", url)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, apperrors.NewAppError(
			models.InternalError,
			"Failed to download PDF",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.NewAppError(
			models.InternalError,
			fmt.Sprintf("Failed to download PDF: status %d", resp.StatusCode),
			nil,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.NewAppError(
			models.InternalError,
			"Failed to read PDF data",
			err,
		)
	}

	return data, nil
}

func (s *PDFService) ExtractText(pdfData []byte) (string, error) {
	log.Printf("Extracting text from PDF (%d bytes)", len(pdfData))

	// Create a reader from the PDF data
	reader := bytes.NewReader(pdfData)

	// Create a new PDF reader
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfData)))
	if err != nil {
		return "", apperrors.NewAppError(
			models.InternalError,
			"Failed to create PDF reader",
			err,
		)
	}

	var textBuilder bytes.Buffer
	totalPages := pdfReader.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", apperrors.NewAppError(
				models.InternalError,
				fmt.Sprintf("Failed to extract text from page %d", pageNum),
				err,
			)
		}

		textBuilder.WriteString(text)
	}

	return textBuilder.String(), nil
}

func (s *PDFService) ValidatePDF(data []byte) error {
	// Check PDF magic number
	if len(data) < 4 {
		return apperrors.NewAppError(
			models.InvalidParams,
			"Invalid PDF: file too small",
			nil,
		)
	}

	if !bytes.HasPrefix(data, []byte("%PDF")) {
		return apperrors.NewAppError(
			models.InvalidParams,
			"Invalid PDF: incorrect file format",
			nil,
		)
	}

	return nil
}