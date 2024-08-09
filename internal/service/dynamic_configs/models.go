package dynamic_configs

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DynamicConfigsDataSourceModel describes the data source data model.
type DynamicConfigListDataSource struct {
	DynamicConfigs []DynamicConfig `tfsdk:"tags"`
}

type DynamicConfig struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	// TODO: Add more fields here
}
