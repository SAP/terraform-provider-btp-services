package utils

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	identityschema "github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ctmsclient "github.com/SAP/terraform-provider-btp-services/internal/ctms/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

// Connection is immutable after Configure and shared by each object's operations.
type Connection struct {
	Client *ctmsclient.CtmsClientFacade
	Scope  shared.CtmsScope
}

func (c *Connection) Configure(data any) diag.Diagnostics {
	var diags diag.Diagnostics
	if data == nil {
		return diags
	}

	clients, ok := data.(*shared.ProviderClients)
	if !ok {
		diags.AddError("Unexpected Provider Data Type", fmt.Sprintf("Expected *shared.ProviderClients, got %T.", data))
		return diags
	}

	if clients.Ctms == nil {
		diags.AddError("Missing CTMS Configuration", "Configure a ctms block in the provider to use CTMS.")
		return diags
	}

	if clients.CtmsScope.Endpoint == "" || clients.CtmsScope.SubaccountID == "" {
		diags.AddError("Missing CTMS Scope", "CTMS endpoint and subaccount_id are required for the landscape scope.")
		return diags
	}

	c.Client, c.Scope = clients.Ctms, clients.CtmsScope

	return diags
}

type Identity struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	SubaccountID types.String `tfsdk:"subaccount_id"`
	ID           types.String `tfsdk:"id"`
}

func IdentitySchema() identityschema.Schema {
	return identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"endpoint": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "CTMS service root.",
			},
			"subaccount_id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "Subaccount associated with the provider credentials.",
			},
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "Decimal CTMS object ID.",
			},
		},
	}
}

func (c Connection) Identity(id types.String) Identity {
	return Identity{
		Endpoint:     types.StringValue(c.Scope.Endpoint),
		SubaccountID: types.StringValue(c.Scope.SubaccountID),
		ID:           id,
	}
}

func (c Connection) CheckIdentity(ctx context.Context, identity *tfsdk.ResourceIdentity, id types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	if identity == nil || identity.Raw.IsNull() {
		return diags
	}

	var saved Identity

	diags.Append(identity.Get(ctx, &saved)...)
	if diags.HasError() {
		return diags
	}

	if saved.Endpoint.ValueString() != c.Scope.Endpoint || saved.SubaccountID.ValueString() != c.Scope.SubaccountID || !saved.ID.Equal(id) {
		diags.AddError(
			"CTMS Identity Scope Mismatch",
			"The saved endpoint, subaccount_id, or object ID differs from the configured scope. Use the provider "+
				"alias for the original landscape; changing credentials or scope must not adopt another object.",
		)
	}

	return diags
}

func (c Connection) Import(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	id := types.StringValue(req.ID)

	if req.Identity != nil && !req.Identity.Raw.IsNull() {
		var identity Identity
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
		id = identity.ID
		resp.Diagnostics.Append(c.CheckIdentity(ctx, req.Identity, id)...)
	}

	if _, err := ParseID(id); err != nil {
		resp.Diagnostics.AddError("Invalid CTMS Import ID", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, c.Identity(id))...)
}

func ParseID(v types.String) (int64, error) {

	if v.IsNull() || v.IsUnknown() {
		return 0, fmt.Errorf("a known decimal int64 ID is required")
	}

	id, err := strconv.ParseInt(v.ValueString(), 10, 64)

	if err != nil || strconv.FormatInt(id, 10) != v.ValueString() {
		return 0, fmt.Errorf("ID must be a canonical decimal int64 string")
	}

	return id, nil
}

func IDValue(id *int64) types.String {

	if id == nil {
		return types.StringNull()
	}

	return types.StringValue(strconv.FormatInt(*id, 10))
}

func StringPointer[T ~string](v types.String) *T {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := T(v.ValueString())
	return &value
}

func BoolPointer(v types.Bool) *bool {

	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueBool()
	return &value
}

func EnumValue[T ~string](value *T) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*value))
}

type decimalIDValidator struct{}

func IDValidator() validator.String {
	return decimalIDValidator{}
}

func (decimalIDValidator) Description(context.Context) string {
	return "must be a canonical decimal int64 ID"
}

func (v decimalIDValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (decimalIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := ParseID(req.ConfigValue); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid CTMS ID", err.Error())
	}
}

func MissingObject(kind string, matches int) error {
	return fmt.Errorf("expected exactly one CTMS %s matching the configured name; found %d", kind, matches)
}

func CreateError(kind string, err error) string {
	return fmt.Sprintf(
		"Failed to create CTMS %s: %s. If the server applied the request before the failure, inspect the "+
			"landscape and import its ID; an existing same-name object is never adopted automatically.",
		kind,
		err,
	)
}
