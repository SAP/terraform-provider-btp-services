package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"

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
	SubaccountID types.String `tfsdk:"subaccount_id"`
	NamedUser    types.String `tfsdk:"named_user"`
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
				MarkdownDescription: "CTMS service root without /v2. Env: BTP_CTMS_ENDPOINT.",
			},
			"token_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Complete OAuth token endpoint. Env: BTP_CTMS_TOKEN_URL.",
			},
			"client_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "OAuth client ID. Env: BTP_CTMS_CLIENT_ID.",
			},
			"client_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "OAuth client secret. Use an ephemeral variable or BTP_CTMS_CLIENT_SECRET to avoid persisting the input in a saved plan.",
			},
			"subaccount_id": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "ID of the subaccount associated with the credentials. Required through configuration or " +
					"BTP_CTMS_SUBACCOUNT_ID. Used with endpoint and object ID for resource identity; it does not select " +
					"the authenticated account and cannot be verified by the landscape API.",
			},
			"named_user": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Attribution shown in CTMS for node/route mutations, including deletion. At most 60 characters. Empty " +
					"uses the authentication user. Changes alone do not cause resource updates. Env: BTP_CTMS_NAMED_USER.",
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

	resolve := func(name string, value types.String, required bool) string {
		p := path.Root("ctms").AtName(name)
		if value.IsUnknown() {
			diags.AddAttributeError(p, "Unknown CTMS Configuration", "CTMS configuration must be known before making API requests.")
			return ""
		}

		result := os.Getenv("BTP_CTMS_" + strings.ToUpper(name))
		if !value.IsNull() {
			result = value.ValueString()
		}

		if required && strings.TrimSpace(result) == "" {
			diags.AddAttributeError(p, "Missing CTMS Configuration", fmt.Sprintf("Set %s or BTP_CTMS_%s.", name, strings.ToUpper(name)))
		}

		return result
	}

	cfg := ctmsclient.CtmsClientConfig{
		Endpoint:     strings.TrimRight(resolve("endpoint", data.Endpoint, true), "/"),
		TokenURL:     resolve("token_url", data.TokenURL, true),
		ClientID:     resolve("client_id", data.ClientID, true),
		ClientSecret: resolve("client_secret", data.ClientSecret, true),
		NamedUser:    resolve("named_user", data.NamedUser, false),
		UserAgent:    buildUserAgent(terraformVersion),
	}

	scope := shared.CtmsScope{
		Endpoint:     cfg.Endpoint,
		SubaccountID: resolve("subaccount_id", data.SubaccountID, true),
	}

	for name, value := range map[string]string{
		"endpoint":  cfg.Endpoint,
		"token_url": cfg.TokenURL,
	} {
		u, err := url.Parse(value)
		if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			diags.AddAttributeError(
				path.Root("ctms").AtName(name),
				"Invalid CTMS URL",
				"Use an absolute HTTP(S) URL without embedded credentials, a query, or a fragment.",
			)
		} else if name == "endpoint" && strings.HasSuffix(u.Path, "/v2") {
			diags.AddAttributeError(
				path.Root("ctms").AtName(name),
				"Invalid CTMS Endpoint",
				"Use the service root without /v2; the client adds the API version.",
			)
		}
	}

	if utf8.RuneCountInString(cfg.NamedUser) > 60 || strings.ContainsFunc(cfg.NamedUser, func(r rune) bool {
		return r < 32 || r == 127
	}) {
		diags.AddAttributeError(
			path.Root("ctms").AtName("named_user"),
			"Invalid CTMS Named User",
			"Use at most 60 characters without HTTP header control characters.",
		)
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
