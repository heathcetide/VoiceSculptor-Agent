package prompt

import (
	"VoiceSculptor/pkg/session"
	"context"
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"testing"
)

// mockToolHandler is a mock implementation of a toolHandler
func mockToolHandler(expectedResult *CallToolResult, expectedError error) toolHandler {
	return func(ctx context.Context, req *CallToolRequest) (*CallToolResult, error) {
		return expectedResult, expectedError
	}
}

func TestRegisterAndGetTool(t *testing.T) {
	manager := newToolManager()

	tool := &Tool{
		Name: "echo",
		InputSchema: &openapi3.Schema{
			Type: &openapi3.Types{openapi3.TypeObject},
		},
	}

	handler := mockToolHandler(&CallToolResult{
		Content: []Content{NewTextContent("ok")},
	}, nil)

	manager.registerTool(tool, handler)

	retrieved, exists := manager.getTool("echo")
	assert.True(t, exists)
	assert.Equal(t, "echo", retrieved.Name)
}

func TestHandleListTools(t *testing.T) {
	manager := newToolManager()

	manager.registerTool(&Tool{
		Name: "tool1",
		InputSchema: &openapi3.Schema{
			Type: &openapi3.Types{openapi3.TypeObject},
		},
	}, nil)

	manager.registerTool(&Tool{
		Name: "tool2",
		InputSchema: &openapi3.Schema{
			Type: &openapi3.Types{openapi3.TypeObject},
		},
	}, nil)

	resp, err := manager.handleListTools(context.Background(), &JSONRPCRequest{}, *session.NewSession())
	assert.NoError(t, err)

	result := resp.(ListToolsResult)
	assert.Len(t, result.Tools, 2)
}

func TestHandleCallTool_Success(t *testing.T) {
	manager := newToolManager()

	manager.registerTool(&Tool{
		Name: "echo",
		InputSchema: &openapi3.Schema{
			Type: &openapi3.Types{openapi3.TypeObject},
		},
	}, mockToolHandler(&CallToolResult{
		Content: []Content{NewTextContent("echoed")},
	}, nil))

	req := &JSONRPCRequest{
		ID: "1",
		Params: map[string]interface{}{
			"name": "echo",
			"arguments": map[string]interface{}{
				"message": "hi",
			},
		},
	}

	resp, err := manager.handleCallTool(context.Background(), req, *session.NewSession())
	assert.NoError(t, err)

	result := resp.(*CallToolResult)
	assert.False(t, result.IsError)
	assert.Equal(t, "echoed", result.Content[0].(TextContent).Text)
}

func TestHandleCallTool_InvalidParams(t *testing.T) {
	manager := newToolManager()

	req := &JSONRPCRequest{
		ID:     "2",
		Params: "not-a-map",
	}

	resp, _ := manager.handleCallTool(context.Background(), req, *session.NewSession())
	assert.Contains(t, resp.(*JSONRPCError).Error.Message, "invalid parameters")
}

func TestHandleCallTool_MissingName(t *testing.T) {
	manager := newToolManager()

	req := &JSONRPCRequest{
		ID:     "3",
		Params: map[string]interface{}{},
	}

	resp, _ := manager.handleCallTool(context.Background(), req, *session.NewSession())
	assert.Contains(t, resp.(*JSONRPCError).Error.Message, "missing tool name")
}

func TestHandleCallTool_ToolNotFound(t *testing.T) {
	manager := newToolManager()

	req := &JSONRPCRequest{
		ID: "4",
		Params: map[string]interface{}{
			"name": "nonexistent",
		},
	}

	resp, _ := manager.handleCallTool(context.Background(), req, *session.NewSession())
	assert.Contains(t, resp.(*JSONRPCError).Error.Message, "tool not found")
}

func TestHandleCallTool_HandlerError(t *testing.T) {
	manager := newToolManager()

	manager.registerTool(&Tool{
		Name: "fail",
		InputSchema: &openapi3.Schema{
			Type: &openapi3.Types{openapi3.TypeObject},
		},
	}, mockToolHandler(nil, errors.New("boom")))

	req := &JSONRPCRequest{
		ID: "5",
		Params: map[string]interface{}{
			"name": "fail",
		},
	}

	resp, _ := manager.handleCallTool(context.Background(), req, *session.NewSession())
	assert.Contains(t, resp.(*JSONRPCError).Error.Message, "tool execution failed")
}
