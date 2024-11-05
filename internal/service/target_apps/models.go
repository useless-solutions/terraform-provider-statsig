package target_apps

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TagsDataSourceModel describes the data source data model.
type TargetAppsDataSourceModel struct {
	TargetApps []TargetApp `tfsdk:"target_apps"`
}

type TargetApp struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Gates          types.List   `tfsdk:"gates"`
	DynamicConfigs types.List   `tfsdk:"dynamic_configs"`
	Experiments    types.List   `tfsdk:"experiments"`
}
