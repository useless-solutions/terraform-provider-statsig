package tags

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TagsDataSourceModel describes the data source data model.
type TagsDataSourceModel struct {
	Tags []Tag `tfsdk:"tags"`
}

type Tag struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsCore      types.Bool   `tfsdk:"is_core"`
}
