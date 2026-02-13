// Package copilot – message_queue.go handles message bursts with debouncing.
// When a session is already processing, incoming messages are queued and
// combined after a debounce period.
package copilot

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jholhewres/goclaw/pkg/goclaw/channels"
)

const (
	// DefaultDebounceMs is the default debounce delay in milliseconds.
	DefaultDebounceMs = 1000
	// DefaultMaxPending is the default max queued messages per session.
	DefaultMaxPending = 20
	// DedupWindowSec is the window for deduplication (skip same content).
	DedupWindowSec = 5
)

// OnDrainFunc is called when the debounce timer fires with drained messages.
type OnDrainFunc func(sessionID string, msgs []*channels.IncomingMessage)

// MessageQueue handles message bursts with per-session debouncing.
type MessageQueue struct {
	queues     map[string]*sessionQueue
	debounceMs int
	maxPending int
	dedupSec   int
	onDrain    OnDrainFunc
	mu         sync.Mutex
	logger     *slog.Logger
}

// sessionQueue holds pending messages for a single session.
type sessionQueue struct {
	items       []*queuedMessage
	timer       *time.Timer
	lastEnqueue time.Time
	processing  bool
}

// queuedMessage wraps an incoming message with enqueue timestamp.
type queuedMessage struct {
	msg      *channels.IncomingMessage
	enqueued time.Time
}

// NewMessageQueue creates a new message queue.
// onDrain is called when the debounce timer fires with drained messages (may be nil).
func NewMessageQueue(debounceMs, maxPending int, onDrain OnDrainFunc, logger *slog.Logger) *MessageQueue {
	if debounceMs <= 0 {
		debounceMs = DefaultDebounceMs
	}
	if maxPending <= 0 {
		maxPending = DefaultMaxPending
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MessageQueue{
		queues:     make(map[string]*sessionQueue),
		debounceMs: debounceMs,
		maxPending: maxPending,
		dedupSec:   DedupWindowSec,
		onDrain:    onDrain,
		logger:     logger.With("component", "message_queue"),
	}
}

// Enqueue adds a message to the session queue. Returns true if enqueued,
// false if deduplicated (same content within 5 seconds).
func (q *MessageQueue) Enqueue(sessionID string, msg *channels.IncomingMessage) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	sq, ok := q.queues[sessionID]
	if !ok {
		sq = &sessionQueue{
			items: make([]*queuedMessage, 0, 4),
		}
		q.queues[sessionID] = sq
	}

	// Deduplication: skip if same content within dedup window.
	now := time.Now()
	for _, m := range sq.items {
		if m.msg.Content == msg.Content && now.Sub(m.enqueued) < time.Duration(q.dedupSec)*time.Second {
			q.logger.Debug("message deduplicated", "session", sessionID, "content_preview", truncate(msg.Content, 30))
			return false
		}
	}

	// Max queue size: drop oldest when exceeded.
	if len(sq.items) >= q.maxPending {
		sq.items = sq.items[1:]
		q.logger.Warn("message queue full, dropped oldest",
			"session", sessionID,
			"max_pending", q.maxPending,
		)
	}

	sq.items = append(sq.items, &queuedMessage{msg: msg, enqueued: now})
	sq.lastEnqueue = now

	// Start or reset debounce timer.
	dur := time.Duration(q.debounceMs) * time.Millisecond
	if sq.timer != nil {
		sq.timer.Stop()
	}
	sid := sessionID
	sq.timer = time.AfterFunc(dur, func() {
		msgs := q.Drain(sid)
		if len(msgs) > 0 && q.onDrain != nil {
			go q.onDrain(sid, msgs)
		}
	})

	return true
}

// Drain returns and clears pending messages for the session.
func (q *MessageQueue) Drain(sessionID string) []*channels.IncomingMessage {
	q.mu.Lock()
	defer q.mu.Unlock()

	sq, ok := q.queues[sessionID]
	if !ok || len(sq.items) == 0 {
		return nil
	}

	if sq.timer != nil {
		sq.timer.Stop()
		sq.timer = nil
	}

	msgs := make([]*channels.IncomingMessage, len(sq.items))
	for i, m := range sq.items {
		msgs[i] = m.msg
	}
	sq.items = sq.items[:0]
	return msgs
}

// IsProcessing returns true if the session has an active run.
func (q *MessageQueue) IsProcessing(sessionID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	sq, ok := q.queues[sessionID]
	return ok && sq.processing
}

// SetProcessing marks the session as processing or not.
func (q *MessageQueue) SetProcessing(sessionID string, active bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	sq, ok := q.queues[sessionID]
	if !ok {
		sq = &sessionQueue{items: make([]*queuedMessage, 0, 4)}
		q.queues[sessionID] = sq
	}
	sq.processing = active
}

// CombineMessages merges multiple messages into one prompt string.
func (q *MessageQueue) CombineMessages(msgs []*channels.IncomingMessage) string {
	if len(msgs) == 0 {
		return ""
	}
	if len(msgs) == 1 {
		return msgs[0].Content
	}
	var b strings.Builder
	b.WriteString("[Multiple messages received while busy]\n")
	for i, m := range msgs {
		b.WriteString(fmt.Sprintf("%d. %s", i+1, strings.TrimSpace(m.Content)))
		if i < len(msgs)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
