package provider

import (
	"context"
	"github.com/checkly/checkly-go-sdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ChecklyProvider satisfies various provider interfaces.
var _ provider.Provider = &ChecklyProvider{}

// ChecklyProvider defines the provider implementation.
type ChecklyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type ChecklyProviderModel struct {
	ApiKey    types.String `tfsdk:"api_key"`
	ApiUrl    types.String `tfsdk:"api_url"`
	AccountId types.String `tfsdk:"account_id"`
}

func (p *ChecklyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "checkly"
	resp.Version = p.version
}

func (p *ChecklyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{

		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The Checkly AccountId to be used",
				Optional:            true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "The Checkly backend to be used",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The Checkly API-Key to be used",
				Optional:            true,
			},
		},
	}
}

func (p *ChecklyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configure")
	var data ChecklyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("CHECKLY_API_KEY")

	apiUrl := os.Getenv("CHECKLY_API_URL")
	if apiUrl == "" {
		apiUrl = "https://api.checklyhq.com"
	}

	accountId := os.Getenv("CHECKLY_ACCOUNT_ID")
	if accountId == "" && !data.AccountId.IsNull() {
		accountId = data.AccountId.ValueString()
	}
	if accountId == "" {
		resp.Diagnostics.AddError("'account_id' variable is missing.", "Please set the 'CHECKLY_ACCOUNT_ID' environment variable or the terraform variable 'account_id'")
		return
	}

	if !data.ApiUrl.IsNull() {
		apiUrl = data.ApiUrl.ValueString()
	}

	client := checkly.NewClient(
		apiUrl,
		apiKey,
		nil,
		nil,
	)
	checklyApiSource := os.Getenv("CHECKLY_API_SOURCE")
	if checklyApiSource == "" {
		checklyApiSource = "TF"
	}

	client.SetAccountId(accountId)
	client.SetChecklySource(checklyApiSource)

	resp.ResourceData = client
}

func (p *ChecklyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEnvironmentVariableResource,
	}
}

func (p *ChecklyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ChecklyProvider{
			version: version,
		}
	}
}
