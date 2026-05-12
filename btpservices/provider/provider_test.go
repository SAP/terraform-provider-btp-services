// btpservices/provider/provider_test.go

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tftest "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

func providerFactoryDev() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"btpservice": providerserver.NewProtocol6WithError(
			New()(),
		),
	}
}

func providerFactoryWithClients() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"btpservice": providerserver.NewProtocol6WithError(
			NewWithClients(&shared.ProviderClients{}),
		),
	}
}

func TestProvider_Schema(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactoryDev(),
		Steps: []tftest.TestStep{
			{
				Config: `provider "bptservice" {}`,
			},
		},
	})
}

func TestProvider_CicdBlock(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactoryDev(),
		Steps: []tftest.TestStep{
			{
				Config: `
provider "btpservice" {
  cicd {
    endpoint      = "https://cicd-service.cfapps.eu12.hana.ondemand.com"
    token_url     = "https://example.authentication.eu12.hana.ondemand.com/oauth/token"
    client_id     = "test-client-id"
    client_secret = "test-client-secret"
    timeout       = 60
  }
}`,
			},
		},
	})
}

func TestProvider_CicdBlock_CustomTimeout(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactoryDev(),
		Steps: []tftest.TestStep{
			{
				Config: `
provider "btpservice" {
  cicd {
    endpoint      = "https://cicd-service.cfapps.eu12.hana.ondemand.com"
    token_url     = "https://example.authentication.eu12.hana.ondemand.com/oauth/token"
    client_id     = "test-client-id"
    client_secret = "test-client-secret"
    timeout       = 120
  }
}`,
			},
		},
	})
}

func TestProvider_EnvVarFallback(t *testing.T) {
	t.Setenv("BTP_CICD_ENDPOINT", "https://cicd-service.cfapps.eu12.hana.ondemand.com")
	t.Setenv("BTP_CICD_TOKEN_URL", "https://example.authentication.eu12.hana.ondemand.com/oauth/token")
	t.Setenv("BTP_CICD_CLIENT_ID", "test-client-id")
	t.Setenv("BTP_CICD_CLIENT_SECRET", "test-client-secret")

	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactoryDev(),
		Steps: []tftest.TestStep{
			{
				Config: `
provider "btpservice" {
  cicd {}
}`,
			},
		},
	})
}

func TestProvider_NewWithClients(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactoryWithClients(),
		Steps: []tftest.TestStep{
			{
				Config: `provider "btpservice" {}`,
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Resources() / DataSources() registration tests
// ---------------------------------------------------------------------------

// expectedResourceTypes lists every resource type name the provider must expose.
var expectedResourceTypes = []string{
	"btpservice_cicd_credential_basic_auth",
	"btpservice_cicd_credential_cloud_connector",
	"btpservice_cicd_credential_webhook_secret",
	"btpservice_cicd_credential_container_registry",
	"btpservice_cicd_credential_kubernetes_config",
}

// expectedDataSourceTypes lists every data source type name the provider must expose.
var expectedDataSourceTypes = []string{
	"btpservice_cicd_credential",
	"btpservice_cicd_credentials",
}

func TestProvider_Resources_TypeNames(t *testing.T) {
	p := New()()
	ctx := context.Background()

	got := make(map[string]bool)
	for _, factory := range p.Resources(ctx) {
		r := factory()
		var meta resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "btpservice"}, &meta)
		got[meta.TypeName] = true
	}

	for _, name := range expectedResourceTypes {
		if !got[name] {
			t.Errorf("missing resource type %q", name)
		}
	}
}

func TestProvider_DataSources_TypeNames(t *testing.T) {
	p := New()()
	ctx := context.Background()

	got := make(map[string]bool)
	for _, factory := range p.DataSources(ctx) {
		ds := factory()
		var meta datasource.MetadataResponse
		ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "btpservice"}, &meta)
		got[meta.TypeName] = true
	}

	for _, name := range expectedDataSourceTypes {
		if !got[name] {
			t.Errorf("missing data source type %q", name)
		}
	}
}
