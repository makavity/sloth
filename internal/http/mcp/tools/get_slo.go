package tools

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	backendapp "github.com/slok/sloth/internal/http/backend/app"
)

type SLOGetter interface {
	GetSLO(ctx context.Context, req backendapp.GetSLORequest) (*backendapp.GetSLOResponse, error)
}

func NewGetSLOTool(app SLOGetter) (*sdkmcp.Tool, sdkmcp.ToolHandlerFor[getSLOToolInput, getSLOToolOutput]) {
	return &sdkmcp.Tool{
		Name:        "get_slo",
		Description: "Get a single SLO with its current budget and alert status.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true},
	}, newGetSLOToolHandler(app)
}

type getSLOToolInput struct {
	SLOID string `json:"slo_id" jsonschema:"required,The SLO ID to retrieve"`
}

type getSLOToolOutput struct {
	SLO listSLOsToolOutputItem `json:"slo" jsonschema:"the requested SLO with its current status"`
}

func newGetSLOToolHandler(app SLOGetter) sdkmcp.ToolHandlerFor[getSLOToolInput, getSLOToolOutput] {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, input getSLOToolInput) (*sdkmcp.CallToolResult, getSLOToolOutput, error) {
		resp, err := app.GetSLO(ctx, backendapp.GetSLORequest{SLOID: input.SLOID})
		if err != nil {
			return nil, getSLOToolOutput{}, err
		}

		return nil, getSLOToolOutput{SLO: mapRealTimeSLOToToolOutputItem(resp.SLO)}, nil
	}
}
