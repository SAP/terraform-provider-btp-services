package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

func TestProvider_CTMSURLDiagnostics(t *testing.T) {
	for _, field := range []string{"endpoint", "token_url"} {
		t.Run(field, func(t *testing.T) {
			for _, tc := range []struct {
				name, env, summary string
				value              types.String
			}{
				{name: "missing", value: types.StringNull(), summary: "Missing CTMS Configuration"},
				{name: "empty", value: types.StringValue(""), summary: "Missing CTMS Configuration"},
				{name: "whitespace", value: types.StringValue(" \t"), summary: "Missing CTMS Configuration"},
				{name: "whitespace environment", value: types.StringNull(), env: " \t", summary: "Missing CTMS Configuration"},
				{name: "unknown", value: types.StringUnknown(), summary: "Unknown CTMS Configuration"},
				{name: "empty overrides environment", value: types.StringValue(""), env: "https://valid.example", summary: "Missing CTMS Configuration"},
				{name: "unknown overrides environment", value: types.StringUnknown(), env: "https://valid.example", summary: "Unknown CTMS Configuration"},
				{name: "malformed", value: types.StringValue("not-a-url"), summary: "Invalid CTMS URL"},
				{name: "slashes", value: types.StringValue("///"), summary: "Invalid CTMS URL"},
				{name: "HTTP", value: types.StringValue("http://invalid.example"), summary: "Invalid CTMS URL"},
				{name: "valid", value: types.StringValue("https://valid.example")},
				{name: "environment fallback", value: types.StringNull(), env: "https://valid.example"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Setenv("BTP_CTMS_"+strings.ToUpper(field), tc.env)
					data := ctmsProviderModel{
						Endpoint: types.StringValue("https://ctms.example"), TokenURL: types.StringValue("https://auth.example/oauth/token"),
						ClientID: types.StringValue("test-client"), ClientSecret: types.StringValue("test-secret"), Timeout: types.Int64Null(),
					}
					if field == "endpoint" {
						data.Endpoint = tc.value
					} else {
						data.TokenURL = tc.value
					}
					var diagnostics diag.Diagnostics
					clients := &shared.ProviderClients{}
					configureCTMS(context.Background(), &data, "1.14.0", clients, &diagnostics)
					if tc.summary == "" {
						require.Empty(t, diagnostics)
						require.NotNil(t, clients.Ctms)
						return
					}
					require.Nil(t, clients.Ctms)
					require.Len(t, diagnostics, 1)
					require.Equal(t, diag.SeverityError, diagnostics[0].Severity())
					require.Equal(t, tc.summary, diagnostics[0].Summary())
					require.Equal(t, path.Root("ctms").AtName(field), diagnostics[0].(diag.DiagnosticWithPath).Path())
				})
			}
		})
	}
}

func TestProvider_CTMSURLDiagnosticsIndependentFields(t *testing.T) {
	data := ctmsProviderModel{
		Endpoint: types.StringUnknown(), TokenURL: types.StringValue("not-a-url"),
		ClientID: types.StringValue("test-client"), ClientSecret: types.StringValue("test-secret"), Timeout: types.Int64Null(),
	}
	var diagnostics diag.Diagnostics
	clients := &shared.ProviderClients{}
	configureCTMS(context.Background(), &data, "1.14.0", clients, &diagnostics)
	require.Nil(t, clients.Ctms)
	require.Len(t, diagnostics, 2)
	require.Equal(t, "Unknown CTMS Configuration", diagnostics[0].Summary())
	require.Equal(t, path.Root("ctms").AtName("endpoint"), diagnostics[0].(diag.DiagnosticWithPath).Path())
	require.Equal(t, "Invalid CTMS URL", diagnostics[1].Summary())
	require.Equal(t, path.Root("ctms").AtName("token_url"), diagnostics[1].(diag.DiagnosticWithPath).Path())
}
