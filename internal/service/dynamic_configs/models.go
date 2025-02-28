package dynamic_configs

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DynamicConfigsDataSourceModel describes the data source data model.
type DynamicConfigListDataSource struct {
	DynamicConfigs []DynamicConfig `tfsdk:"dynamic_configs"`
}

type DynamicConfig struct {
	ID                types.String        `tfsdk:"id"`
	Name              types.String        `tfsdk:"name"`
	Description       types.String        `tfsdk:"description"`
	IDType            types.String        `tfsdk:"id_type"`
	LastModifierName  types.String        `tfsdk:"last_modifier_name"`
	LastModifierEmail types.String        `tfsdk:"last_modifier_email"`
	CreatorName       types.String        `tfsdk:"creator_name"`
	CreatorEmail      types.String        `tfsdk:"creator_email"`
	TargetApps        []types.String      `tfsdk:"target_apps"`
	Tags              []types.String      `tfsdk:"tags"`
	Team              types.String        `tfsdk:"team"`
	HoldoutIDs        []types.String      `tfsdk:"holdout_ids"`
	IsEnabled         types.Bool          `tfsdk:"is_enabled"`
	Rules             []DynamicConfigRule `tfsdk:"rules"`
	DefaultValue      types.Dynamic       `tfsdk:"default_value"`
}

type DynamicConfigRule struct {
	ID             types.String                 `tfsdk:"id"`
	BaseID         types.String                 `tfsdk:"base_id"`
	Name           types.String                 `tfsdk:"name"`
	PassPercentage types.Int64                  `tfsdk:"pass_percentage"`
	Conditions     []DynamicConfigRuleCondition `tfsdk:"conditions"`
	ReturnValue    types.Dynamic                `tfsdk:"return_value"`
	Environments   []types.String               `tfsdk:"environments"`
}

type DynamicConfigRuleCondition struct {
	Type        string `tfsdk:"type"`
	Operator    string `tfsdk:"operator"`
	TargetValue any    `tfsdk:"target_value"`
	Field       string `tfsdk:"field"`
}
