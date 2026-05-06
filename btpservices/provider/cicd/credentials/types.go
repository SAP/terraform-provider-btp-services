// btpservices/provider/cicd/credentials/types.go

package cicdcredentials

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
)

// ---------------------------------------------------------------------------
// Basic Auth
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Cloud Connector
// ---------------------------------------------------------------------------

// cloudConnectorResourceModel is the Terraform state model for the cloud-connector credential resource.
type cloudConnectorResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	LocationID  types.String `tfsdk:"location_id"`
}

func cloudConnectorResourceValueFrom(v cicdmodels.Credential) cloudConnectorResourceModel {
	m := cloudConnectorResourceModel{
		ID:          types.StringValue(v.ID),
		Name:        types.StringValue(v.Name),
		Description: types.StringValue(v.Description),
	}
	if v.CloudConnector != nil {
		m.LocationID = types.StringValue(v.CloudConnector.LocationID)
	}
	return m
}

func (m cloudConnectorResourceModel) toCreateRequest() cicdmodels.CreateCredentialRequest {
	return cicdmodels.CreateCredentialRequest{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		CloudConnector: &cicdmodels.CloudConnector{
			LocationID: m.LocationID.ValueString(),
		},
	}
}

func (m cloudConnectorResourceModel) toPatchRequest() cicdmodels.PatchCredentialRequest {
	name := m.Name.ValueString()
	desc := m.Description.ValueString()
	return cicdmodels.PatchCredentialRequest{
		Name:        &name,
		Description: &desc,
		CloudConnector: &cicdmodels.CloudConnector{
			LocationID: m.LocationID.ValueString(),
		},
	}
}

// ---------------------------------------------------------------------------
// Webhook Secret
// ---------------------------------------------------------------------------

// webhookSecretResourceModel is the Terraform state model for the webhook-secret credential resource.
type webhookSecretResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	// Token is never returned by the API on reads — preserved from prior state.
	Token types.String `tfsdk:"token"`
}

func webhookSecretResourceValueFrom(v cicdmodels.Credential) webhookSecretResourceModel {
	return webhookSecretResourceModel{
		ID:          types.StringValue(v.ID),
		Name:        types.StringValue(v.Name),
		Description: types.StringValue(v.Description),
	}
}

func (m webhookSecretResourceModel) toCreateRequest() cicdmodels.CreateCredentialRequest {
	return cicdmodels.CreateCredentialRequest{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		WebhookToken: &cicdmodels.WebhookToken{
			Token: m.Token.ValueString(),
		},
	}
}

func (m webhookSecretResourceModel) toPatchRequest() cicdmodels.PatchCredentialRequest {
	name := m.Name.ValueString()
	desc := m.Description.ValueString()
	return cicdmodels.PatchCredentialRequest{
		Name:        &name,
		Description: &desc,
		WebhookToken: &cicdmodels.WebhookToken{
			Token: m.Token.ValueString(),
		},
	}
}

// ---------------------------------------------------------------------------
// Container Registry
// ---------------------------------------------------------------------------

// containerRegistryResourceModel is the Terraform state model for the container-registry credential resource.
type containerRegistryResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	// Content is never returned by the API on reads — preserved from prior state.
	Content types.String `tfsdk:"content"`
}

func containerRegistryResourceValueFrom(v cicdmodels.Credential) containerRegistryResourceModel {
	return containerRegistryResourceModel{
		ID:          types.StringValue(v.ID),
		Name:        types.StringValue(v.Name),
		Description: types.StringValue(v.Description),
	}
}

func (m containerRegistryResourceModel) toCreateRequest() cicdmodels.CreateCredentialRequest {
	return cicdmodels.CreateCredentialRequest{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		ContainerRegistryConfiguration: &cicdmodels.ContainerRegistryConfiguration{
			Content: m.Content.ValueString(),
		},
	}
}

func (m containerRegistryResourceModel) toPatchRequest() cicdmodels.PatchCredentialRequest {
	name := m.Name.ValueString()
	desc := m.Description.ValueString()
	return cicdmodels.PatchCredentialRequest{
		Name:        &name,
		Description: &desc,
		ContainerRegistryConfiguration: &cicdmodels.ContainerRegistryConfiguration{
			Content: m.Content.ValueString(),
		},
	}
}

// ---------------------------------------------------------------------------
// Kubernetes Config
// ---------------------------------------------------------------------------

// kubernetesConfigResourceModel is the Terraform state model for the kubernetes-config credential resource.
type kubernetesConfigResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	// Content is never returned by the API on reads — preserved from prior state.
	Content types.String `tfsdk:"content"`
}

func kubernetesConfigResourceValueFrom(v cicdmodels.Credential) kubernetesConfigResourceModel {
	return kubernetesConfigResourceModel{
		ID:          types.StringValue(v.ID),
		Name:        types.StringValue(v.Name),
		Description: types.StringValue(v.Description),
	}
}

func (m kubernetesConfigResourceModel) toCreateRequest() cicdmodels.CreateCredentialRequest {
	return cicdmodels.CreateCredentialRequest{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		KubernetesConfiguration: &cicdmodels.KubernetesConfiguration{
			Content: m.Content.ValueString(),
		},
	}
}

func (m kubernetesConfigResourceModel) toPatchRequest() cicdmodels.PatchCredentialRequest {
	name := m.Name.ValueString()
	desc := m.Description.ValueString()
	return cicdmodels.PatchCredentialRequest{
		Name:        &name,
		Description: &desc,
		KubernetesConfiguration: &cicdmodels.KubernetesConfiguration{
			Content: m.Content.ValueString(),
		},
	}
}

// ---------------------------------------------------------------------------
// Data source models (shared)
// ---------------------------------------------------------------------------

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
