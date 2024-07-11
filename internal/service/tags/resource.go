package tags

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/useless-solutions/terraform-provider-statsig/internal/statsig"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &TagResource{}
	_ resource.ResourceWithImportState = &TagResource{}
	_ resource.ResourceWithConfigure   = &TagResource{}
)

func NewTagResource() resource.Resource {
	return &TagResource{}
}

type TagResource struct {
	client *statsig.Client
}

func (r *TagResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *TagResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Create a tag in the Statsig Project.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the new tag",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of this tag",
				Required:            true,
			},
			"is_core": schema.BoolAttribute{
				MarkdownDescription: "Whether or not the tag is a Core tag",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the tag",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *TagResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*statsig.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *statsig.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

/*
Create a new tag with the provided attributes.

The ID of the created tag is saved into the Terraform state once the value is returned from the API.
*/
func (r *TagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Tag

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Map the Terraform plan data to the API request model
	apiReq := statsig.TagAPIRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		IsCore:      plan.IsCore.ValueBool(),
	}

	// Create the tag
	tag, err := r.client.CreateTag(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create tag, got error: %s", err),
		)
		return
	}

	// Update the plan attributes with the tag attributes
	plan = Tag{
		ID:          types.StringValue(tag.ID),
		Name:        types.StringValue(tag.Name),
		Description: types.StringValue(tag.Description),
		IsCore:      types.BoolValue(tag.IsCore),
	}

	tflog.Trace(ctx, fmt.Sprintf("Tag created with ID: %s", plan.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

/*
Read the tag from the API and update the Terraform state with the tag attributes.
*/
func (r *TagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Tag

	// Get the current state of the resource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the tag from the API
	tag, err := r.client.GetTag(ctx, state.ID.ValueString()+"fs")
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			err.Error(),
		)
		return
	}

	// Update the state with the tag attributes
	state = Tag{
		ID:          types.StringValue(tag.ID),
		Name:        types.StringValue(tag.Name),
		Description: types.StringValue(tag.Description),
		IsCore:      types.BoolValue(tag.IsCore),
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

/*
The API does not support updating. This resource is immutable.

If the user wants to change a tag, they should create a new one and manually delete the old one in the Console.
This is a limitation of the Statsig API.

This method returns an error to the user to inform them of this limitation.
*/
func (r *TagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Tags are immutable in the Statsig API. If you need to change a tag, create a new one and manually delete the old one in the Console.",
	)
}

func (r *TagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// The API does not support deleting. This resource is immutable.
	// If the user wants to delete a tag, they should do so manually in the Console. Following that, they can remove the tag from the Terraform state.
	// This is a limitation of the Statsig API.
	// We will return an error to the user to inform them of this limitation.
	resp.Diagnostics.AddError(
		"Delete Not Supported",
		"Tags are immutable in the Statsig API. If you need to delete a tag, do so manually in the Console. Following that, remove the tag from the Terraform state. Do this using the `terraform state rm` command.",
	)
}

// TODO: Need to implement and test this functionality.
func (r *TagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
