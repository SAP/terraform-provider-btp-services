// btpservices/provider/cicd/service_package.go

package cicd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	credentials "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/credentials"
)

// ServicePackage wires all CI/CD resources and data sources into the provider.d
type ServicePackage struct{}

func (s ServicePackage) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		credentials.NewBasicAuthResource,
		credentials.NewCloudConnectorResource,
		credentials.NewWebhookSecretResource,
		credentials.NewContainerRegistryResource,
		credentials.NewKubernetesConfigResource,
		credentials.NewBasicAuthCIdPResource,
		credentials.NewCertCIdPResource,
		credentials.NewServiceKeyResource,
		credentials.NewSecretTextResource,
	}
}

func (s ServicePackage) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		credentials.NewCredentialDataSource,
		credentials.NewCredentialsDataSource,
	}
}

func (s ServicePackage) ListResources(_ context.Context) []func() list.ListResource {
	return []func() list.ListResource{
		credentials.NewBasicAuthListResource,
		credentials.NewCloudConnectorListResource,
		credentials.NewWebhookSecretListResource,
		credentials.NewContainerRegistryListResource,
		credentials.NewKubernetesConfigListResource,
		credentials.NewBasicAuthCIdPListResource,
		credentials.NewCertCIdPListResource,
		credentials.NewServiceKeyListResource,
		credentials.NewSecretTextListResource,
	}
}
