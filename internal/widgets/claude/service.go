package claude

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sync"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const systemPrompt = "You are a helpful personal assistant embedded in a personal dashboard. Be concise and friendly."

const (
	sessionTTL     = 24 * time.Hour
	cleanupInterval = time.Hour
	maxSessionTurns = 100 // prevent unbounded history growth
)

// ChatRequest is the body expected by the /chat endpoint.
type ChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type sessionEntry struct {
	messages []anthropic.MessageParam
	lastUsed time.Time
}

// Service wraps the Anthropic SDK with server-side session management.
// History is kept server-side so that thinking blocks are never lost across turns.
type Service struct {
	client   *anthropic.Client
	log      *slog.Logger
	mu       sync.RWMutex
	sessions map[string]*sessionEntry
}

func NewService(apiKey string, log *slog.Logger) *Service {
	s := &Service{
		client:   func() *anthropic.Client { c := anthropic.NewClient(option.WithAPIKey(apiKey)); return &c }(),
		log:      log,
		sessions: make(map[string]*sessionEntry),
	}
	go s.cleanupLoop()
	return s
}

// cleanupLoop purges sessions that have been idle longer than sessionTTL.
func (s *Service) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		cutoff := time.Now().Add(-sessionTTL)
		for id, entry := range s.sessions {
			if entry.lastUsed.Before(cutoff) {
				delete(s.sessions, id)
				s.log.Info("session expired", "session_id_hash", hashID(id))
			}
		}
		s.mu.Unlock()
	}
}

// ClearSession wipes the history for a given session ID.
func (s *Service) ClearSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	s.log.Info("session cleared", "session_id_hash", hashID(sessionID))
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
	entry := s.sessions[req.SessionID]
	var prev []anthropic.MessageParam
	if entry != nil {
		prev = entry.messages
	}
	s.mu.RUnlock()

	s.log.Info("stream start",
		"session_id_hash", hashID(req.SessionID),
		"msg_len", len(req.Message),
		"history_turns", len(prev),
	)

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

	// Accumulate the full response to preserve all content blocks
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
		s.log.Error("anthropic stream error",
			"session_id_hash", hashID(req.SessionID),
			"error", err,
		)
		errCh <- err
		return
	}

	// Persist the completed turn. Trim to maxSessionTurns to bound memory.
	updated := append(msgs, accumulated.ToParam())
	if len(updated) > maxSessionTurns {
		updated = updated[len(updated)-maxSessionTurns:]
	}

	s.mu.Lock()
	s.sessions[req.SessionID] = &sessionEntry{messages: updated, lastUsed: time.Now()}
	s.mu.Unlock()

	s.log.Info("stream complete",
		"session_id_hash", hashID(req.SessionID),
		"total_turns", len(updated),
	)
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

// hashID returns a short hash of a session ID for safe log output.
func hashID(id string) string {
	sum := sha256.Sum256([]byte(id))
	return fmt.Sprintf("%x", sum[:6])
}
