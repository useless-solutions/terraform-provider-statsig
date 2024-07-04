package statsig

// TagsDataSourceModel describes the data source data model.
type TagsDataSourceModel struct {
	Tags []Tag `tfsdk:"tags"`
}

type Tag struct {
	ID          string `tfsdk:"id"`
	Name        string `tfsdk:"name"`
	Description string `tfsdk:"description"`
	IsCore      bool   `tfsdk:"is_core"`
}
