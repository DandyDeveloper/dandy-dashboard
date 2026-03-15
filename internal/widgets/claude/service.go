package claude

import (
	"context"
	"sync"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const systemPrompt = "You are a helpful personal assistant embedded in a personal dashboard. Be concise and friendly."

// ChatRequest is the body expected by the /chat endpoint.
type ChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// Service wraps the Anthropic SDK with server-side session management.
// History is kept server-side so that thinking blocks and compaction markers
// are never lost across turns.
type Service struct {
	client   *anthropic.Client
	mu       sync.RWMutex
	sessions map[string][]anthropic.MessageParam
}

func NewService(apiKey string) *Service {
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Service{
		client:   &c,
		sessions: make(map[string][]anthropic.MessageParam),
	}
}

// ClearSession wipes the history for a given session ID.
func (s *Service) ClearSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// Stream sends a chat message and streams the response as text deltas.
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
	s.mu.RLock()
	prev := s.sessions[req.SessionID]
	s.mu.RUnlock()

	// Copy history and append the new user turn.
	msgs := make([]anthropic.MessageParam, len(prev), len(prev)+1)
	copy(msgs, prev)
	msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(req.Message)))

	adaptive := anthropic.NewThinkingConfigAdaptiveParam()

	stream := s.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeOpus4_6,
		MaxTokens: 16000,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		// Adaptive thinking: Claude decides when and how much to think.
		Thinking: anthropic.ThinkingConfigParamUnion{OfAdaptive: &adaptive},
		Messages: msgs,
	})

	// Accumulate the full response so we capture all content blocks
	// (text, thinking). Accumulate must be called on every event.
	var accumulated anthropic.Message
	for stream.Next() {
		event := stream.Current()
		accumulated.Accumulate(event)
		if text, ok := extractDelta(event); ok {
			textCh <- text
		}
	}

	if err := stream.Err(); err != nil {
		errCh <- err
		return
	}

	// Persist the completed turn. ToParam() converts the full Message —
	// including thinking blocks — into a MessageParam for the next request.
	s.mu.Lock()
	s.sessions[req.SessionID] = append(msgs, accumulated.ToParam())
	s.mu.Unlock()
}

// extractDelta pulls text from a stream event using the SDK's type-switch pattern.
// Thinking deltas are intentionally skipped — only text is sent to the client.
func extractDelta(event anthropic.MessageStreamEventUnion) (string, bool) {
	deltaEvent, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent)
	if !ok {
		return "", false
	}
	textDelta, ok := deltaEvent.Delta.AsAny().(anthropic.TextDelta)
	if !ok {
		return "", false
	}
	return textDelta.Text, true
}
