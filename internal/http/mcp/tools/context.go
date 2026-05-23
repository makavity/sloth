package tools

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/slok/sloth/internal/info"
)

func NewContextTool() (*sdkmcp.Tool, sdkmcp.ToolHandlerFor[contextToolInput, contextToolOutput]) {
	return &sdkmcp.Tool{
		Name:        "context",
		Description: "Get context about Sloth and its SLO framework.",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, newContextToolHandler()
}

type contextToolInput struct{}

type contextToolOutput struct {
	Version     string `json:"version" jsonschema:"the running Sloth version"`
	Description string `json:"description" jsonschema:"what Sloth is and what it does"`
}

func newContextToolHandler() sdkmcp.ToolHandlerFor[contextToolInput, contextToolOutput] {
	return func(_ context.Context, _ *sdkmcp.CallToolRequest, _ contextToolInput) (*sdkmcp.CallToolResult, contextToolOutput, error) {
		return nil, contextToolOutput{
			Version:     info.Version,
			Description: "Sloth is a Prometheus SLO framework that helps teams define service level objectives and creates a uniform, standardized layer of low-level Prometheus rules to implement SLOs, including the recording and alerting rules required to measure them.",
		}, nil
	}
}
