package ctmsnodes_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	btpservicesprovider "github.com/SAP/terraform-provider-btp-services/btpservices/provider"
	ctmsclient "github.com/SAP/terraform-provider-btp-services/internal/ctms/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

// Schema checks use the real provider without a recorder or network access.
func nodeValidationTestCase(t *testing.T, steps ...resource.TestStep) resource.TestCase {
	t.Helper()
	return nodeTestCaseWithTransport(t, nodeValidationTransport{t: t}, steps...)
}

func nodeTestCaseWithTransport(t *testing.T, transport http.RoundTripper, steps ...resource.TestStep) resource.TestCase {
	t.Helper()
	return resource.TestCase{
		IsUnitTest:             true,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"btpservice": providerserver.NewProtocol6WithError(btpservicesprovider.NewWithClients(&shared.ProviderClients{
				Ctms: ctmsclient.NewCtmsClientFacadeWithHTTP(ctmsclient.CtmsClientConfig{
					Endpoint: "https://ctms.example", TokenURL: "https://auth.example/oauth/token",
					ClientID: "test-client", ClientSecret: "test-secret",
				}, &http.Client{Transport: transport}),
				CtmsScope: shared.CtmsScope{Endpoint: "https://ctms.example"},
			})),
		},
		Steps: steps,
	}
}

type nodeValidationTransport struct{ t *testing.T }

func (transport nodeValidationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport.t.Errorf("unexpected HTTP request during schema validation: %s %s", req.Method, req.URL.Path)
	return nil, fmt.Errorf("schema validation must not access CTMS")
}
