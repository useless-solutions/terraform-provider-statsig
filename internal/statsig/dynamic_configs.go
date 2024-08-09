package statsig

import (
	"context"
	"encoding/json"
	"fmt"
)

type DynamicConfig struct {
	ID               string              `json:"id"`
	Description      string              `json:"description"`
	IDType           string              `json:"idType"`
	LastModifierID   string              `json:"lastModifierID"`
	LastModifierName string              `json:"lastModifierName"`
	CreatorEmail     string              `json:"creatorEmail"`
	CreatorName      string              `json:"creatorName"`
	CreatedTime      int64               `json:"createdTime"`
	HoldoutIDs       []string            `json:"holdoutIDs"`
	IsEnabled        bool                `json:"isEnabled"`
	Rules            []DynamicConfigRule `json:"rules"`
	DefaultValue     DynamicConfigValue  `json:"defaultValue"`
	Tags             []string            `json:"tags"`
}

type DynamicConfigRule struct {
	Name           string                   `json:"name"`
	PassPercentage int                      `json:"passPercentage"`
	Conditions     []DynamicConfigCondition `json:"conditions"`
	ReturnValue    DynamicConfigValue       `json:"returnValue"`
}

type DynamicConfigCondition struct {
	Type        string `json:"type"`
	Operator    string `json:"operator"`
	TargetValue int    `json:"targetValue"`
	Field       string `json:"field"`
	CustomID    string `json:"customID"`
}

type DynamicConfigValue struct {
	Key interface{} `json:"key"`
}

func (c *Client) GetDynamicConfigs(ctx context.Context) ([]DynamicConfig, error) {
	params := map[string]string{"page": "1", "limit": "100"}
	response, err := c.Get("dynamic_configs", params)
	if err != nil {
		return nil, err
	}

	dynamicConfigs := APIListResponse[DynamicConfig]{}
	if err := json.Unmarshal(response, &dynamicConfigs); err != nil {
		return nil, err
	}

	return dynamicConfigs.Data, nil
}

func (c *Client) GetDynamicConfig(ctx context.Context, id string) (*DynamicConfig, error) {
	response, err := c.Get(fmt.Sprintf("dynamic_configs/%s", id), nil)
	if err != nil {
		return nil, err
	}

	dynamicConfig := APIResponse[DynamicConfig]{}
	if err := json.Unmarshal(response, &dynamicConfig); err != nil {
		return nil, err
	}

	return &dynamicConfig.Data, nil
}
