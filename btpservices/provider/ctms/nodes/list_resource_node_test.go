package ctmsnodes_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestListResourceCTMSNode(t *testing.T) {
	rec := testutils.SetupVCR(t, "list_resource_node")
	identity := map[string]knownvalue.Check{
		"id":       knownvalue.StringExact("183"),
		"endpoint": knownvalue.StringExact(strings.TrimRight(rec.Values["endpoint"], "/")),
	}

	resource.Test(t, rec.TestCase(
		resource.TestStep{
			Query:  true,
			Config: testListResourceCTMSNodeConfig(false),
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_node.nodes", 1),
				querycheck.ExpectIdentity("btpservice_ctms_node.nodes", identity),
			},
		},
		resource.TestStep{
			Query:          true,
			GenerateConfig: true,
			Config:         testListResourceCTMSNodeConfig(true),
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_node.nodes", 1),
				querycheck.ExpectIdentity("btpservice_ctms_node.nodes", identity),
				querycheck.ExpectResourceKnownValues(
					"btpservice_ctms_node.nodes",
					queryfilter.ByResourceIdentity(identity),
					[]querycheck.KnownValueCheck{
						{Path: tfjsonpath.New("id"), KnownValue: knownvalue.StringExact("183")},
						{Path: tfjsonpath.New("name"), KnownValue: knownvalue.StringExact("Terraform_dev")},
						{Path: tfjsonpath.New("description"), KnownValue: knownvalue.StringExact("Development Space")},
						{Path: tfjsonpath.New("forward_mode"), KnownValue: knownvalue.StringExact("AUTO")},
						{Path: tfjsonpath.New("upload_allowed"), KnownValue: knownvalue.Bool(true)},
						{Path: tfjsonpath.New("notification_enabled"), KnownValue: knownvalue.Bool(true)},
						{Path: tfjsonpath.New("auto_import"), KnownValue: knownvalue.Bool(false)},
						{Path: tfjsonpath.New("virtual"), KnownValue: knownvalue.Bool(false)},
						{Path: tfjsonpath.New("controlled_by"), KnownValue: knownvalue.Null()},
						{Path: tfjsonpath.New("job_id"), KnownValue: knownvalue.Null()},
						{Path: tfjsonpath.New("targets"), KnownValue: knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"content_type":     knownvalue.StringExact("MTA"),
								"destination_name": knownvalue.StringExact("ctms-deployer-dev"),
								"import_options": knownvalue.ObjectExact(map[string]knownvalue.Check{
									"strategy": knownvalue.StringExact("default"),
								}),
							}),
						})},
					},
				),
			},
		},
	))
}

func testListResourceCTMSNodeConfig(includeResource bool) string {
	return fmt.Sprintf(`
provider "btpservice" {}

list "btpservice_ctms_node" "nodes" {
  provider         = btpservice
  include_resource = %t
  config { name = "Terraform_dev" }
}
`, includeResource)
}

func TestListResourceCTMSNodeQueryConfig(t *testing.T) {
	for _, includeResource := range []bool{false, true} {
		t.Run(fmt.Sprintf("include_resource=%t", includeResource), func(t *testing.T) {
			checks := []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_node.nodes", 1),
				querycheck.ExpectIdentity("btpservice_ctms_node.nodes", map[string]knownvalue.Check{
					"id":       knownvalue.StringExact("183"),
					"endpoint": knownvalue.StringExact("https://ctms.example"),
				}),
			}
			if includeResource {
				checks = append(checks, querycheck.ExpectResourceKnownValues(
					"btpservice_ctms_node.nodes", nil, []querycheck.KnownValueCheck{
						{Path: tfjsonpath.New("id"), KnownValue: knownvalue.StringExact("183")},
						{Path: tfjsonpath.New("name"), KnownValue: knownvalue.StringExact("Terraform_dev")},
					},
				))
			}
			resource.Test(t, nodeTestCaseWithTransport(t, nodeListQueryTransport{t: t}, resource.TestStep{
				Query: true, Config: testListResourceCTMSNodeConfig(includeResource), QueryResultChecks: checks,
			}))
		})
	}
}

type nodeListQueryTransport struct{ t *testing.T }

func (transport nodeListQueryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body string
	switch {
	case req.Method == http.MethodPost && req.URL.String() == "https://auth.example/oauth/token":
		body = `{"access_token":"test-token","expires_in":3600}`
	case req.Method == http.MethodGet && req.URL.String() == "https://ctms.example/v2/nodes":
		body = `{"nodes":[{"id":183,"name":"Terraform_dev","description":"Development Space","uploadAllowed":true,"forwardMode":"AUTO","notificationEnabled":true,"autoImport":false,"virtual":false,"targets":[]}]}`
	default:
		transport.t.Errorf("unexpected HTTP request during local query test: %s %s", req.Method, req.URL)
		return nil, fmt.Errorf("unexpected local query request")
	}
	return &http.Response{
		StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}},
		Body: io.NopCloser(strings.NewReader(body)), Request: req,
	}, nil
}
