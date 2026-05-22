// btpservices/provider/cicd/service_package.go

package cicd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	credentials "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/credentials"
	jobs "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/jobs"
	repositories "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/repositories"
)

// ServicePackage wires all CI/CD resources and data sources into the provider.
type ServicePackage struct{}

func (s ServicePackage) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Credentials Resource
		credentials.NewBasicAuthResource,
		credentials.NewCloudConnectorResource,
		credentials.NewWebhookSecretResource,
		credentials.NewContainerRegistryResource,
		credentials.NewKubernetesConfigResource,
		credentials.NewBasicAuthCIdPResource,
		credentials.NewCertCIdPResource,
		credentials.NewServiceKeyResource,
		credentials.NewSecretTextResource,

		// Repository Resources
		repositories.NewRepositoryResource,

		// Job Resources
		jobs.NewBuildTriggerResource,
	}
}

func (s ServicePackage) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		credentials.NewCredentialDataSource,
		credentials.NewCredentialsDataSource,
		credentials.NewCredentialUsageDataSource,
		credentials.NewJobCredentialsDataSource,

		// Repository Datasources
		repositories.NewRepositoryDataSource,
		repositories.NewRepositoriesDataSource,
		repositories.NewRepositoryJobsDataSource,
		repositories.NewRepositoryEventReceiverDataSource,
		repositories.NewRepositoryWebhookConfigDataSource,
	}
}

func (s ServicePackage) ListResources(_ context.Context) []func() list.ListResource {
	return []func() list.ListResource{
		// Repository ListResources
		repositories.NewRepositoryListResource,
	}
}
