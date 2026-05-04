// btpservices/provider/cicd/credentials/types.go

package cicdcredentials

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
)

// basicAuthResourceModel is the Terraform state model for the basic-auth credential resource.
type basicAuthResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Username    types.String `tfsdk:"username"`
	// Password is never returned by the API on reads — preserved from prior state.
	Password types.String `tfsdk:"password"`
}

func basicAuthResourceValueFrom(v cicdmodels.Credential) basicAuthResourceModel {
	m := basicAuthResourceModel{
		ID:          types.StringValue(v.ID),
		Name:        types.StringValue(v.Name),
		Description: types.StringValue(v.Description),
	}
	if v.Basic != nil {
		m.Username = types.StringValue(v.Basic.Username)
	}
	return m
}

func (m basicAuthResourceModel) toCreateRequest() cicdmodels.CreateCredentialRequest {
	return cicdmodels.CreateCredentialRequest{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Basic: &cicdmodels.BasicAuth{
			Username: m.Username.ValueString(),
			Password: m.Password.ValueString(),
		},
	}
}

func (m basicAuthResourceModel) toPatchRequest() cicdmodels.PatchCredentialRequest {
	name := m.Name.ValueString()
	desc := m.Description.ValueString()
	return cicdmodels.PatchCredentialRequest{
		Name:        &name,
		Description: &desc,
		Basic: &cicdmodels.BasicAuth{
			Username: m.Username.ValueString(),
			Password: m.Password.ValueString(),
		},
	}
}

// basicAuthDSModel is the Terraform state model for the basic-auth credential data source.
// No password field — the API never returns it.
type basicAuthDSModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// credentialsDSModel is the top-level Terraform state for the credentials list datasource.
type credentialsDSModel struct {
	ID     types.String `tfsdk:"id"`
	Values types.List   `tfsdk:"values"`
}

// credentialsDSItemType is the object type used in the credentials list.
var credentialsDSItemType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
	},
}

func basicCredsDSValueFrom(v cicdmodels.Credential) basicAuthDSModel {
	m := basicAuthDSModel{
		ID:          types.StringValue(v.ID),
		Name:        types.StringValue(v.Name),
		Description: types.StringValue(v.Description),
	}

	return m
}

func credentialsDSItemsFrom(list []cicdmodels.Credential) types.List {
	items := make([]attr.Value, 0, len(list))
	for _, c := range list {
		obj, _ := types.ObjectValue(credentialsDSItemType.AttrTypes, map[string]attr.Value{
			"id":          types.StringValue(c.ID),
			"name":        types.StringValue(c.Name),
			"description": types.StringValue(c.Description),
		})
		items = append(items, obj)
	}
	result, _ := types.ListValue(credentialsDSItemType, items)
	return result
}
