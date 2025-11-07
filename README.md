# Lagbaja - A2A Flashcard Generator

An Agent-to-Agent (A2A) compliant backend service for generating study flashcards from PDF documents and text using AI.

## Features

- ✅ **A2A Protocol Compliant**: Fully implements JSON-RPC 2.0 and A2A response format
- ✅ **Multiple Input Methods**:
  - PDF URL download and processing
  - Direct PDF file upload
  - Plain text input
- ✅ **AI-Powered Generation**: Uses Google Gemini AI for intelligent flashcard creation
- ✅ **Comprehensive Testing**: Full test coverage for handlers and services
- ✅ **Error Handling**: Robust error handling with standard JSON-RPC error codes

## Prerequisites

- Go 1.25.3 or higher
- Google Gemini API key

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd lagbaja
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file:
```bash
cat > .env << EOF
GEMINI_API_KEY=your-api-key-here
PORT=8080
EOF
```

4. Build and run:
```bash
go build -o lagbaja .
./lagbaja
```

## API Endpoints

### 1. A2A Message Endpoint (JSON-RPC 2.0)

**Endpoint**: `POST /a2a`

**Request Format**:
```json
{
  "jsonrpc": "2.0",
  "id": "request-123",
  "method": "message/send",
  "params": {
    "message": {
      "kind": "message",
      "role": "user",
      "parts": [
        {
          "kind": "text",
          "text": "Generate flashcards about photosynthesis"
        }
      ],
      "messageId": "msg-001",
      "taskId": "task-001"
    },
    "configuration": {
      "blocking": true
    }
  }
}
```

**With PDF URL**:
```json
{
  "jsonrpc": "2.0",
  "id": "request-123",
  "method": "message/send",
  "params": {
    "message": {
      "kind": "message",
      "role": "user",
      "parts": [
        {
          "kind": "text",
          "text": "Generate flashcards from https://example.com/document.pdf"
        }
      ],
      "messageId": "msg-001",
      "taskId": "task-001"
    }
  }
}
```

**Response Format**:
```json
{
  "jsonrpc": "2.0",
  "id": "request-123",
  "result": {
    "id": "task-001",
    "contextId": "ctx-uuid",
    "status": {
      "state": "completed",
      "timestamp": "2025-11-07T10:30:00.000Z",
      "message": {
        "messageId": "msg-002",
        "role": "agent",
        "parts": [
          {
            "kind": "text",
            "text": "# Flashcards from PDF\n\n**Card 1** (Overview)\nQ: What is photosynthesis?\nA: ..."
          }
        ],
        "kind": "message"
      }
    },
    "artifacts": [
      {
        "artifactId": "artifact-uuid",
        "name": "flashcardSet",
        "parts": [
          {
            "kind": "data",
            "data": {
              "title": "Study Flashcards",
              "flashcards": [
                {
                  "question": "What is photosynthesis?",
                  "answer": "The process by which plants...",
                  "topic": "Biology"
                }
              ],
              "source": "user_input",
              "createdAt": "2025-11-07T10:30:00.000Z",
              "totalCards": 5
            }
          }
        ]
      }
    ],
    "history": [...],
    "kind": "task"
  }
}
```

### 2. File Upload Endpoint

**Endpoint**: `POST /upload`

**Request**: Multipart form data with a `pdf` file field

**Example using curl**:
```bash
curl -X POST http://localhost:8080/upload \
  -F "pdf=@/path/to/document.pdf"
```

**Response**:
```json
{
  "title": "Flashcards from PDF",
  "flashcards": [
    {
      "question": "What is the main concept?",
      "answer": "The main concept is...",
      "topic": "Overview"
    }
  ],
  "source": "uploaded_pdf",
  "createdAt": "2025-11-07T10:30:00.000Z",
  "totalCards": 5
}
```

### 3. Health Check Endpoint

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "healthy",
  "service": "flashcard-generator"
}
```

## Testing

Run all tests:
```bash
go test ./... -v
```

Run tests with coverage:
```bash
go test ./... -cover
```

Run specific test suite:
```bash
go test ./internal/handler/... -v
go test ./internal/service/... -v
```

## Usage Examples

### Using curl with A2A endpoint

```bash
curl -X POST http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "test-123",
    "method": "message/send",
    "params": {
      "message": {
        "kind": "message",
        "role": "user",
        "parts": [{
          "kind": "text",
          "text": "Create flashcards about the water cycle"
        }],
        "messageId": "msg-001"
      }
    }
  }'
```

### Upload PDF file

```bash
curl -X POST http://localhost:8080/upload \
  -F "pdf=@study_notes.pdf"
```

### Process PDF from URL

```bash
curl -X POST http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d @- << EOF
{
  "jsonrpc": "2.0",
  "id": "url-test",
  "method": "message/send",
  "params": {
    "message": {
      "kind": "message",
      "role": "user",
      "parts": [{
        "kind": "text",
        "text": "https://example.com/biology-notes.pdf"
      }],
      "messageId": "msg-002"
    }
  }
}
EOF
```

## Error Codes

The service uses standard JSON-RPC 2.0 error codes:

| Code | Message | Description |
|------|---------|-------------|
| -32700 | Parse error | Invalid JSON |
| -32600 | Invalid Request | Missing or invalid jsonrpc field |
| -32601 | Method not found | Unknown method |
| -32602 | Invalid params | Missing or invalid parameters |
| -32603 | Internal error | Server-side processing error |

## Architecture

```
lagbaja/
├── main.go                 # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   │   └── config.go
│   ├── handler/           # HTTP handlers
│   │   ├── a2a_handler.go
│   │   └── a2a_handler_test.go
│   ├── models/            # Data models
│   │   ├── a2a.go        # A2A protocol models
│   │   ├── flashcard.go  # Flashcard models
│   │   └── jsonrpc.go    # JSON-RPC models
│   └── service/           # Business logic
│       ├── flashcard_service.go
│       ├── pdf_service.go
│       └── service_test.go
└── pkg/
    └── errors/            # Error handling
        └── errors.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GEMINI_API_KEY` | Google Gemini API key (required) | - |
| `PORT` | Server port | 8080 |

## Development

### Adding New Features

1. Add models to `internal/models/`
2. Implement business logic in `internal/service/`
3. Add HTTP handlers in `internal/handler/`
4. Write tests alongside your code
5. Update this README

### Running in Development

```bash
# With auto-reload (requires air)
air

# Without auto-reload
go run main.go
```

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]

## Support

For issues and questions, please open an issue on the repository.
