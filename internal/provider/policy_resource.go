package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/your-org/terraform-provider-envserver/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &PolicyResource{}
var _ resource.ResourceWithImportState = &PolicyResource{}
var _ resource.ResourceWithConfigValidators = &PolicyResource{}

// NewPolicyResource creates a new policy resource
func NewPolicyResource() resource.Resource {
	return &PolicyResource{}
}

// PolicyResource defines the resource implementation
type PolicyResource struct {
	client *client.Client
}

// PolicyResourceModel describes the resource data model
type PolicyResourceModel struct {
	ID          types.String           `tfsdk:"id"`
	Description types.String           `tfsdk:"description"`
	Policy      types.String           `tfsdk:"policy"`
	Statements  []PolicyStatementModel `tfsdk:"statement"`
	Managed     types.Bool             `tfsdk:"managed"`
}

// PolicyStatementModel represents a single policy statement
type PolicyStatementModel struct {
	Actions  []types.String `tfsdk:"actions"`
	Effect   types.String   `tfsdk:"effect"`
	Resource types.String   `tfsdk:"resource"`
}

// PolicyDocument represents the JSON structure of a policy
type PolicyDocument struct {
	Statements []PolicyStatement `json:"Statements"`
}

// PolicyStatement represents a single statement in the policy JSON
type PolicyStatement struct {
	Actions  []string `json:"actions"`
	Effect   string   `json:"effect"`
	Resource string   `json:"resource"`
}

// Metadata returns the resource type name
func (r *PolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

// Schema defines the schema for the resource
func (r *PolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Environment Server access policy. Policies define permissions for users and service accounts using a statement-based model. You can define the policy either as a JSON/YAML string using the `policy` attribute, or using `statement` blocks. Only one method can be used at a time.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Policy identifier. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Policy description.",
				Required:            true,
			},
			"policy": schema.StringAttribute{
				MarkdownDescription: "Policy document in YAML or JSON format. Defines the access control statements. Cannot be used together with `statement` blocks.",
				Optional:            true,
			},
			"managed": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a managed policy (system-defined). Computed value.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"statement": schema.ListNestedBlock{
				MarkdownDescription: "Policy statement blocks. Cannot be used together with the `policy` attribute.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"actions": schema.ListAttribute{
							MarkdownDescription: "List of actions allowed or denied (e.g., ['projects:get', 'apps:*']).",
							Required:            true,
							ElementType:         types.StringType,
						},
						"effect": schema.StringAttribute{
							MarkdownDescription: "Effect of the statement. Must be 'Allow' or 'Deny'.",
							Required:            true,
						},
						"resource": schema.StringAttribute{
							MarkdownDescription: "Resource ARN pattern (e.g., 'arn:iam:project:my-project' or '.*').",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

// ConfigValidators returns validators for the resource configuration
func (r *PolicyResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("policy"),
			path.MatchRoot("statement"),
		),
	}
}

// Configure adds the provider configured client to the resource
func (r *PolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state
func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert statement blocks to policy JSON if needed
	policyDoc := plan.Policy.ValueString()
	if len(plan.Statements) > 0 {
		var err error
		policyDoc, err = r.statementsToJSON(plan.Statements)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Converting Statements",
				"Could not convert statement blocks to policy JSON: "+err.Error(),
			)
			return
		}
	}

	// Create API request
	createReq := &client.CreatePolicyRequest{
		ID:          plan.ID.ValueString(),
		Description: plan.Description.ValueString(),
		Policy:      policyDoc,
	}

	created, err := r.client.CreatePolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Policy",
			"Could not create policy, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to model
	plan.ID = types.StringValue(created.ID)
	plan.Description = types.StringValue(created.Description)
	plan.Managed = types.BoolValue(created.Managed)

	// Keep the original format (policy string or statements)
	if len(plan.Statements) > 0 {
		// If using statement blocks, populate them from the response
		statements, err := r.jsonToStatements(created.Policy)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Policy",
				"Could not parse policy JSON into statements: "+err.Error(),
			)
			return
		}
		plan.Statements = statements
		plan.Policy = types.StringNull()
	} else {
		// If using policy string, keep it
		plan.Policy = types.StringValue(created.Policy)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state from API
	policy, err := r.client.GetPolicy(ctx, state.ID.ValueString())
	if err != nil {
		// If the policy is not found (404), remove it from state so Terraform will recreate it
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Policy",
			"Could not read policy ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	state.Description = types.StringValue(policy.Description)
	state.Managed = types.BoolValue(policy.Managed)

	// Preserve the original format (policy string or statements)
	if len(state.Statements) > 0 {
		// If using statement blocks, populate them from the response
		statements, err := r.jsonToStatements(policy.Policy)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Policy",
				"Could not parse policy JSON into statements: "+err.Error(),
			)
			return
		}
		state.Statements = statements
		state.Policy = types.StringNull()
	} else {
		// If using policy string, keep it
		state.Policy = types.StringValue(policy.Policy)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert statement blocks to policy JSON if needed
	policyDoc := plan.Policy.ValueString()
	if len(plan.Statements) > 0 {
		var err error
		policyDoc, err = r.statementsToJSON(plan.Statements)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Converting Statements",
				"Could not convert statement blocks to policy JSON: "+err.Error(),
			)
			return
		}
	}

	// Update API request
	description := plan.Description.ValueString()
	updateReq := &client.PatchPolicyRequest{
		Description: &description,
		Policy:      &policyDoc,
	}

	updated, err := r.client.UpdatePolicy(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Policy",
			"Could not update policy ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	plan.Description = types.StringValue(updated.Description)
	plan.Managed = types.BoolValue(updated.Managed)

	// Preserve the original format (policy string or statements)
	if len(plan.Statements) > 0 {
		// If using statement blocks, populate them from the response
		statements, err := r.jsonToStatements(updated.Policy)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Policy",
				"Could not parse policy JSON into statements: "+err.Error(),
			)
			return
		}
		plan.Statements = statements
		plan.Policy = types.StringNull()
	} else {
		// If using policy string, keep it
		plan.Policy = types.StringValue(updated.Policy)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePolicy(ctx, state.ID.ValueString())
	if err != nil {
		// If the policy is already deleted (404), consider it a success
		if client.IsNotFoundError(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Policy",
			"Could not delete policy ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state
func (r *PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// statementsToJSON converts statement blocks to JSON policy document
func (r *PolicyResource) statementsToJSON(statements []PolicyStatementModel) (string, error) {
	policyDoc := PolicyDocument{
		Statements: make([]PolicyStatement, len(statements)),
	}

	for i, stmt := range statements {
		actions := make([]string, len(stmt.Actions))
		for j, action := range stmt.Actions {
			actions[j] = action.ValueString()
		}

		policyDoc.Statements[i] = PolicyStatement{
			Actions:  actions,
			Effect:   stmt.Effect.ValueString(),
			Resource: stmt.Resource.ValueString(),
		}
	}

	jsonBytes, err := json.Marshal(policyDoc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy document: %w", err)
	}

	return string(jsonBytes), nil
}

// jsonToStatements converts JSON policy document to statement blocks
func (r *PolicyResource) jsonToStatements(policyJSON string) ([]PolicyStatementModel, error) {
	var policyDoc PolicyDocument
	if err := json.Unmarshal([]byte(policyJSON), &policyDoc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy document: %w", err)
	}

	statements := make([]PolicyStatementModel, len(policyDoc.Statements))
	for i, stmt := range policyDoc.Statements {
		actions := make([]types.String, len(stmt.Actions))
		for j, action := range stmt.Actions {
			actions[j] = types.StringValue(action)
		}

		statements[i] = PolicyStatementModel{
			Actions:  actions,
			Effect:   types.StringValue(stmt.Effect),
			Resource: types.StringValue(stmt.Resource),
		}
	}

	return statements, nil
}