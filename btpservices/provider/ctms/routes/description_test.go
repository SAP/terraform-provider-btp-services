package ctmsroutes

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

func TestRouteDescriptionValidation(t *testing.T) {
	ctx := context.Background()
	var response resource.SchemaResponse
	NewRouteResource().Schema(ctx, resource.SchemaRequest{}, &response)
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

func TestRouteDescriptionValueFrom(t *testing.T) {
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
			value, diags := routeValueFrom(context.Background(), ctmsmodels.RouteResponse{
				Id: new(int64(1)), Name: new("route"), Description: tc.description,
				SourceNodeId: new(int64(183)), TargetNodeId: new(int64(184)),
			})
			require.False(t, diags.HasError(), diags)
			require.Equal(t, types.StringValue(tc.want), value.Description)
		})
	}
}

func TestRouteDescriptionResetDiagnostics(t *testing.T) {
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
			diags := routeDescriptionResetDiagnostics(tc.requested, tc.returned)
			require.Equal(t, tc.wantError, diags.HasError(), diags)
			if tc.wantError {
				require.Equal(t, "CTMS Route Description Reset Failed", diags[0].Summary())
			}
		})
	}
}
