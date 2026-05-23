package tools

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	backendapp "github.com/slok/sloth/internal/http/backend/app"
)

type SLOLister interface {
	ListSLOs(ctx context.Context, req backendapp.ListSLOsRequest) (*backendapp.ListSLOsResponse, error)
}

func NewListSLOsTool(app SLOLister) (*sdkmcp.Tool, sdkmcp.ToolHandlerFor[listSLOsToolInput, listSLOsToolOutput]) {
	return &sdkmcp.Tool{
		Name:        "list_slos",
		Description: "List SLOs with filtering, fuzzy search, sorting, and pagination. Search matches SLO IDs and grouped label values, and defaults to 100 results per page.",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, newListSLOsToolHandler(app)
}

type listSLOsToolInput struct {
	ServiceID                   string `json:"service_id,omitempty" jsonschema:"Filter SLOs by service ID"`
	Search                      string `json:"search,omitempty" jsonschema:"Optional fuzzy, case-insensitive search against SLO IDs and grouped label values. Spaces are ignored. This does not search service IDs; use service_id for that"`
	AlertFiring                 bool   `json:"alert_firing,omitempty" jsonschema:"Only include SLOs with firing alerts"`
	PeriodBudgetConsumed        bool   `json:"period_budget_consumed,omitempty" jsonschema:"Only include SLOs that already consumed their budget in the period window"`
	CurrentBurningBudgetOver100 bool   `json:"current_burning_budget_over_100,omitempty" jsonschema:"Only include SLOs currently burning budget over 100 percent"`
	Size                        int    `json:"size,omitempty" jsonschema:"Optional pagination size override, defaults to 100 and is capped at 100"`
	Sort                        string `json:"sort,omitempty" jsonschema:"Sort mode: slo-id-asc, slo-id-desc, service-name-asc, service-name-desc, current-burning-budget-asc, current-burning-budget-desc, budget-burned-window-period-asc, budget-burned-window-period-desc, alert-severity-asc, alert-severity-desc"`
	Cursor                      string `json:"cursor,omitempty" jsonschema:"Pagination cursor returned by a previous response"`
}

type listSLOsToolOutput struct {
	SLOs       []listSLOsToolOutputItem     `json:"slos" jsonschema:"the SLOs that matched the request"`
	Pagination listSLOsToolOutputPagination `json:"pagination" jsonschema:"pagination cursors for requesting the next or previous page"`
}

type listSLOsToolOutputItem struct {
	ID                        string            `json:"id" jsonschema:"the unique SLO ID"`
	SlothID                   string            `json:"sloth_id" jsonschema:"the underlying Sloth SLO identifier"`
	Name                      string            `json:"name" jsonschema:"the SLO name"`
	ServiceID                 string            `json:"service_id" jsonschema:"the service this SLO belongs to"`
	Objective                 float64           `json:"objective" jsonschema:"the objective percentage target"`
	Period                    string            `json:"period" jsonschema:"the SLO evaluation period"`
	IsGrouped                 bool              `json:"is_grouped" jsonschema:"whether this SLO is part of a grouped SLO"`
	GroupLabels               map[string]string `json:"group_labels,omitempty" jsonschema:"the group labels for grouped SLOs"`
	BurningBudgetPercent      float64           `json:"burning_budget_percent" jsonschema:"the current error budget burn percentage"`
	BurnedBudgetWindowPercent float64           `json:"burned_budget_window_percent" jsonschema:"the error budget consumed in the evaluation window"`
	HasPageAlert              bool              `json:"has_page_alert" jsonschema:"whether a page alert is firing"`
	PageAlertName             string            `json:"page_alert_name,omitempty" jsonschema:"the firing page alert name when present"`
	HasWarningAlert           bool              `json:"has_warning_alert" jsonschema:"whether a warning alert is firing"`
	WarningAlertName          string            `json:"warning_alert_name,omitempty" jsonschema:"the firing warning alert name when present"`
}

type listSLOsToolOutputPagination struct {
	NextCursor  string `json:"next_cursor,omitempty" jsonschema:"the cursor to request the next page"`
	PrevCursor  string `json:"prev_cursor,omitempty" jsonschema:"the cursor to request the previous page"`
	HasNext     bool   `json:"has_next" jsonschema:"whether there is a next page"`
	HasPrevious bool   `json:"has_previous" jsonschema:"whether there is a previous page"`
}

func newListSLOsToolHandler(app SLOLister) sdkmcp.ToolHandlerFor[listSLOsToolInput, listSLOsToolOutput] {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, input listSLOsToolInput) (*sdkmcp.CallToolResult, listSLOsToolOutput, error) {
		pageSize := input.Size
		if pageSize <= 0 {
			pageSize = 100
		}

		resp, err := app.ListSLOs(ctx, backendapp.ListSLOsRequest{
			FilterServiceID:                   input.ServiceID,
			FilterSearchInput:                 input.Search,
			FilterAlertFiring:                 input.AlertFiring,
			FilterPeriodBudgetConsumed:        input.PeriodBudgetConsumed,
			FilterCurrentBurningBudgetOver100: input.CurrentBurningBudgetOver100,
			PageSize:                          pageSize,
			SortMode:                          backendapp.SLOListSortMode(input.Sort),
			Cursor:                            input.Cursor,
		})
		if err != nil {
			return nil, listSLOsToolOutput{}, err
		}

		output := listSLOsToolOutput{
			SLOs: make([]listSLOsToolOutputItem, 0, len(resp.SLOs)),
			Pagination: listSLOsToolOutputPagination{
				NextCursor:  resp.PaginationCursors.NextCursor,
				PrevCursor:  resp.PaginationCursors.PrevCursor,
				HasNext:     resp.PaginationCursors.HasNext,
				HasPrevious: resp.PaginationCursors.HasPrevious,
			},
		}

		for _, slo := range resp.SLOs {
			item := listSLOsToolOutputItem{
				ID:                        slo.SLO.ID,
				SlothID:                   slo.SLO.SlothID,
				Name:                      slo.SLO.Name,
				ServiceID:                 slo.SLO.ServiceID,
				Objective:                 slo.SLO.Objective,
				Period:                    slo.SLO.PeriodDuration.String(),
				IsGrouped:                 slo.SLO.IsGrouped,
				GroupLabels:               slo.SLO.GroupLabels,
				BurningBudgetPercent:      slo.Budget.BurningBudgetPercent,
				BurnedBudgetWindowPercent: slo.Budget.BurnedBudgetWindowPercent,
				HasPageAlert:              slo.Alerts.FiringPage != nil,
				HasWarningAlert:           slo.Alerts.FiringWarning != nil,
			}

			if slo.Alerts.FiringPage != nil {
				item.PageAlertName = slo.Alerts.FiringPage.Name
			}
			if slo.Alerts.FiringWarning != nil {
				item.WarningAlertName = slo.Alerts.FiringWarning.Name
			}

			output.SLOs = append(output.SLOs, item)
		}

		return nil, output, nil
	}
}
