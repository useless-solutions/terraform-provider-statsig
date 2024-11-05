package statsig

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TargetAppAPIRequest struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Gates          []string `json:"gates"`
	DynamicConfigs []string `json:"dynamicConfigs"`
	Experiments    []string `json:"experiments"`
}

func (c *Client) GetTargetApps(ctx context.Context) ([]TargetAppAPIRequest, error) {
	params := map[string]string{"page": "1", "limit": "100"}
	response, err := c.Get("target_apps", params)
	if err != nil {
		return nil, err
	}

	// Log the response body
	tflog.Debug(ctx, fmt.Sprintf("Response Body: %s", map[string]interface{}{"response": string(response)}))
	targetAppsResponse := APIListResponse[TargetAppAPIRequest]{}
	if err := json.Unmarshal(response, &targetAppsResponse); err != nil {
		return nil, err
	}

	return targetAppsResponse.Data, nil
}

func (c *Client) CreateTargetApp(ctx context.Context, targetApp TargetAppAPIRequest) (*TargetAppAPIRequest, error) {
	response, err := c.Post("target_app", targetApp)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating target app: %s", err))
		return nil, err
	}

	tflog.Debug(ctx, fmt.Sprintf("Response Body: %s", map[string]interface{}{"response": string(response)}))
	createdTargetApp := APIResponse[TargetAppAPIRequest]{}
	if err := json.Unmarshal(response, &createdTargetApp); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error unmarshalling target app response: %s", err))
		return nil, err
	}

	tflog.Trace(ctx, fmt.Sprintf("Target App created with Name: %s; and ID: %s", createdTargetApp.Data.Name, createdTargetApp.Data.Description))

	return &createdTargetApp.Data, nil
}

func (c *Client) GetTargetApp(ctx context.Context, targetAppID string) (*TargetAppAPIRequest, error) {
	response, err := c.Get(fmt.Sprintf("target_apps/%s", targetAppID), nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error getting target app: %s", err))
		return nil, err
	}

	targetApp := APIResponse[TargetAppAPIRequest]{}
	if err := json.Unmarshal(response, &targetApp); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error unmarshalling target app: %s", err))
		return nil, err
	}

	tflog.Trace(ctx, fmt.Sprintf("Target App retrieved with Name: %s; and ID: %s", targetApp.Data.Name, targetApp.Data.ID))
	return &targetApp.Data, nil
}
