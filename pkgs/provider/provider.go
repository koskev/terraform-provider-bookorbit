package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*bookorbitProvider)(nil)

type (
	bookorbitProvider struct{}

	bookrobitProviderData struct {
		Url      types.String `tfsdk:"url"`
		Username types.String `tfsdk:"username"`
		Password types.String `tfsdk:"password"`
	}
)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &bookorbitProvider{}
	}
}

func (p *bookorbitProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Das konfigurierte Model aus dem Request holen
	var config bookrobitProviderData
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
	resp.ResourceData = config
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
		NewSetupResource,
	}
}

func (p *bookorbitProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}
