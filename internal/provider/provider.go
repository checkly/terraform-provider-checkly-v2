// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &ScaffoldingProvider{}

// ScaffoldingProvider defines the provider implementation.
type ScaffoldingProvider struct {
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

func (p *ScaffoldingProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scaffolding"
	resp.Version = p.version
}

func (p *ScaffoldingProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
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

func (p *ScaffoldingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configure")
	var data ChecklyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("CHECKLY_API_KEY")
	apiUrl := os.Getenv("CHECKLY_API_URL")
	accountId := os.Getenv("CHECKLY_ACCOUNT_ID")
	/*"https://api.checklyhq.com"*/
	if !data.AccountId.IsNull() {
		accountId = data.AccountId.ValueString()
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
	client.SetChecklySource("checklyApiSource")

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ScaffoldingProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEnvironmentVariableResource,
	}
}

func (p *ScaffoldingProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ScaffoldingProvider{
			version: version,
		}
	}
}
