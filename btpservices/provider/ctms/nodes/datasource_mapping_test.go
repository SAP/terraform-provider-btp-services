package ctmsnodes_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestNodeDataSourcesImportStatusMapping(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fields   string
		disabled knownvalue.Check
		reason   knownvalue.Check
	}{
		{"disabled", `,"importDisabled":true,"importDisabledReason":"Maintenance"`, knownvalue.Bool(true), knownvalue.StringExact("Maintenance")},
		{"enabled", `,"importDisabled":false,"importDisabledReason":""`, knownvalue.Bool(false), knownvalue.StringExact("")},
		{"omitted", "", knownvalue.Null(), knownvalue.Null()},
		{"null", `,"importDisabled":null,"importDisabledReason":null`, knownvalue.Null(), knownvalue.Null()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected := map[string]knownvalue.Check{
				"id":                     knownvalue.StringExact("183"),
				"name":                   knownvalue.StringExact("Terraform_dev"),
				"description":            knownvalue.StringExact("Development Space"),
				"upload_allowed":         knownvalue.Bool(true),
				"forward_mode":           knownvalue.StringExact("MANUAL"),
				"notification_enabled":   knownvalue.Bool(false),
				"virtual":                knownvalue.Bool(false),
				"auto_import":            knownvalue.Bool(true),
				"controlled_by":          knownvalue.StringExact("CHARM"),
				"job_id":                 knownvalue.Int64Exact(5434),
				"import_disabled":        tc.disabled,
				"import_disabled_reason": tc.reason,
				"targets": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"content_type":     knownvalue.StringExact("MTA"),
						"destination_name": knownvalue.StringExact("CF_TARGET_DESTINATION"),
						"import_options":   knownvalue.ObjectExact(map[string]knownvalue.Check{"strategy": knownvalue.StringExact("blue-green")}),
					}),
				}),
			}
			var checks []statecheck.StateCheck
			for _, address := range []string{"data.btpservice_ctms_node.id", "data.btpservice_ctms_node.name"} {
				for attribute, value := range expected {
					checks = append(checks, statecheck.ExpectKnownValue(address, tfjsonpath.New(attribute), value))
				}
			}
			checks = append(checks, statecheck.ExpectKnownValue("data.btpservice_ctms_nodes.all", tfjsonpath.New("values"),
				knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(expected)})))
			body := `{"id":183,"name":"Terraform_dev","description":"Development Space","uploadAllowed":true,"forwardMode":"MANUAL","notificationEnabled":false,"virtual":false,"autoImport":true,"controlledBy":"CHARM","jobId":5434,"targets":[{"contentType":"MTA","destinationName":"CF_TARGET_DESTINATION","importOptions":{"strategy":"blue-green"}}]` + tc.fields + `}`
			resource.Test(t, nodeTestCaseWithTransport(t, nodeDataSourceTransport{t: t, node: body}, resource.TestStep{
				Config:            testDataSourceCTMSNodeConfig() + testDataSourceCTMSNodesConfig(),
				ConfigStateChecks: checks,
			}))
		})
	}
}

type nodeDataSourceTransport struct {
	t    *testing.T
	node string
}

func (transport nodeDataSourceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body string
	switch {
	case req.Method == http.MethodPost && req.URL.String() == "https://auth.example/oauth/token":
		body = `{"access_token":"test-token","expires_in":3600}`
	case req.Method == http.MethodGet && req.URL.String() == "https://ctms.example/v2/nodes/183":
		body = transport.node
	case req.Method == http.MethodGet && req.URL.String() == "https://ctms.example/v2/nodes":
		body = `{"nodes":[` + transport.node + `]}`
	default:
		transport.t.Errorf("unexpected HTTP request during local data source test: %s %s", req.Method, req.URL)
		return nil, fmt.Errorf("unexpected local data source request")
	}
	return &http.Response{
		StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}},
		Body: io.NopCloser(strings.NewReader(body)), Request: req,
	}, nil
}
