// btpservices/provider/services/cicd/cicdtest/testhelper.go

package cicdtest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	btpservicesprovider "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/testutil"
	cicdclient "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

// Redacted holds the placeholder values written into cassettes.
// These are used on replay — no live credentials are needed.
var Redacted = testutil.TestCredentials{
	"endpoint":      "https://cicd-service.cfapps.eu12.hana.ondemand.com",
	"token_url":     "https://test-ci-cd-l646hcu6.authentication.eu12.hana.ondemand.com/oauth/token",
	"client_id":     "redacted-client-id",
	"client_secret": "redacted-client-secret",
}

// liveEnvVars maps each credential key to the environment variable that
// holds the real value when recording a cassette against the live API.
var liveEnvVars = map[string]string{
	"endpoint":      "SAPBTP_CICD_ENDPOINT",
	"token_url":     "SAPBTP_CICD_TOKEN_URL",
	"client_id":     "SAPBTP_CICD_CLIENT_ID",
	"client_secret": "SAPBTP_CICD_CLIENT_SECRET",
}

// SetupVCR wraps testutil.SetupVCR with CI/CD-specific env vars and
// redacted placeholders. All CI/CD tests call this instead of testutil.SetupVCR directly.
func SetupVCR(t *testing.T, cassetteName string) (*recorder.Recorder, testutil.TestCredentials) {
	t.Helper()
	return testutil.SetupVCR(t, cassetteName, liveEnvVars, Redacted)
}

// GetTestProviders injects a VCR-wrapped http.Client into the provider,
// bypassing the real OAuth2 token flow during tests.
// Pass nil for rec when no HTTP calls are expected (e.g. required-field error tests).
func GetTestProviders(creds testutil.TestCredentials, rec *recorder.Recorder) map[string]func() (tfprotov6.ProviderServer, error) {
	var httpClient *http.Client
	if rec != nil {
		httpClient = rec.GetDefaultClient()
	} else {
		httpClient = http.DefaultClient
	}
	return map[string]func() (tfprotov6.ProviderServer, error){
		"btpservice": providerserver.NewProtocol6WithError(
			btpservicesprovider.NewWithClients(&shared.ProviderClients{
				Cicd: cicdclient.NewCicdClientFacadeWithHTTP(cicdclient.CicdClientConfig{
					Endpoint:     creds["endpoint"],
					TokenURL:     creds["token_url"],
					ClientID:     creds["client_id"],
					ClientSecret: creds["client_secret"],
				}, httpClient),
			}),
		),
	}
}

// HCLProviderBlock returns the provider HCL block for test configs.
func HCLProviderBlock(creds testutil.TestCredentials) string {
	return fmt.Sprintf(`
provider "btpservice" {
  cicd {
    endpoint      = "%s"
    token_url     = "%s"
    client_id     = "%s"
    client_secret = "%s"
  }
}
`, creds["endpoint"], creds["token_url"], creds["client_id"], creds["client_secret"])
}
