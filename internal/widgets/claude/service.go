package claude

import (
	"context"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const systemPrompt = "You are a helpful personal assistant embedded in a personal dashboard. Be concise and friendly."

// Message represents a single chat turn.
type Message struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

// ChatRequest is the body expected by the /chat endpoint.
type ChatRequest struct {
	Message string    `json:"message"`
	History []Message `json:"history"`
}

// Service wraps the Anthropic SDK.
type Service struct {
	client *anthropic.Client
	model  anthropic.Model
}

func NewService(apiKey string) *Service {
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Service{
		client: c,
		model:  anthropic.ModelClaude3_5SonnetLatest,
	}
}

// Stream sends a chat message and streams the response via a channel.
// Each string sent on textCh is a text delta; errCh receives at most one error.
// Both channels are closed when the goroutine exits.
func (s *Service) Stream(ctx context.Context, req ChatRequest) (<-chan string, <-chan error) {
	textCh := make(chan string, 32)
	errCh := make(chan error, 1)

	go func() {
		defer close(textCh)
		defer close(errCh)
		s.runStream(ctx, req, textCh, errCh)
	}()

	return textCh, errCh
}

func (s *Service) runStream(ctx context.Context, req ChatRequest, textCh chan<- string, errCh chan<- error) {
	msgs := buildMessages(req)

	stream := s.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(s.model),
		MaxTokens: anthropic.Int(4096),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		}),
		Messages: anthropic.F(msgs),
	})

	for stream.Next() {
		if text, ok := extractDelta(stream.Current()); ok {
			textCh <- text
		}
	}

	if err := stream.Err(); err != nil {
		errCh <- err
	}
}

// buildMessages converts a ChatRequest into the SDK message slice.
func buildMessages(req ChatRequest) []anthropic.MessageParam {
	msgs := make([]anthropic.MessageParam, 0, len(req.History)+1)
	for _, m := range req.History {
		if m.Role == "user" {
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		} else {
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}
	return append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(req.Message)))
}

// extractDelta pulls the text out of a stream event, if any.
func extractDelta(event anthropic.MessageStreamEvent) (string, bool) {
	delta, ok := event.Delta.(anthropic.ContentBlockDeltaEventDelta)
	if !ok || delta.Type != anthropic.ContentBlockDeltaEventDeltaTypeTextDelta {
		return "", false
	}
	return delta.Text, true
}
