# Go Spring AI

A Go library inspired by Spring AI Alibaba, providing a unified API for building AI-powered applications. This library offers model-agnostic abstractions for chat completion, embeddings, image generation, audio processing, RAG (Retrieval-Augmented Generation), tool calling, and more.

## Features

- **Model-Agnostic API**: Switch between different AI providers with minimal code changes
- **Chat Client**: Fluent API for chat interactions with support for streaming
- **Embedding Models**: Generate vector embeddings for text
- **Vector Store**: In-memory vector store with similarity search
- **RAG Support**: Retrieval-Augmented Generation with document processing pipelines
- **Tool/Function Calling**: Enable models to call external functions
- **Prompt Templates**: Dynamic prompt generation with variable substitution
- **Structured Output**: Parse model output into structured types
- **Chat Memory**: Conversation history management
- **Advisors**: Intercept and modify requests/responses (logging, RAG, etc.)

## Installation

```bash
go get github.com/go-spring/ai
```

## Quick Start

### Basic Chat

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/go-spring/ai/core"
    "github.com/go-spring/ai/openai"
)

func main() {
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
    chatModel := openai.NewChatModel(client)
    chatClient := core.NewChatClient(chatModel)

    ctx := context.Background()
    response, err := chatClient.Prompt().
        User("What is the capital of France?").
        CallWithContext(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response.Content())
}
```

### Streaming Chat

```go
stream, err := chatClient.Prompt().
    User("Tell me a story.").
    StreamWithContext(ctx)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    chunk, err := stream.Next()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(chunk.Content)
}
```

### RAG (Retrieval-Augmented Generation)

```go
// Create embedding model and vector store
embeddingModel := openai.NewEmbeddingModel(client)
vectorStore := core.NewSimpleVectorStore(embeddingModel)

// Add documents
ctx := context.Background()
documents := []core.Document{
    core.NewDocument("The Eiffel Tower is in Paris."),
    core.NewDocument("The Great Wall is in China."),
}
vectorStore.Add(ctx, documents)

// Create RAG chain
retriever := &core.VectorStoreRetriever{VectorStore: vectorStore}
ragChain := core.NewRAGChain(chatModel, retriever)

// Query
result, err := ragChain.Run(ctx, "Where is the Eiffel Tower?")
```

### Tool Calling

```go
registry := core.NewToolRegistry()

// Register a tool
weatherTool := core.NewFunctionToolBuilder("get_weather").
    WithDescription("Get weather information.").
    WithParameter("location", "string", "City name", true).
    Build()

registry.Register(weatherTool, func(ctx context.Context, args string) (string, error) {
    // Implement weather lookup
    return "Sunny, 25°C", nil
})

// Use tool-calling model
toolModel := core.NewToolCallingChatModel(chatModel, registry)
prompt := core.NewPromptBuilder().
    User("What's the weather in Beijing?").
    Build()

response, err := toolModel.Call(ctx, prompt)
```

### Structured Output

```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

var person Person
parser := core.NewBeanOutputParser(&person)
generator := core.NewStructuredOutputGenerator(chatModel, parser)

prompt := core.NewPromptBuilder().
    User("Generate a person named Alice who is 30 years old.").
    Build()

generator.Generate(ctx, prompt)
fmt.Printf("Person: %+v\n", person)
```

### Prompt Templates

```go
// Using Go text/template
template := core.NewPromptTemplate("Hello, {{.Name}}! You are {{.Age}} years old.")
result, err := template.Render(map[string]interface{}{
    "Name": "Alice",
    "Age":  30,
})

// Using simple string substitution
template2 := core.NewStringPromptTemplate("Hello, {{name}}!")
result2, err := template2.Render(map[string]interface{}{
    "name": "World",
})
```

### Chat Memory

```go
memory := core.NewInMemoryChatMemory()
conversationID := "user-123"

// Add messages
memory.Add(ctx, conversationID,
    core.Message{Role: core.RoleUser, Content: "My name is Alice."},
    core.Message{Role: core.RoleAssistant, Content: "Nice to meet you!"},
)

// Retrieve messages
messages, err := memory.Get(ctx, conversationID, 10)
```

## Project Structure

```
go-spring/ai/
├── core/           # Core abstractions and implementations
│   ├── model.go    # Model interfaces (ChatModel, EmbeddingModel, etc.)
│   ├── chatclient.go  # Fluent ChatClient API
│   ├── prompt.go   # Prompt templates and builders
│   ├── vectorstore.go # Vector store interfaces and implementations
│   ├── rag.go      # RAG chain and document processing
│   ├── tool.go     # Tool/function calling
│   ├── output.go   # Output parsers
│   ├── memory.go   # Chat memory management
│   └── utils.go    # Utility functions
├── openai/         # OpenAI implementation
│   ├── client.go   # OpenAI API client
│   ├── chat.go     # OpenAI ChatModel
│   ├── embedding.go # OpenAI EmbeddingModel
│   ├── image.go    # OpenAI ImageModel
│   └── audio.go    # OpenAI AudioModel
└── examples/       # Example applications
    ├── basic_chat/
    ├── rag/
    ├── tools/
    ├── structured_output/
    ├── memory/
    └── streaming/
```

## Core Concepts

### Models

The library defines several model interfaces:

- **ChatModel**: For text generation and conversation
- **EmbeddingModel**: For generating vector embeddings
- **ImageModel**: For image generation
- **AudioModel**: For audio transcription and synthesis

### ChatClient

The `ChatClient` provides a fluent API for interacting with chat models:

```go
client := core.NewChatClient(model)
response, err := client.Prompt().
    System("You are a helpful assistant.").
    User("Hello!").
    WithTemperature(0.7).
    WithMaxTokens(100).
    Call()
```

### Advisors

Advisors intercept and modify requests and responses:

- **QuestionAnswerAdvisor**: Implements RAG by augmenting prompts with retrieved documents
- **SimpleLoggerAdvisor**: Logs requests and responses
- **ChatMemoryAdvisor**: Adds conversation history to prompts

### Vector Store

The `VectorStore` interface provides methods for storing and retrieving documents:

```go
type VectorStore interface {
    Add(ctx context.Context, documents []Document) error
    Delete(ctx context.Context, ids []string) error
    SimilaritySearch(ctx context.Context, query string, options SearchOptions) ([]Document, error)
}
```

### Document Processing

The ETL pipeline supports document loading, transformation, and storage:

```go
pipeline := core.NewETLPipeline(loader, writer)
pipeline.AddTransformer(&core.SplitterTransformer{Splitter: splitter})
pipeline.AddTransformer(&core.MetadataTransformer{Metadata: map[string]interface{}{"source": "web"}})
pipeline.Run(ctx)
```

## Configuration

### OpenAI Client Options

```go
client := openai.NewClient(
    apiKey,
    openai.WithBaseURL("https://api.openai.com/v1"),
    openai.WithModel("gpt-4"),
    openai.WithHTTPClient(&http.Client{Timeout: 120 * time.Second}),
)
```

## Error Handling

The library uses Go's standard error handling patterns. All operations return errors that can be checked and handled:

```go
response, err := chatClient.Prompt().User("Hello").Call()
if err != nil {
    // Handle error
    log.Printf("Chat failed: %v", err)
    return
}
```

## Retry and Rate Limiting

The library provides utilities for retry and rate limiting:

```go
config := core.RetryConfig{
    MaxRetries: 3,
    Delay:      time.Second,
    RetryableFn: func(err error) bool {
        return isTemporaryError(err)
    },
}

err := core.Retry(config, func() error {
    _, err := chatModel.Call(ctx, prompt)
    return err
})
```

## Testing

Run the test suite:

```bash
go test ./...
```

## Examples

See the `examples/` directory for complete working examples:

- `basic_chat/` - Basic chat interactions
- `rag/` - Retrieval-Augmented Generation
- `tools/` - Tool/function calling
- `structured_output/` - Structured output parsing
- `memory/` - Chat memory management
- `streaming/` - Streaming responses

## License

MIT License
