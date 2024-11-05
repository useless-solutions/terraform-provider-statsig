package target_apps

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/useless-solutions/terraform-provider-statsig/internal/statsig"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &TargetAppResource{}
	_ resource.ResourceWithImportState = &TargetAppResource{}
	_ resource.ResourceWithConfigure   = &TargetAppResource{}
)

func NewTargetAppResource() resource.Resource {
	return &TargetAppResource{}
}

type TargetAppResource struct {
	client *statsig.Client
}

func (r *TargetAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_target_app"
}

func (r *TargetAppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Create a target_app in the Statsig Project.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the new target_app",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the target_app",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The description of the target_app",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *TargetAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create builds a new target_app with the provided attributes.
//
// The ID of the created target_app is saved into the Terraform state once the value is returned from the API.
// Statsig references objects by Name, which is unique. The ID is not used for identifying unique objects.
func (r *TargetAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TargetApp

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Map the Terraform plan data to the API request model
	apiReq := statsig.TargetAppAPIRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Create the target_app
	target_app, err := r.client.CreateTargetApp(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create target_app, got error: %s", err),
		)
		return
	}

	// Update the plan attributes with the target_app attributes
	plan = TargetApp{
		ID:          types.StringValue(target_app.ID),
		Name:        types.StringValue(target_app.Name),
		Description: types.StringValue(target_app.Description),
	}

	tflog.Trace(ctx, fmt.Sprintf("TargetApp created with Name: %s; and ID: %s", plan.Name, plan.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read fetches the target_app from the API and updates the Terraform state with the target_app attributes.
func (r *TargetAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TargetApp

	// Get the current state of the resource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the target_app from the API
	target_app, err := r.client.GetTargetApp(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			err.Error(),
		)
		return
	}

	// Update the state with the target_app attributes
	state = TargetApp{
		ID:          types.StringValue(target_app.ID),
		Name:        types.StringValue(target_app.Name),
		Description: types.StringValue(target_app.Description),
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update changes the attributes of the target_app as specified in the Terraform plan.
//
// The ID of the target_app is not modified, as it is immutable in the Statsig API. Additionally, the IsCore attribute cannot
// be modified via the API. This is a limitation of the Statsig API.
func (r *TargetAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TargetApp
	var state TargetApp

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
	// apiReq := statsig.TargetAppAPIRequest{
	// 	ID:          plan.ID.ValueString(),
	// 	Name:        plan.Name.ValueString(),
	// 	Description: plan.Description.ValueString(),
	// 	IsCore:      plan.IsCore.ValueBool(),
	// }

	// // Create the target_app
	// target_app, err := r.client.UpdateTargetApp(ctx, state.Name.ValueString(), apiReq)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Updating TargetApp",
	// 		fmt.Sprintf("Unable to update target_app, got error: %s", err),
	// 	)
	// 	return
	// }

	// // Update the plan attributes with the target_app attributes
	// plan = TargetApp{
	// 	ID:          types.StringValue(target_app.ID),
	// 	Name:        types.StringValue(target_app.Name),
	// 	Description: types.StringValue(target_app.Description),
	// 	IsCore:      plan.IsCore, // IsCore is not modifiable via the API. Set the value to the current state.
	// }

	// tflog.Trace(ctx, fmt.Sprintf("TargetApp created with Name: %s; and ID: %s", plan.Name, plan.ID))

	// // Save data into Terraform state
	// resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

}

func (r *TargetAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TargetApp

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if err := r.client.DeleteTargetApp(ctx, state.Name.ValueString()); err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Deleting TargetApp",
	// 		"Unable to delete target_app, unexpected error: "+err.Error(),
	// 	)
	// 	return
	// }
}

// TODO: Need to implement and test this functionality.
func (r *TargetAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
