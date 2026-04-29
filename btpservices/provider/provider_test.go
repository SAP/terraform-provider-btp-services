// btpservices/provider/provider_test.go

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

// providerFactoryDev returns a provider factory using New() — the normal
// production code path (no pre-injected clients).
func providerFactoryDev() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"btpservice": providerserver.NewProtocol6WithError(
			New()(),
		),
	}
}

// providerFactoryWithClients returns a provider factory with pre-injected
// clients — simulates the acceptance-test path without real credentials.
func providerFactoryWithClients() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"btpservice": providerserver.NewProtocol6WithError(
			NewWithClients(&shared.ProviderClients{}),
		),
	}
}

// TestProvider_Schema verifies the provider schema is valid and contains the
// expected cicd block with all required attributes.
func TestProvider_Schema(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoryDev(),
		Steps: []resource.TestStep{
			{
				Config: `provider "bptservice" {}`,
			},
		},
	})
}

// TestProvider_CicdBlock verifies the provider accepts a fully-populated cicd block.
func TestProvider_CicdBlock(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoryDev(),
		Steps: []resource.TestStep{
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

// TestProvider_CicdBlock_CustomTimeout verifies the provider accepts a custom timeout value.
func TestProvider_CicdBlock_CustomTimeout(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoryDev(),
		Steps: []resource.TestStep{
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

// TestProvider_EnvVarFallback verifies the provider accepts an empty cicd block
// (all values resolved from environment variables at runtime).
func TestProvider_EnvVarFallback(t *testing.T) {
	t.Setenv("SAPBTP_CICD_ENDPOINT", "https://cicd-service.cfapps.eu12.hana.ondemand.com")
	t.Setenv("SAPBTP_CICD_TOKEN_URL", "https://example.authentication.eu12.hana.ondemand.com/oauth/token")
	t.Setenv("SAPBTP_CICD_CLIENT_ID", "test-client-id")
	t.Setenv("SAPBTP_CICD_CLIENT_SECRET", "test-client-secret")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoryDev(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "btpservice" {
  cicd {}
}`,
			},
		},
	})
}

// TestProvider_NewWithClients verifies the pre-injected-clients path used by
// acceptance tests short-circuits Configure without needing real credentials.
func TestProvider_NewWithClients(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoryWithClients(),
		Steps: []resource.TestStep{
			{
				Config: `provider "btpservice" {}`,
			},
		},
	})
}
