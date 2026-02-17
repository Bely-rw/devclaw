// Package mcp implements a Model Context Protocol server that exposes
// DevClaw tools, resources, and prompts to MCP-compatible clients
// (Cursor, VSCode, etc.) via stdio and SSE transports.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

const (
	ProtocolVersion = "2024-11-05"
	ServerName      = "devclaw"
	ServerVersion   = "1.0.0"
)

// Server implements the MCP JSON-RPC 2.0 protocol.
type Server struct {
	logger   *slog.Logger
	tools    []ToolDef
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

// HandlerFunc handles an MCP JSON-RPC request.
type HandlerFunc func(ctx context.Context, params json.RawMessage) (any, error)

// ToolDef describes a tool exposed via MCP.
type ToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ToolCallResult is the result of executing an MCP tool.
type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock is a single content item in a tool result.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Resource describes an MCP resource.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// Prompt describes an MCP prompt template.
type Prompt struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Arguments   []PromptArg   `json:"arguments,omitempty"`
}

// PromptArg describes an argument to a prompt template.
type PromptArg struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// JSON-RPC types
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// New creates a new MCP server.
func New(logger *slog.Logger) *Server {
	s := &Server{
		logger:   logger,
		handlers: make(map[string]HandlerFunc),
	}
	s.registerCoreHandlers()
	return s
}

// RegisterTool adds a tool to the server.
func (s *Server) RegisterTool(def ToolDef, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools = append(s.tools, def)
	s.handlers["tool:"+def.Name] = handler
}

// RegisterHandler adds a custom method handler.
func (s *Server) RegisterHandler(method string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// ServeStdio runs the MCP server over stdin/stdout (JSON-RPC over stdio).
func (s *Server) ServeStdio(ctx context.Context) error {
	s.logger.Info("MCP server starting on stdio")
	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("reading stdin: %w", err)
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeError(writer, nil, -32700, "Parse error")
			continue
		}

		resp := s.handleRequest(ctx, &req)
		if resp != nil {
			data, _ := json.Marshal(resp)
			data = append(data, '\n')
			writer.Write(data)
		}
	}
}

func (s *Server) registerCoreHandlers() {
	s.handlers["initialize"] = s.handleInitialize
	s.handlers["initialized"] = s.handleInitialized
	s.handlers["tools/list"] = s.handleToolsList
	s.handlers["tools/call"] = s.handleToolsCall
	s.handlers["resources/list"] = s.handleResourcesList
	s.handlers["prompts/list"] = s.handlePromptsList
	s.handlers["ping"] = s.handlePing
}

func (s *Server) handleRequest(ctx context.Context, req *jsonRPCRequest) *jsonRPCResponse {
	s.mu.RLock()
	handler, ok := s.handlers[req.Method]
	s.mu.RUnlock()

	if !ok {
		// Notifications (no ID) don't get error responses.
		if req.ID == nil {
			return nil
		}
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		}
	}

	result, err := handler(ctx, req.Params)
	if err != nil {
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCError{Code: -32000, Message: err.Error()},
		}
	}

	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *Server) handleInitialize(_ context.Context, _ json.RawMessage) (any, error) {
	return map[string]any{
		"protocolVersion": ProtocolVersion,
		"capabilities": map[string]any{
			"tools":     map[string]any{"listChanged": false},
			"resources": map[string]any{"subscribe": false, "listChanged": false},
			"prompts":   map[string]any{"listChanged": false},
		},
		"serverInfo": map[string]any{
			"name":    ServerName,
			"version": ServerVersion,
		},
	}, nil
}

func (s *Server) handleInitialized(_ context.Context, _ json.RawMessage) (any, error) {
	s.logger.Info("MCP client initialized")
	return nil, nil
}

func (s *Server) handleToolsList(_ context.Context, _ json.RawMessage) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]any{"tools": s.tools}, nil
}

func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (any, error) {
	var req struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid tool call params: %w", err)
	}

	s.mu.RLock()
	handler, ok := s.handlers["tool:"+req.Name]
	s.mu.RUnlock()

	if !ok {
		return &ToolCallResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", req.Name)}},
			IsError: true,
		}, nil
	}

	argData, _ := json.Marshal(req.Arguments)
	result, err := handler(ctx, argData)
	if err != nil {
		return &ToolCallResult{
			Content: []ContentBlock{{Type: "text", Text: err.Error()}},
			IsError: true,
		}, nil
	}

	text := fmt.Sprintf("%v", result)
	return &ToolCallResult{
		Content: []ContentBlock{{Type: "text", Text: text}},
	}, nil
}

func (s *Server) handleResourcesList(_ context.Context, _ json.RawMessage) (any, error) {
	return map[string]any{"resources": []Resource{}}, nil
}

func (s *Server) handlePromptsList(_ context.Context, _ json.RawMessage) (any, error) {
	prompts := []Prompt{
		{Name: "review", Description: "Review code changes for issues and improvements"},
		{Name: "explain", Description: "Explain code structure and purpose", Arguments: []PromptArg{{Name: "path", Description: "File or directory to explain", Required: true}}},
		{Name: "fix", Description: "Analyze and fix errors in code"},
		{Name: "deploy-check", Description: "Pre-deployment checklist and verification"},
	}
	return map[string]any{"prompts": prompts}, nil
}

func (s *Server) handlePing(_ context.Context, _ json.RawMessage) (any, error) {
	return map[string]any{}, nil
}

func (s *Server) writeError(w io.Writer, id any, code int, message string) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: message},
	}
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	w.Write(data)
}
