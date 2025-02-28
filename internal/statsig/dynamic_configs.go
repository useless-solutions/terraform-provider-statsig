package statsig

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DynamicConfig struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	IDType            string `json:"idType"`
	LastModifierName  string `json:"lastModifierName"`
	LastModifierEmail string `json:"lastModifierEmail"`
	CreatorName       string `json:"creatorName"`
	CreatorEmail      string `json:"creatorEmail"`
	// Config related elements
	TargetApps []string `json:"targetApps"`
	Tags       []string `json:"tags"`
	Team       string   `json:"team"`

	HoldoutIDs   []string            `json:"holdoutIDs"`
	IsEnabled    bool                `json:"isEnabled"`
	Rules        []DynamicConfigRule `json:"rules"`
	DefaultValue types.DynamicValue  `json:"defaultValue"`
}

type DynamicConfigRule struct {
	ID             string                       `json:"id"`
	BaseID         string                       `json:"baseID"`
	Name           string                       `json:"name"`
	PassPercentage int                          `json:"passPercentage"`
	Conditions     []DynamicConfigRuleCondition `json:"conditions"`
	ReturnValue    interface{}                  `json:"returnValue"`
	Environments   []string                     `json:"environments"`
}

type DynamicConfigRuleCondition struct {
	Type        string `json:"type"`
	Operator    string `json:"operator"`
	TargetValue any    `json:"targetValue"` // This value can be anything (string, list, int, etc)
	Field       string `json:"field"`
}

const dynamicConfigEndpoint = "dynamic_configs"

func (c *Client) GetAllDynamicConfigs(ctx context.Context) ([]DynamicConfig, error) {
	params := QueryParams{"page": "1", "limit": "100"}
	response, err := c.Get(dynamicConfigEndpoint, params)
	if err != nil {
		return nil, err
	}

	dynamicConfigs := APIListResponse[DynamicConfig]{}
	if err := json.Unmarshal(response, &dynamicConfigs); err != nil {
		return nil, err
	}

	return dynamicConfigs.Data, nil
}

func (c *Client) GetDynamicConfig(ctx context.Context, name string) (*DynamicConfig, error) {
	response, err := c.Get(createEndpointPath(dynamicConfigEndpoint, name), nil)
	if err != nil {
		return nil, err
	}

	dynamicConfig := APIResponse[DynamicConfig]{}
	if err := json.Unmarshal(response, &dynamicConfig); err != nil {
		return nil, err
	}

	return &dynamicConfig.Data, nil
}

func (c *Client) CreateDynamicConfig(ctx context.Context, config DynamicConfig) (*DynamicConfig, error) {
	response, err := c.Post(dynamicConfigEndpoint, config)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating tag: %s", err))
		return nil, err
	}

	// Log the response body
	tflog.Debug(ctx, fmt.Sprintf("Create tag response: %s", response))
	createdConfig := APIResponse[DynamicConfig]{}
	if err := json.Unmarshal(response, &createdConfig); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error unmarshalling tag: %s", err))
		return nil, err
	}

	tflog.Trace(ctx, fmt.Sprintf("Tag created with ID: %s", createdConfig.Data.ID))

	return &createdConfig.Data, nil
}
