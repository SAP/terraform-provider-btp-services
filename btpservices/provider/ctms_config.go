package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ctmsclient "github.com/SAP/terraform-provider-btp-services/internal/ctms/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

type ctmsProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	TokenURL     types.String `tfsdk:"token_url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Timeout      types.Int64  `tfsdk:"timeout"`
}

func ctmsProviderBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "Configuration for SAP Cloud Transport Management. Requires Terraform 1.14 or later. Configure this " +
			"block (which may be empty when using environment variables) to enable CTMS. Each provider alias has " +
			"its own client and landscape scope.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "CTMS HTTPS service root without /v2. Maps to the `api` field in the service key JSON. Env: `BTP_CTMS_ENDPOINT`.",
			},
			"token_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Complete HTTPS OAuth token endpoint. Construct by appending `/oauth/token` to the `uaa.url` field in the service key JSON. Env: `BTP_CTMS_TOKEN_URL`.",
			},
			"client_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "OAuth client ID. Maps to the `uaa.clientid` field in the service key JSON. Env: `BTP_CTMS_CLIENT_ID`.",
			},
			"client_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "OAuth client secret. Maps to the `uaa.clientsecret` field in the service key JSON. Env: `BTP_CTMS_CLIENT_SECRET`.",
			},
			"timeout": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional HTTP request timeout in seconds. Defaults to 60. This limits requests, not the overall Terraform resource operation.",
			},
		},
	}
}

func configureCTMS(_ context.Context, data *ctmsProviderModel, terraformVersion string, clients *shared.ProviderClients, diags *diag.Diagnostics) {
	if terraformVersion != "" && terraformVersion != "unknown" {
		current, err := version.NewVersion(terraformVersion)
		if err != nil || current.LessThan(version.Must(version.NewVersion("1.14.0"))) {
			diags.AddError("Unsupported Terraform Version", "CTMS requires Terraform 1.14 or later.")
			return
		}
	}

	resolved := make(map[string]bool)
	resolve := func(name string, value types.String) string {
		p := path.Root("ctms").AtName(name)
		if value.IsUnknown() {
			diags.AddAttributeError(p, "Unknown CTMS Configuration", "CTMS configuration must be known before making API requests.")
			return ""
		}

		result := os.Getenv("BTP_CTMS_" + strings.ToUpper(name))
		if !value.IsNull() {
			result = value.ValueString()
		}

		if strings.TrimSpace(result) == "" {
			diags.AddAttributeError(p, "Missing CTMS Configuration", fmt.Sprintf("Set %s or BTP_CTMS_%s.", name, strings.ToUpper(name)))
			return result
		}

		resolved[name] = true
		return result
	}

	cfg := ctmsclient.CtmsClientConfig{
		Endpoint:     strings.TrimRight(resolve("endpoint", data.Endpoint), "/"),
		TokenURL:     resolve("token_url", data.TokenURL),
		ClientID:     resolve("client_id", data.ClientID),
		ClientSecret: resolve("client_secret", data.ClientSecret),
		UserAgent:    buildUserAgent(terraformVersion),
	}

	scope := shared.CtmsScope{
		Endpoint: cfg.Endpoint,
	}

	for name, value := range map[string]string{
		"endpoint":  cfg.Endpoint,
		"token_url": cfg.TokenURL,
	} {
		if !resolved[name] {
			continue
		}
		u, err := url.Parse(value)
		if err != nil || u.Host == "" || u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			diags.AddAttributeError(
				path.Root("ctms").AtName(name),
				"Invalid CTMS URL",
				"Use an absolute HTTPS URL without embedded credentials, a query, or a fragment.",
			)
		} else if name == "endpoint" && strings.HasSuffix(u.Path, "/v2") {
			diags.AddAttributeError(
				path.Root("ctms").AtName(name),
				"Invalid CTMS Endpoint",
				"Use the service root without /v2; the client adds the API version.",
			)
		}
	}

	if data.Timeout.IsUnknown() {
		diags.AddAttributeError(
			path.Root("ctms").AtName("timeout"),
			"Unknown CTMS Timeout",
			"The timeout must be known before configuring the client.",
		)
	} else if !data.Timeout.IsNull() {
		seconds := data.Timeout.ValueInt64()
		if seconds <= 0 || seconds > int64((1<<63-1)/time.Second) {
			diags.AddAttributeError(
				path.Root("ctms").AtName("timeout"),
				"Invalid CTMS Timeout",
				"Use a positive number of seconds representable as a Go duration.",
			)
		} else {
			cfg.Timeout = time.Duration(seconds) * time.Second
		}
	}

	if diags.HasError() {
		return
	}

	clients.Ctms = ctmsclient.NewCtmsClientFacade(cfg)
	clients.CtmsScope = scope
}
