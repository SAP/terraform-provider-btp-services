package ctmsnodes

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

func TestNodeDescriptionValidation(t *testing.T) {
	ctx := context.Background()
	var response resource.SchemaResponse
	NewNodeResource().Schema(ctx, resource.SchemaRequest{}, &response)
	description := response.Schema.Attributes["description"].(schema.StringAttribute)

	for _, tc := range []struct {
		name  string
		value types.String
		valid bool
	}{
		{"omitted", types.StringNull(), true},
		{"unknown", types.StringUnknown(), true},
		{"empty", types.StringValue(""), true},
		{"letters numbers and spaces", types.StringValue("Development Space 123"), true},
		{"allowed punctuation", types.StringValue("-._~:\\/?#[]@!$&()*+,;=%"), true},
		{"maximum length", types.StringValue(strings.Repeat("a", 512)), true},
		{"too long", types.StringValue(strings.Repeat("a", 513)), false},
		{"non Latin", types.StringValue("開発"), false},
		{"accented letter", types.StringValue("développement"), false},
		{"newline", types.StringValue("Development\nSpace"), false},
		{"tab", types.StringValue("Development\tSpace"), false},
		{"quote", types.StringValue("Development \"Space\""), false},
		{"angle brackets", types.StringValue("<Development>"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var result validator.StringResponse
			for _, check := range description.Validators {
				check.ValidateString(ctx, validator.StringRequest{
					Path: path.Root("description"), ConfigValue: tc.value,
				}, &result)
			}
			require.Equal(t, !tc.valid, result.Diagnostics.HasError(), result.Diagnostics)
		})
	}
}

func TestNodeDescriptionResetDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name                string
		requested, returned types.String
		wantError           bool
	}{
		{"reset confirmed", types.StringValue(""), types.StringValue(""), false},
		{"old description returned", types.StringValue(""), types.StringValue("initial"), true},
		{"omitted description", types.StringNull(), types.StringValue("initial"), false},
		{"unknown description", types.StringUnknown(), types.StringValue("initial"), false},
		{"non-empty description", types.StringValue("updated"), types.StringValue("updated"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			diagnostics := nodeDescriptionResetDiagnostics(tc.requested, tc.returned)
			if !tc.wantError {
				require.Empty(t, diagnostics)
				return
			}
			require.Len(t, diagnostics, 1)
			require.Equal(t, diag.SeverityError, diagnostics[0].Severity())
			require.Equal(t, "CTMS Node Description Reset Failed", diagnostics[0].Summary())
			require.Equal(t, path.Root("description"), diagnostics[0].(diag.DiagnosticWithPath).Path())
		})
	}
}

func TestNodeDescriptionValueFrom(t *testing.T) {
	for _, tc := range []struct {
		name        string
		description *string
		want        string
	}{
		{"null", nil, ""},
		{"empty", new(""), ""},
		{"non-empty", new("Development Space"), "Development Space"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, diags := nodeValueFrom(context.Background(), ctmsmodels.NodeResponseTarget{
				Id: new(int64(1)), Name: new("node"), Description: tc.description,
			})
			require.False(t, diags.HasError(), diags)
			require.Equal(t, types.StringValue(tc.want), value.Description)
		})
	}
}
