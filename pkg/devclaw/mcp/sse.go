// Package mcp – sse.go implements the SSE (Server-Sent Events) transport
// for the MCP server, allowing HTTP-based clients to connect.
package mcp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// SSETransport serves MCP over HTTP with SSE for responses.
type SSETransport struct {
	server   *Server
	logger   *slog.Logger
	sessions sync.Map // sessionID -> *sseSession
}

type sseSession struct {
	id      string
	msgCh   chan []byte
	doneCh  chan struct{}
}

// NewSSETransport creates a new SSE transport wrapping the MCP server.
func NewSSETransport(server *Server, logger *slog.Logger) *SSETransport {
	return &SSETransport{
		server: server,
		logger: logger,
	}
}

// Handler returns an http.Handler that serves the MCP SSE endpoints.
// GET /sse — establishes SSE connection
// POST /message?sessionId=X — sends JSON-RPC messages
func (t *SSETransport) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /sse", t.handleSSE)
	mux.HandleFunc("POST /message", t.handleMessage)
	return mux
}

func (t *SSETransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	sessionID := uuid.New().String()
	sess := &sseSession{
		id:     sessionID,
		msgCh:  make(chan []byte, 64),
		doneCh: make(chan struct{}),
	}
	t.sessions.Store(sessionID, sess)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send endpoint event
	fmt.Fprintf(w, "event: endpoint\ndata: /message?sessionId=%s\n\n", sessionID)
	flusher.Flush()

	t.logger.Info("MCP SSE client connected", "session_id", sessionID)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			t.sessions.Delete(sessionID)
			close(sess.doneCh)
			t.logger.Info("MCP SSE client disconnected", "session_id", sessionID)
			return
		case msg := <-sess.msgCh:
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (t *SSETransport) handleMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "sessionId required", http.StatusBadRequest)
		return
	}

	raw, ok := t.sessions.Load(sessionID)
	if !ok {
		http.Error(w, "unknown session", http.StatusNotFound)
		return
	}
	sess := raw.(*sseSession)

	var req jsonRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON-RPC request", http.StatusBadRequest)
		return
	}

	resp := t.server.handleRequest(r.Context(), &req)
	if resp != nil {
		data, _ := json.Marshal(resp)
		select {
		case sess.msgCh <- data:
		default:
			t.logger.Warn("MCP SSE session buffer full", "session_id", sessionID)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
