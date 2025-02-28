package tags

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/useless-solutions/terraform-provider-statsig/internal/statsig"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &TagsDataSource{}
	_ datasource.DataSourceWithConfigure = &TagsDataSource{}
)

func NewTagsDataSource() datasource.DataSource {
	return &TagsDataSource{}
}

// TagsDataSource defines the data source implementation.
type TagsDataSource struct {
	client *statsig.Client
}

func (d *TagsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

func (d *TagsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this resource to retrieve the tags associated with the Statsig Project.",

		Attributes: map[string]schema.Attribute{
			"tags": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"is_core": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *TagsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*statsig.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *statsig.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *TagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TagsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := d.client.GetAllTags(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Statsig Tags, got error: %s", err))
		return
	}

	for _, tag := range tags {
		state.Tags = append(state.Tags, Tag{
			ID:          types.StringValue(tag.ID),
			Name:        types.StringValue(tag.Name),
			Description: types.StringValue(tag.Description),
			IsCore:      types.BoolValue(tag.IsCore),
		})
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
