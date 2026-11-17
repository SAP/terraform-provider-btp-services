package utils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

func TestImportIdentity(t *testing.T) {
	ctx := context.Background()
	connection := Connection{Scope: shared.CtmsScope{Endpoint: "https://ctms.example"}}
	for _, tc := range []struct {
		name       string
		id         string
		endpoint   string
		structured bool
		wantError  bool
	}{
		{name: "numeric ID", id: "9007199254740993"},
		{name: "structured identity", id: "9007199254740993", endpoint: connection.Scope.Endpoint, structured: true},
		{name: "wrong endpoint", id: "1", endpoint: "https://other.example", structured: true, wantError: true},
		{name: "invalid numeric ID", id: "01", wantError: true},
		{name: "invalid structured ID", id: "01", endpoint: connection.Scope.Endpoint, structured: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := resource.ImportStateRequest{ID: tc.id}
			if tc.structured {
				req.ID = ""
				req.Identity = &tfsdk.ResourceIdentity{Schema: IdentitySchema()}
				require.False(t, req.Identity.Set(ctx, Identity{
					Endpoint: types.StringValue(tc.endpoint),
					ID:       types.StringValue(tc.id),
				}).HasError())
			}
			resp := resource.ImportStateResponse{
				State: tfsdk.State{Schema: schema.Schema{Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{Computed: true},
				}}},
				Identity: &tfsdk.ResourceIdentity{Schema: IdentitySchema()},
			}
			diags := resp.State.Set(ctx, &struct {
				ID types.String `tfsdk:"id"`
			}{ID: types.StringNull()})
			require.False(t, diags.HasError(), "%v", diags)
			connection.Import(ctx, req, &resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			if tc.wantError {
				return
			}
			var id types.String
			require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
			require.Equal(t, tc.id, id.ValueString())
			var identity Identity
			require.False(t, resp.Identity.Get(ctx, &identity).HasError())
			require.Equal(t, connection.Identity(id), identity)
			require.False(t, connection.CheckIdentity(ctx, resp.Identity, id).HasError())
			require.True(t, connection.CheckIdentity(ctx, resp.Identity, types.StringValue("2")).HasError())
		})
	}
}
