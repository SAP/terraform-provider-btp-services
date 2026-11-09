package provider

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

type cicdProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	TokenURL     types.String `tfsdk:"token_url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Timeout      types.Int64  `tfsdk:"timeout"`
}

func cicdProviderBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "Configuration for the SAP BTP CI/CD service.\n\n" +
			"All credentials come from a **service key** created on a CI/CD service instance in your BTP subaccount.\n" +
			"Before configuring this block you need:\n\n" +
			"1. A subaccount entitlement for **Continuous Integration & Delivery**.\n" +
			"2. A service instance of that service (plan `default` or `free`). " +
			"You can create one with the [`btp_subaccount_service_instance`](https://registry.terraform.io/providers/SAP/btp/latest/docs/resources/subaccount_service_instance) resource from the SAP BTP Terraform Provider.\n" +
			"3. A service key on that instance — the JSON it returns maps to the attributes below. " +
			"You can create one with the [`btp_subaccount_service_binding`](https://registry.terraform.io/providers/SAP/btp/latest/docs/resources/subaccount_service_binding) resource.\n\n" +
			"See the Quick Start Guide for step-by-step instructions.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The CI/CD service base URL. Maps to the `api` field in the service key JSON. Env: `BTP_CICD_ENDPOINT`.",
				Optional:            true,
			},
			"token_url": schema.StringAttribute{
				MarkdownDescription: "The OAuth2 token endpoint. Construct by appending `/oauth/token` to the `uaa.url` field in the service key JSON. Env: `BTP_CICD_TOKEN_URL`.",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The OAuth2 client ID. Maps to the `uaa.clientid` field in the service key JSON. Env: `BTP_CICD_CLIENT_ID`.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The OAuth2 client secret. Maps to the `uaa.clientsecret` field in the service key JSON. Env: `BTP_CICD_CLIENT_SECRET`.",
				Optional:            true,
				Sensitive:           true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "HTTP request timeout in seconds. Defaults to 60.",
				Optional:            true,
			},
		},
	}
}

func configureCICD(_ context.Context, data *cicdProviderModel, terraformVersion string, clients *shared.ProviderClients, diags *diag.Diagnostics) {
	cfg := cicdclient.CicdClientConfig{
		Endpoint:     resolveString(data.Endpoint, "BTP_CICD_ENDPOINT"),
		TokenURL:     resolveString(data.TokenURL, "BTP_CICD_TOKEN_URL"),
		ClientID:     resolveString(data.ClientID, "BTP_CICD_CLIENT_ID"),
		ClientSecret: resolveString(data.ClientSecret, "BTP_CICD_CLIENT_SECRET"),
		UserAgent:    buildUserAgent(terraformVersion),
	}

	if !data.Timeout.IsNull() && !data.Timeout.IsUnknown() {
		cfg.Timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	facade, err := cicdclient.NewCicdClientFacade(cfg)
	if err != nil {
		diags.AddError("Invalid CI/CD Configuration", err.Error())
		return
	}

	clients.Cicd = facade
}

func resolveString(v types.String, envKey string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}

	return os.Getenv(envKey)
}
