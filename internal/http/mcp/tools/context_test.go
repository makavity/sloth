package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextTool(t *testing.T) {
	tests := map[string]struct {
		expErr  bool
		expResp contextToolOutput
	}{
		"Tool should expose metadata and static context payload.": {
			expResp: contextToolOutput{
				Version:     "dev",
				Description: "Sloth is a Prometheus SLO framework that helps teams define service level objectives and creates a uniform, standardized layer of low-level Prometheus rules to implement SLOs, including the recording and alerting rules required to measure them.",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tool, handler := NewContextTool()

			require.NotNil(t, tool)
			assert.Equal(t, "context", tool.Name)
			assert.Equal(t, "Get context about Sloth and its SLO framework.", tool.Description)

			result, gotResp, err := handler(context.Background(), nil, contextToolInput{})
			assert.Nil(t, result)

			if test.expErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, test.expResp, gotResp)
		})
	}
}
