package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"bookorbit-provider/pkgs/bookrobitnet"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*OIDCProvider)(nil)

const PROVIDER_ENDPOINT = "api/v1/app-settings/oidc/providers"

type OIDCProvider struct {
	client bookrobitnet.BookorbitClient
}

func NewResource() resource.Resource {
	return &OIDCProvider{}
}

func (e *OIDCProvider) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_provider"
}

func (e *OIDCProvider) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"client_id": schema.StringAttribute{
				Required: true,
			},
			"issuer_uri": schema.StringAttribute{
				Required: true,
			},
			"client_secret": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

type oidcProviderData struct {
	Name         types.String `tfsdk:"name" json:"slug"`
	ClientID     types.String `tfsdk:"client_id" json:"clientId"`
	IssuerURI    types.String `tfsdk:"issuer_uri" json:"issuerUri"`
	ClientSecret types.String `tfsdk:"client_secret" json:"-"`
}

type bookorbitOidcData struct {
	ClientId     string `json:"clientId"`
	Slug         string `json:"slug,omitempty"`
	IssuerUri    string `json:"issuerUri"`
	ClientSecret string `json:"clientSecret"`
	DisplayName  string `json:"displayName"`
	Enabled      bool   `json:"enabled"`
}

func fromProviderData(data oidcProviderData) bookorbitOidcData {
	return bookorbitOidcData{
		data.ClientID.ValueString(),
		data.Name.ValueString(),
		data.IssuerURI.ValueString(),
		data.ClientSecret.ValueString(),
		data.Name.ValueString(),
		true,
	}
}

func (d *bookorbitOidcData) toProviderData() oidcProviderData {
	return oidcProviderData{
		types.StringValue(d.Slug),
		types.StringValue(d.ClientId),
		types.StringValue(d.IssuerUri),
		types.StringValue(d.ClientSecret),
	}
}

func (e *OIDCProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data oidcProviderData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create resource using 3rd party API.
	marshalledData, err := json.Marshal(fromProviderData(data))
	if err != nil {
		resp.Diagnostics.AddError("Marshalling data", fmt.Sprintf("%v", err))
		return
	}
	respBody, err := e.client.SendRequest("POST", PROVIDER_ENDPOINT, marshalledData)
	if err != nil {
		resp.Diagnostics.AddError("Creating oidc provider",
			fmt.Sprintf("%v: %s\n", err, respBody))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (e *OIDCProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data oidcProviderData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read resource using 3rd party API.
	respBody, err := e.client.Get(fmt.Sprintf("%s/%s", PROVIDER_ENDPOINT, data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("reading oidc provider",
			fmt.Sprintf("%v: %s\n", err, respBody))
		return
	}

	var reponseData bookorbitOidcData
	err = json.Unmarshal([]byte(respBody), &reponseData)
	if err != nil {
		resp.Diagnostics.AddError("Unmarshalling read data",
			fmt.Sprintf("%v: %s\n", err, respBody))
		return
	}
	readData := reponseData.toProviderData()
	readData.ClientSecret = data.ClientSecret // Get the old secret as the api returns "***"

	diags = resp.State.Set(ctx, &readData)
	resp.Diagnostics.Append(diags...)
}

func (e *OIDCProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data oidcProviderData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update resource using 3rd party API.
	bookData := fromProviderData(data)
	bookData.Slug = "" // We don't want the slug here since it is already in the url
	marshalledData, err := json.Marshal(bookData)
	if err != nil {
		resp.Diagnostics.AddError("Marshalling data", fmt.Sprintf("%v", err))
		return
	}
	respBody, err := e.client.SendRequest("PUT",
		fmt.Sprintf("%s/%s", PROVIDER_ENDPOINT, data.Name.ValueString()), marshalledData)
	if err != nil {
		resp.Diagnostics.AddError("Updating oidc provider",
			fmt.Sprintf("%v: %s\nReq %+v", err, respBody, string(marshalledData)))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (e *OIDCProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data oidcProviderData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete resource using 3rd party API.
	respBody, err := e.client.SendRequest("DELETE",
		fmt.Sprintf("%s/%s", PROVIDER_ENDPOINT, data.Name.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Deleting oidc provider",
			fmt.Sprintf("%v: %s\n", err, respBody))
		return
	}
}

func (e *OIDCProvider) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	url, ok := req.ProviderData.(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Unknown type in url",
			fmt.Sprintf("Expected string, got: %T", req.ProviderData),
		)
		return
	}

	e.client = bookrobitnet.New(url)
}
