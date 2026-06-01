package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/cristian-scherer/memoos/internal/logger"
)

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     HandlerFunc            `json:"-"`
}

type HandlerFunc func(ctx context.Context, input json.RawMessage) (interface{}, error)

type Server struct {
	tools   map[string]Tool
	logger  *logger.Logger
	mu      sync.RWMutex
	running bool
}

func NewServer(log *logger.Logger) *Server {
	return &Server{
		tools:  make(map[string]Tool),
		logger: log,
	}
}

func (s *Server) RegisterTool(tool Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
	s.logger.Debugf("Registered tool: %s", tool.Name)
}

func (s *Server) GetTool(name string) (Tool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tool, ok := s.tools[name]
	return tool, ok
}

func (s *Server) ListTools() []Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("Starting MCP server on stdio")

	go s.serveStdio(ctx)

	return nil
}

func (s *Server) serveStdio(ctx context.Context) {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("MCP server stopped")
			return
		default:
		}

		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				s.logger.Info("EOF received, stopping server")
				return
			}
			s.logger.Errorf("Failed to decode message: %v", err)
			continue
		}

		resp := s.handleMessage(ctx, msg)

		if resp != nil {
			if err := encoder.Encode(resp); err != nil {
				s.logger.Errorf("Failed to encode response: %v", err)
			}
		}
	}
}

func (s *Server) handleMessage(ctx context.Context, msg map[string]interface{}) map[string]interface{} {
	method, _ := msg["method"].(string)
	id, hasID := msg["id"]

	if !hasID {
		s.handleNotification(method, msg)
		return nil
	}

	switch method {
	case "initialize":
		return s.handleInitialize(id)

	case "tools/list":
		return s.handleToolsList(id)

	case "tools/call":
		return s.handleToolsCall(ctx, id, msg)

	case "shutdown":
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result":  nil,
		}

	default:
		return errorResponse(id, -32601, fmt.Sprintf("Method not found: %s", method))
	}
}

func (s *Server) handleNotification(method string, msg map[string]interface{}) {
	switch method {
	case "notifications/initialized":
		s.logger.Info("Client initialized")
	case "notifications/cancelled":
		s.logger.Info("Request cancelled")
	default:
		s.logger.Debugf("Received notification: %s", method)
	}
}

func (s *Server) handleInitialize(id interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]interface{}{
				"name":    "memoos",
				"version": "0.1.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
		},
	}
}

func (s *Server) handleToolsList(id interface{}) map[string]interface{} {
	tools := s.ListTools()
	toolList := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		toolList[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": toolList,
		},
	}
}

func (s *Server) handleToolsCall(ctx context.Context, id interface{}, msg map[string]interface{}) map[string]interface{} {
	params, _ := msg["params"].(map[string]interface{})
	toolName, _ := params["name"].(string)
	argumentsBytes, _ := json.Marshal(params["arguments"])

	tool, ok := s.GetTool(toolName)
	if !ok {
		return errorResponse(id, -32602, fmt.Sprintf("Tool not found: %s", toolName))
	}

	result, err := tool.Handler(ctx, argumentsBytes)
	if err != nil {
		return errorResponse(id, -32603, err.Error())
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": toJSONString(result),
				},
			},
		},
	}
}

func errorResponse(id interface{}, code int, message string) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
}

func toJSONString(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
