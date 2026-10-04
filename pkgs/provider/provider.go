package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*bookorbitProvider)(nil)

type (
	bookorbitProvider struct{}

	bookrobitProviderModel struct {
		Url types.String `tfsdk:"url"`
	}
)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &bookorbitProvider{}
	}
}

func (p *bookorbitProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Das konfigurierte Model aus dem Request holen
	var config bookrobitProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Url.IsNull() || len(config.Url.String()) == 0 {
		resp.Diagnostics.AddError(
			"Url is empty!",
			"",
		)
		return
	}
	resp.ResourceData = strings.Trim(config.Url.String(), "\"")
}

func (p *bookorbitProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bookorbit"
}

func (p *bookorbitProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (p *bookorbitProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResource,
	}
}

func (p *bookorbitProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required: true,
			},
		},
	}
}
