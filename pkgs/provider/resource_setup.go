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

var _ resource.Resource = (*Setup)(nil)

type Setup struct {
	client bookrobitnet.BookorbitClient
}

func NewSetupResource() resource.Resource {
	return &Setup{}
}

func (e *Setup) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setup"
}

func (e *Setup) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"email": schema.StringAttribute{
				Required: true,
			},
			"setup_token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"setup_done": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type setupData struct {
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Email      types.String `tfsdk:"email"`
	SetupToken types.String `tfsdk:"setup_token"`
	SetupDone  types.Bool   `tfsdk:"setup_done"`
}

const (
	SETUP_ENDPOINT        = "api/v1/auth/setup"
	SETUP_STATUS_ENDPOINT = "api/v1/auth/setup-status"
)

func (e *Setup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data setupData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create resource using 3rd party API.
	setupData := struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}{
		Username: data.Username.ValueString(),
		Name:     data.Username.ValueString(),
		Password: data.Password.ValueString(),
		Email:    data.Email.ValueString(),
	}
	marshalledData, err := json.Marshal(setupData)
	if err != nil {
		resp.Diagnostics.AddError("Marshalling data", fmt.Sprintf("%v", err))
		return
	}
	respBody, err := e.client.SendRequest("POST", SETUP_ENDPOINT, marshalledData, map[string]string{
		"x-setup-token": data.SetupToken.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Running setup",
			fmt.Sprintf("%v: %s\n", err, respBody))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (e *Setup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data setupData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read resource using 3rd party API.
	respBody, err := e.client.SendRequest("GET", SETUP_STATUS_ENDPOINT, nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Reading setup",
			fmt.Sprintf("%v: %s\n", err, respBody))
		return
	}

	var setupState struct {
		NeedsSetup bool `json:"needsSetup"`
	}
	err = json.Unmarshal([]byte(respBody), &setupState)
	if err != nil {
		resp.Diagnostics.AddError("Unmarshalling setup state response",
			fmt.Sprintf("%v: %s\n", err, respBody))
		return
	}

	data.SetupDone = types.BoolValue(setupState.NeedsSetup)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (e *Setup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	respBody, err := e.client.SendAuthenticatedRequest("PUT",
		fmt.Sprintf("%s/%s", PROVIDER_ENDPOINT, data.Name.ValueString()), marshalledData)
	if err != nil {
		resp.Diagnostics.AddError("Updating oidc provider",
			fmt.Sprintf("%v: %s\nReq %+v", err, respBody, string(marshalledData)))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (e *Setup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data oidcProviderData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (e *Setup) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(bookrobitProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unknown type in url",
			fmt.Sprintf("Expected string, got: %T", req.ProviderData),
		)
		return
	}

	e.client = bookrobitnet.New(config.Url.ValueString(), config.Username.ValueString(), config.Password.ValueString(), true)
}
