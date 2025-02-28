package dynamic_configs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/useless-solutions/terraform-provider-statsig/internal/service/utils"
	"github.com/useless-solutions/terraform-provider-statsig/internal/statsig"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &DynamicConfigResource{}
	_ resource.ResourceWithImportState = &DynamicConfigResource{}
	_ resource.ResourceWithConfigure   = &DynamicConfigResource{}
)

func NewDynamicConfigResource() resource.Resource {
	return &DynamicConfigResource{}
}

// DynamicConfigResource defines the resource implementation.
type DynamicConfigResource struct {
	client *statsig.Client
}

func (r *DynamicConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dynamic_config"
}

func (r *DynamicConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Create a Dynamic Config in the Statsig Project.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the dynamic config",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the new dynamic config",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of this dynamic config",
				Optional:            true,
			},
			"id_type": schema.StringAttribute{
				MarkdownDescription: "The type of ID used in the dynamic config",
				Required:            true,
				Validators: []validator.String{
					// One of userID, stableID, or a custom type unknown to the SDK
					stringvalidator.OneOf("userID", "stableID"),
				},
			},
			"last_modifier_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the last user to modify the dynamic config",
			},
			"last_modifier_email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The email of the last user to modify the dynamic config",
			},
			"creator_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the user who created the dynamic config",
			},
			"creator_email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The email of the user who created the dynamic config",
			},
			"target_apps": schema.ListAttribute{
				MarkdownDescription: "The list of Target Apps that this dynamic config applies to",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "The list of tags associated with this dynamic config",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"team": schema.StringAttribute{
				MarkdownDescription: "The team that this dynamic config belongs to",
				Optional:            true,
			},
			"holdout_ids": schema.ListAttribute{
				MarkdownDescription: "The list of IDs of configured Holdout groups",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"is_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the dynamic config is enabled. Otherwise, default values are returned.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			// Configurable rules for the dynamic config
			"rules": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The display name of the rule",
							Required:            true,
						},
						"pass_percentage": schema.Int64Attribute{
							MarkdownDescription: "The percentage of users that should pass this rule",
							Required:            true,
						},
						"conditions": schema.ListNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										MarkdownDescription: "The type of condition",
										Required:            true,
									},
									"operator": schema.StringAttribute{
										MarkdownDescription: "The operator to use for the condition",
										Required:            true,
									},
									"target_value": schema.Int64Attribute{
										MarkdownDescription: "The target value for the condition",
										Required:            true,
									},
									"field": schema.StringAttribute{
										MarkdownDescription: "The field to check for the condition",
										Required:            true,
									},
									"custom_id": schema.StringAttribute{
										MarkdownDescription: "The custom ID for the condition",
										Optional:            true,
									},
								},
							},
						},
						"return_value": schema.MapAttribute{
							Required:            true,
							MarkdownDescription: "The value to return if the rule passes",
							ElementType:         types.StringType,
						},
					},
				},
			},
			"default_value": schema.DynamicAttribute{
				Required:            true,
				MarkdownDescription: "The default value to return if no rules pass",
			},
		},
	}
}

func (r *DynamicConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DynamicConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DynamicConfig

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create list objects separately before populating parent object
	rules := createDynamicConfigRules(plan)

	apiReq := statsig.DynamicConfig{
		ID:                plan.ID.ValueString(),
		Name:              plan.Name.ValueString(),
		Description:       plan.Description.ValueString(),
		IDType:            plan.IDType.ValueString(),
		LastModifierName:  plan.LastModifierName.ValueString(),
		LastModifierEmail: plan.LastModifierEmail.ValueString(),
		CreatorName:       plan.CreatorName.ValueString(),
		CreatorEmail:      plan.CreatorEmail.ValueString(),
		TargetApps:        utils.ConvertStringList(plan.TargetApps),
		Tags:              utils.ConvertStringList(plan.Tags),
		Team:              plan.Team.ValueString(),
		HoldoutIDs:        utils.ConvertStringList(plan.HoldoutIDs),
		IsEnabled:         plan.IsEnabled.ValueBool(),
		Rules:             rules,
		DefaultValue:      plan.DefaultValue,
	}

	dynamicConfig, err := r.client.CreateDynamicConfig(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Dynamic Config",
			fmt.Sprintf("Unable to create dynamic config, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("DynamicConfig created with Name: %s; and ID: %s", dynamicConfig.Name, dynamicConfig.ID))

	// Set the created config's values back into the plan, updating as needed
	setStateFromObject(&plan, dynamicConfig)

	tflog.Trace(ctx, "Updated plan with created DynamicConfig values")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DynamicConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DynamicConfig

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dynamicConfig, err := r.client.GetDynamicConfig(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dynamic Config",
			fmt.Sprintf("Unable to read dynamic config, got error: %s", err),
		)
		return
	}

	// Update the state with the dynamic config attributes
	setStateFromObject(&state, dynamicConfig)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DynamicConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DynamicConfig

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DynamicConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DynamicConfig

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *DynamicConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func createDynamicConfigRules(plan DynamicConfig) []statsig.DynamicConfigRule {
	var rules []statsig.DynamicConfigRule
	for _, rule := range plan.Rules {
		var conditions []statsig.DynamicConfigRuleCondition
		for _, condition := range rule.Conditions {
			conditions = append(conditions, statsig.DynamicConfigRuleCondition{
				Type:        condition.Type,
				Operator:    condition.Operator,
				TargetValue: condition.TargetValue,
				Field:       condition.Field,
			})
		}

		rules = append(rules, statsig.DynamicConfigRule{
			ID:             rule.ID.ValueString(),
			BaseID:         rule.BaseID.ValueString(),
			Name:           rule.Name.ValueString(),
			PassPercentage: int(rule.PassPercentage.ValueInt64()),
			Conditions:     conditions,
			ReturnValue:    rule.ReturnValue,
			Environments:   utils.ConvertStringList(rule.Environments),
		})
	}

	return rules
}

func setStateFromObject(state *DynamicConfig, dynamicConfig *statsig.DynamicConfig) {
	state.ID = types.StringValue(dynamicConfig.ID)
	state.Name = types.StringValue(dynamicConfig.Name)
	state.Description = types.StringValue(dynamicConfig.Description)
	state.IDType = types.StringValue(dynamicConfig.IDType)
	state.LastModifierName = types.StringValue(dynamicConfig.LastModifierName)
	state.LastModifierEmail = types.StringValue(dynamicConfig.LastModifierEmail)
	state.CreatorName = types.StringValue(dynamicConfig.CreatorName)
	state.CreatorEmail = types.StringValue(dynamicConfig.CreatorEmail)
	state.Team = types.StringValue(dynamicConfig.Team)
	state.IsEnabled = types.BoolValue(dynamicConfig.IsEnabled)
	state.DefaultValue = dynamicConfig.DefaultValue

	// Update the list elements in the plan
	for index, app := range dynamicConfig.TargetApps {
		state.TargetApps[index] = types.StringValue(app)
	}
	for index, tag := range dynamicConfig.Tags {
		state.Tags[index] = types.StringValue(tag)
	}
	for index, holdout := range dynamicConfig.HoldoutIDs {
		state.HoldoutIDs[index] = types.StringValue(holdout)
	}

	// Update the rules in the plan
	for ruleIndex, rule := range dynamicConfig.Rules {
		state.Rules[ruleIndex] = DynamicConfigRule{
			ID:             types.StringValue(rule.ID),
			BaseID:         types.StringValue(rule.BaseID),
			Name:           types.StringValue(rule.Name),
			PassPercentage: types.Int64Value(int64(rule.PassPercentage)),
			ReturnValue:    rule.ReturnValue,
		}

		for envIndex, env := range rule.Environments {
			state.Rules[ruleIndex].Environments[envIndex] = types.StringValue(env)
		}

		// Update the conditions in the rule
		for conditionIndex, condition := range rule.Conditions {
			state.Rules[ruleIndex].Conditions[conditionIndex] = DynamicConfigRuleCondition{
				Type:        condition.Type,
				Operator:    condition.Operator,
				TargetValue: condition.TargetValue,
				Field:       condition.Field,
			}
		}
	}
}
