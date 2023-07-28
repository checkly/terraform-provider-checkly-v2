package provider

import (
	"context"
	"fmt"
	"github.com/checkly/checkly-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &EnvironmentVariableResource{}
var _ resource.ResourceWithImportState = &EnvironmentVariableResource{}

func NewEnvironmentVariableResource() resource.Resource {
	return &EnvironmentVariableResource{}
}

// EnvironmentVariableResource defines the resource implementation.
type EnvironmentVariableResource struct {
	client checkly.Client
}

// EnvironmentVariableResourceModel describes the resource data model.
type EnvironmentVariableResourceModel struct {
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
	Locked types.Bool   `tfsdk:"locked"`
	Id     types.String `tfsdk:"id"`
}

func (r *EnvironmentVariableResourceModel) ToChecklyEntity() checkly.EnvironmentVariable {
	return checkly.EnvironmentVariable{
		Key:    r.Key.ValueString(),
		Value:  r.Value.ValueString(),
		Locked: r.Locked.ValueBool(),
	}
}

func (r *EnvironmentVariableResourceModel) UpdateWithChecklyEntity(environmentVar *checkly.EnvironmentVariable) {
	r.Key = types.StringValue(environmentVar.Key)
	r.Value = types.StringValue(environmentVar.Value)
	r.Locked = types.BoolValue(environmentVar.Locked)
}

func (r *EnvironmentVariableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable"
}

func (r *EnvironmentVariableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "EnvironmentVariable resource",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "Key of the environment variable",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value of the environment variable",
				Required:            true,
			},
			"locked": schema.BoolAttribute{
				Optional: true,
				//apparently values with a default value need to be computed
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the environment variable is locked or not. Set to true for storing sensitive data.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Id of the environment variable",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *EnvironmentVariableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(checkly.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *checkly.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *EnvironmentVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *EnvironmentVariableResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// save into the Terraform state.
	environmentVariable, err := r.client.CreateEnvironmentVariable(ctx, data.ToChecklyEntity())
	if err != nil {
		resp.Diagnostics.AddError("Creating environment variable with Checkly Go-SDK failed", "Checkly Go-SDK error:"+err.Error())
		return
	}
	data.Id = types.StringValue(environmentVariable.Key)

	tflog.Trace(ctx, "created a new environment variable", map[string]interface{}{"variable": data.Key.ValueString(), "locked": data.Locked.ValueBool(), "id": data.Id.ValueString()})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *EnvironmentVariableResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	environmentVar, err := r.client.GetEnvironmentVariable(ctx, data.Id.ValueString())
	if err != nil && strings.Contains(err.Error(), "404") {
		//if resource is deleted remotely, then mark it as
		//successfully gone by unsetting it's ID
		tflog.Debug(ctx, "environment variable not found, assuming it was deleted",
			map[string]interface{}{"variable": data.Key.ValueString(), "locked": data.Locked.ValueBool(), "id": data.Id.ValueString()})
		data.Id = types.StringValue("")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Getting environment variable with Checkly Go-SDK failed", "Checkly Go-SDK error:"+err.Error())
		return
	}
	tflog.Trace(ctx, "read environment variable", map[string]interface{}{"variable": data.Key.ValueString(), "locked": data.Locked.ValueBool(), "id": data.Id.ValueString()})
	data.UpdateWithChecklyEntity(environmentVar)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *EnvironmentVariableResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "updating environment variable", map[string]interface{}{"variable": data.Key.ValueString(), "locked": data.Locked.ValueBool(), "id": data.Id.ValueString()})
	updatedEnvVar, err := r.client.UpdateEnvironmentVariable(ctx, data.Id.ValueString(), data.ToChecklyEntity())
	if err != nil {
		resp.Diagnostics.AddError("Updating environment variable with Checkly Go-SDK failed", "Checkly Go-SDK error:"+err.Error())
		return
	}
	data.UpdateWithChecklyEntity(updatedEnvVar)
	tflog.Trace(ctx, "updated environment variable", map[string]interface{}{"variable": data.Key.ValueString(), "locked": data.Locked.ValueBool(), "id": data.Id.ValueString()})
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *EnvironmentVariableResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteEnvironmentVariable(ctx, data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Deleting environment variable with Checkly Go-SDK failed", "Checkly Go-SDK error:"+err.Error())
		return
	}
	tflog.Trace(ctx, "deleted environment variable", map[string]interface{}{"variable": data.Key.ValueString(), "locked": data.Locked.ValueBool(), "id": data.Id.ValueString()})
}

func (r *EnvironmentVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
