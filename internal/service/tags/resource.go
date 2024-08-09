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
		MarkdownDescription: "Create a Tag in the Statsig Project.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the new tag",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of this tag",
				Required:            true,
			},
			// The IsCore attribute is read-only, as only one tag can be a Core tag.
			// TODO: Do not allow the user to set this value. It will default to false.
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

// Create builds a new tag with the provided attributes.
//
// The ID of the created tag is saved into the Terraform state once the value is returned from the API.
// Statsig references objects by Name, which is unique. The ID is not used for identifying unique objects.
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

	tflog.Trace(ctx, fmt.Sprintf("Tag created with Name: %s; and ID: %s", plan.Name, plan.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read fetches the tag from the API and updates the Terraform state with the tag attributes.
func (r *TagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Tag

	// Get the current state of the resource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the tag from the API
	tag, err := r.client.GetTag(ctx, state.Name.ValueString())
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

// Update changes the attributes of the tag as specified in the Terraform plan.
//
// The ID of the tag is not modified, as it is immutable in the Statsig API. Additionally, the IsCore attribute cannot
// be modified via the API. This is a limitation of the Statsig API.
func (r *TagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Tag
	var state Tag

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the current state of the resource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the Terraform plan data to the API request model
	apiReq := statsig.TagAPIRequest{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		IsCore:      plan.IsCore.ValueBool(),
	}

	// Create the tag
	tag, err := r.client.UpdateTag(ctx, state.Name.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Tag",
			fmt.Sprintf("Unable to update tag, got error: %s", err),
		)
		return
	}

	// Update the plan attributes with the tag attributes
	plan = Tag{
		ID:          types.StringValue(tag.ID),
		Name:        types.StringValue(tag.Name),
		Description: types.StringValue(tag.Description),
		IsCore:      plan.IsCore, // IsCore is not modifiable via the API. Set the value to the current state.
	}

	tflog.Trace(ctx, fmt.Sprintf("Tag created with Name: %s; and ID: %s", plan.Name, plan.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *TagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Tag

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTag(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Tag",
			"Unable to delete tag, unexpected error: "+err.Error(),
		)
		return
	}
}

// TODO: Need to implement and test this functionality.
func (r *TagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
