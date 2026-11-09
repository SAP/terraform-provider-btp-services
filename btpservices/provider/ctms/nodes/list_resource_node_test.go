package ctmsnodes_test

import (
	"strconv"
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
	id := rec.CreateVirtualNode(t, "tf-ctms-query-node")
	identity := map[string]knownvalue.Check{
		"id":            knownvalue.StringExact(strconv.FormatInt(id, 10)),
		"endpoint":      knownvalue.StringExact(strings.TrimRight(rec.Values["endpoint"], "/")),
		"subaccount_id": knownvalue.StringExact(rec.Values["subaccount_id"]),
	}

	resource.Test(t, rec.TestCase(
		resource.TestStep{
			Query: true,
			Config: `
list "btpservice_ctms_node" "nodes" {
  provider = btpservice
  config { name = "tf-ctms-query-node" }
}
`,
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_node.nodes", 1),
				querycheck.ExpectIdentity("btpservice_ctms_node.nodes", identity),
			},
		},
		resource.TestStep{
			Query:          true,
			GenerateConfig: true,
			Config: `
list "btpservice_ctms_node" "nodes" {
  provider         = btpservice
  include_resource = true
  config { name = "tf-ctms-query-node" }
}
`,
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_node.nodes", 1),
				querycheck.ExpectResourceKnownValues(
					"btpservice_ctms_node.nodes",
					queryfilter.ByResourceIdentity(identity),
					[]querycheck.KnownValueCheck{
						{Path: tfjsonpath.New("name"), KnownValue: knownvalue.StringExact("tf-ctms-query-node")},
					},
				),
			},
		},
	))
}
