package ctmsnodes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
	ctmsclient "github.com/SAP/terraform-provider-btp-services/internal/ctms/client"
	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

func TestNodeResourceDescriptionReset(t *testing.T) {
	ctx := context.Background()
	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			for _, tc := range []struct {
				name, response, want string
				wantError            bool
			}{
				{name: "null confirms reset", response: "null"},
				{name: "empty confirms reset", response: `""`},
				{name: "ignored reset", response: `"initial"`, want: "initial", wantError: true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					method, url, payload := http.MethodPatch, "https://ctms.example/v2/nodes/1", `{"description":null}`
					if operation == "create" {
						method, url, payload = http.MethodPost, "https://ctms.example/v2/nodes", `{"name":"node","description":null}`
					}
					writes, reads := 0, 0
					transport := nodeDescriptionTransport(func(req *http.Request) (*http.Response, error) {
						body := fmt.Sprintf(`{"id":1,"name":"node","description":%s}`, tc.response)
						switch {
						case req.Method == http.MethodPost && req.URL.String() == "https://auth.example/oauth/token":
							body = `{"access_token":"test-token","expires_in":3600}`
						case req.Method == method && req.URL.String() == url:
							writes++
							requestBody, err := io.ReadAll(req.Body)
							require.NoError(t, err)
							require.JSONEq(t, payload, string(requestBody))
						case req.Method == http.MethodGet && req.URL.String() == "https://ctms.example/v2/nodes/1":
							reads++
						default:
							t.Fatalf("unexpected HTTP request: %s %s", req.Method, req.URL)
						}
						return &http.Response{
							StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}},
							Body: io.NopCloser(strings.NewReader(body)), Request: req,
						}, nil
					})
					r := &nodeResource{
						Client: ctmsclient.NewCtmsClientFacadeWithHTTP(ctmsclient.CtmsClientConfig{
							Endpoint: "https://ctms.example", TokenURL: "https://auth.example/oauth/token",
							ClientID: "test-client", ClientSecret: "test-secret",
						}, &http.Client{Transport: transport}),
						Scope: shared.CtmsScope{Endpoint: "https://ctms.example"}}
					var schema resource.SchemaResponse
					r.Schema(ctx, resource.SchemaRequest{}, &schema)
					initial, diagnostics := nodeValueFrom(ctx, ctmsmodels.NodeResponseTarget{
						Id: new(int64(1)), Name: new("node"), Description: new("initial"),
					})
					require.Empty(t, diagnostics)
					state := tfsdk.State{Schema: schema.Schema}
					require.Empty(t, state.Set(ctx, initial))
					identity := &tfsdk.ResourceIdentity{Schema: utils.IdentitySchema()}
					require.Empty(t, identity.Set(ctx, r.Identity(initial.ID)))
					planned := initial
					planned.Description = types.StringValue("")
					if operation == "create" {
						planned.ID = types.StringUnknown()
					}
					plan := tfsdk.Plan{Schema: schema.Schema}
					require.Empty(t, plan.Set(ctx, planned))
					if operation == "create" {
						resp := resource.CreateResponse{
							State: tfsdk.State{Schema: schema.Schema}, Identity: &tfsdk.ResourceIdentity{Schema: utils.IdentitySchema()},
						}
						r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
						state, identity, diagnostics = resp.State, resp.Identity, resp.Diagnostics
					} else {
						resp := resource.UpdateResponse{State: state, Identity: identity}
						r.Update(ctx, resource.UpdateRequest{
							Plan: plan, Config: tfsdk.Config{Schema: schema.Schema, Raw: plan.Raw}, State: state, Identity: identity,
						}, &resp)
						state, identity, diagnostics = resp.State, resp.Identity, resp.Diagnostics
					}
					require.Equal(t, 1, writes)
					if tc.wantError {
						require.Len(t, diagnostics, 1)
						require.Equal(t, diag.SeverityError, diagnostics[0].Severity())
						require.Equal(t, "CTMS Node Description Reset Failed", diagnostics[0].Summary())
						require.Equal(t, path.Root("description"), diagnostics[0].(diag.DiagnosticWithPath).Path())
					} else {
						require.Empty(t, diagnostics)
					}
					var saved nodeModel
					require.Empty(t, state.Get(ctx, &saved))
					require.Equal(t, types.StringValue(tc.want), saved.Description)
					require.Equal(t, initial.ID, saved.ID)
					var savedIdentity utils.Identity
					require.Empty(t, identity.Get(ctx, &savedIdentity))
					require.Equal(t, r.Identity(initial.ID), savedIdentity)

					refresh := resource.ReadResponse{State: state, Identity: identity}
					r.Read(ctx, resource.ReadRequest{State: state, Identity: identity}, &refresh)
					require.Empty(t, refresh.Diagnostics)
					require.Equal(t, 1, reads)
					require.True(t, state.Raw.Equal(refresh.State.Raw), "refresh changed the saved state")
				})
			}
		})
	}
}

type nodeDescriptionTransport func(*http.Request) (*http.Response, error)

func (transport nodeDescriptionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport(req)
}
