# Example A2A Requests

## 1. Text Input Request

Generate flashcards from plain text about a specific topic.

```json
{
  "jsonrpc": "2.0",
  "id": "request-text-001",
  "method": "message/send",
  "params": {
    "message": {
      "kind": "message",
      "role": "user",
      "parts": [
        {
          "kind": "text",
          "text": "Create flashcards about the solar system. The solar system consists of the Sun and the objects that orbit it, including eight planets, their moons, and various other celestial bodies like asteroids and comets."
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

## 2. PDF URL Request

Generate flashcards from a PDF document available at a URL.

```json
{
  "jsonrpc": "2.0",
  "id": "request-pdf-url-001",
  "method": "message/send",
  "params": {
    "message": {
      "kind": "message",
      "role": "user",
      "parts": [
        {
          "kind": "text",
          "text": "Please analyze https://example.com/study-notes.pdf and generate flashcards"
        }
      ],
      "messageId": "msg-002",
      "taskId": "task-002"
    },
    "configuration": {
      "blocking": true
    }
  }
}
```

## 3. Complex Topic Request

Generate flashcards on a more complex topic.

```json
{
  "jsonrpc": "2.0",
  "id": "request-complex-001",
  "method": "message/send",
  "params": {
    "message": {
      "kind": "message",
      "role": "user",
      "parts": [
        {
          "kind": "text",
          "text": "Generate comprehensive flashcards about machine learning fundamentals including supervised learning, unsupervised learning, neural networks, and common algorithms."
        }
      ],
      "messageId": "msg-003",
      "taskId": "task-003"
    }
  }
}
```

## 4. Expected Success Response Structure

```json
{
  "jsonrpc": "2.0",
  "id": "request-001",
  "result": {
    "id": "task-001",
    "contextId": "ctx-550e8400-e29b-41d4-a716-446655440000",
    "status": {
      "state": "completed",
      "timestamp": "2025-11-07T12:00:00.000Z",
      "message": {
        "messageId": "msg-response-001",
        "role": "agent",
        "parts": [
          {
            "kind": "text",
            "text": "# Study Flashcards\n\nGenerated 5 flashcards from: user_input\n\n**Card 1** (Definition)\nQ: What is machine learning?\nA: Machine learning is a subset of artificial intelligence...\n\n**Card 2** (Concept)\nQ: What is supervised learning?\nA: Supervised learning is a type of machine learning..."
          }
        ],
        "kind": "message"
      }
    },
    "artifacts": [
      {
        "artifactId": "artifact-550e8400-e29b-41d4-a716-446655440001",
        "name": "flashcardSet",
        "parts": [
          {
            "kind": "data",
            "data": {
              "title": "Study Flashcards",
              "flashcards": [
                {
                  "question": "What is machine learning?",
                  "answer": "Machine learning is a subset of artificial intelligence that enables computers to learn and improve from experience without being explicitly programmed.",
                  "topic": "Definition"
                },
                {
                  "question": "What is supervised learning?",
                  "answer": "Supervised learning is a type of machine learning where the algorithm learns from labeled training data to make predictions or decisions.",
                  "topic": "Concept"
                },
                {
                  "question": "What is unsupervised learning?",
                  "answer": "Unsupervised learning is a type of machine learning that finds patterns in data without labeled responses, often used for clustering and dimensionality reduction.",
                  "topic": "Concept"
                },
                {
                  "question": "What are neural networks?",
                  "answer": "Neural networks are computing systems inspired by biological neural networks, consisting of interconnected nodes (neurons) that process information.",
                  "topic": "Architecture"
                },
                {
                  "question": "What are common machine learning algorithms?",
                  "answer": "Common algorithms include linear regression, logistic regression, decision trees, random forests, support vector machines, and k-means clustering.",
                  "topic": "Application"
                }
              ],
              "source": "user_input",
              "createdAt": "2025-11-07T12:00:00.000Z",
              "totalCards": 5
            }
          }
        ]
      }
    ],
    "history": [
      {
        "kind": "message",
        "role": "user",
        "parts": [
          {
            "kind": "text",
            "text": "Generate comprehensive flashcards about machine learning fundamentals..."
          }
        ],
        "messageId": "msg-003",
        "taskId": "task-003"
      },
      {
        "kind": "message",
        "role": "agent",
        "parts": [
          {
            "kind": "text",
            "text": "# Study Flashcards\n\nGenerated 5 flashcards..."
          }
        ],
        "messageId": "msg-response-001"
      }
    ],
    "kind": "task"
  }
}
```

## 5. Error Response Examples

### Parse Error (-32700)
```json
{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32700,
    "message": "Parse error",
    "data": "invalid character 'i' looking for beginning of value"
  }
}
```

### Invalid Request (-32600)
```json
{
  "jsonrpc": "2.0",
  "id": "request-001",
  "error": {
    "code": -32600,
    "message": "Invalid Request",
    "data": "jsonrpc must be 2.0"
  }
}
```

### Method Not Found (-32601)
```json
{
  "jsonrpc": "2.0",
  "id": "request-001",
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": "Method unknown/method not supported"
  }
}
```

### Invalid Params (-32602)
```json
{
  "jsonrpc": "2.0",
  "id": "request-001",
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "No text content found in message"
  }
}
```

### Internal Error (-32603)
```json
{
  "jsonrpc": "2.0",
  "id": "request-001",
  "error": {
    "code": -32603,
    "message": "failed to generate flashcards",
    "data": "failed to generate flashcards: API error details..."
  }
}
```

## Testing with curl

Save one of the requests above to a file (e.g., `request.json`) and use:

```bash
curl -X POST http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d @request.json | jq .
```

Or pipe directly:

```bash
cat examples/text_request.json | curl -X POST http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d @- | jq .
```
