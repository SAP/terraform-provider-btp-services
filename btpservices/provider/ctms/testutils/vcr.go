package testutils

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	btpservicesprovider "github.com/SAP/terraform-provider-btp-services/btpservices/provider"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
	ctmsclient "github.com/SAP/terraform-provider-btp-services/internal/ctms/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

// Recording supplies CTMS clients and test cases backed by the shared recorder.
type Recording struct {
	rec    *recorder.Recorder
	Values tfutils.TestCredentials
}

// SetupVCR configures shared recording with CTMS credentials and redaction rules.
// Missing cassettes are skipped unless TEST_FORCE_REC explicitly enables recording.
func SetupVCR(t *testing.T, name string, extra ...string) *Recording {
	t.Helper()

	redacted := tfutils.TestCredentials{
		"endpoint":      "https://redacted-api.example",
		"token_url":     "https://redacted-auth.example/oauth/token",
		"client_id":     "redacted-client-id",
		"client_secret": "redacted-client-secret",
		"subaccount_id": "redacted-subaccount-id",
	}

	for _, key := range extra {
		redacted[key] = "redacted-" + strings.ReplaceAll(key, "_", "-")
	}

	liveEnvVars := make(map[string]string, len(redacted))
	for key := range redacted {
		liveEnvVars[key] = "BTP_CTMS_" + strings.ToUpper(key)
	}

	rec, values := tfutils.SetupVCRWithOptions(t, "../fixtures/"+name, liveEnvVars, redacted, tfutils.VCROptions{
		RequireExplicitRecording: true,
		MatchRequestBody:         true,
		MatchHeaders:             []string{"Tms-Named-User"},
		Redaction: &tfutils.VCRRedaction{
			URLKeys:      []string{"endpoint", "token_url"},
			NormalizeURL: normalizeCTMSURL,
			RedactString: redactCTMSString,
		},
	})

	t.Cleanup(func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("stop CTMS recorder: %v", err)
		}
	})

	return &Recording{
		rec:    rec,
		Values: values,
	}
}

func normalizeCTMSURL(parsed *url.URL, values tfutils.TestCredentials) {
	query := parsed.Query()
	for key, input := range map[string]string{
		"transportRequestId": "test_transport_request_id",
		"nodeId":             "test_node_id",
	} {
		if value := values[input]; value != "" && query.Get(key) == value {
			query.Set(key, "redacted-"+strings.ReplaceAll(input, "_", "-"))
		}
	}

	parsed.RawQuery = query.Encode()
	if destination := values["test_destination"]; destination != "" && parsed.Path == "/v2/destinations/"+destination {
		parsed.Path = "/v2/destinations/redacted-test-destination"
		parsed.RawPath = ""
	}
}

func redactCTMSString(key, value string, values tfutils.TestCredentials) string {
	lower := strings.ToLower(key)
	if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") ||
		lower == "client_id" || lower == "clientid" || lower == "subaccount_id" {
		return "redacted"
	}

	if (key == "destinationName" || key == "name") && values["test_destination"] != "" && value == values["test_destination"] {
		return "redacted-test-destination"
	}

	if key == "url" || key == "tmsUiUrl" || key == "tmsUiLogUrl" {
		return "https://redacted-ui.example/" + key
	}

	return value
}

// Client creates a fresh facade using the cassette transport.
func (r *Recording) Client(namedUser string) *ctmsclient.CtmsClientFacade {
	httpClient := r.rec.GetDefaultClient()
	httpClient.Timeout = 60 * time.Second
	return ctmsclient.NewCtmsClientFacadeWithHTTP(
		ctmsclient.CtmsClientConfig{
			Endpoint:     r.Values["endpoint"],
			TokenURL:     r.Values["token_url"],
			ClientID:     r.Values["client_id"],
			ClientSecret: r.Values["client_secret"],
			NamedUser:    namedUser,
		},
		httpClient,
	)
}

func (r *Recording) factories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"btpservice": func() (tfprotov6.ProviderServer, error) {
			// A fresh facade on each factory call also exercises imports without a
			// preceding create on that client.
			return providerserver.NewProtocol6WithError(btpservicesprovider.NewWithClients(&shared.ProviderClients{
				Ctms: r.Client("terraform-provider-ctms-test"),
				CtmsScope: shared.CtmsScope{
					Endpoint:     strings.TrimRight(r.Values["endpoint"], "/"),
					SubaccountID: r.Values["subaccount_id"],
				},
			}))()
		},
	}
}

// TestCase configures a Terraform acceptance case with the recorded client.
func (r *Recording) TestCase(steps ...resource.TestStep) resource.TestCase {
	return resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: r.factories(),
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		Steps:                    steps,
	}
}
