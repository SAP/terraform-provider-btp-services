package ctmsroutes_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

func TestListResourceCTMSRoute(t *testing.T) {
	rec := testutils.SetupVCR(t, "list_resource_route")
	sourceID := rec.CreateVirtualNode(t, "tf-ctms-query-route-source")
	targetID := rec.CreateVirtualNode(t, "tf-ctms-query-route-target")
	client := rec.Client("terraform-provider-ctms-test")
	route, err := client.Routes.Create(context.Background(), ctmsmodels.RouteCreateRequest{
		Name:         "tf-ctms-query-route",
		SourceNodeId: sourceID,
		TargetNodeId: targetID,
	})
	if err != nil {
		t.Fatal(err)
	}

	// LIFO cleanup removes the route before either node and stops the recorder last.
	t.Cleanup(func() {
		err := client.Routes.Delete(context.Background(), *route.Id)
		if err != nil && !ctmsmodels.IsNotFound(err) {
			t.Errorf("cleanup CTMS route: %v", err)
		}
	})
	identity := map[string]knownvalue.Check{
		"id":            knownvalue.StringExact(strconv.FormatInt(*route.Id, 10)),
		"endpoint":      knownvalue.StringExact(strings.TrimRight(rec.Values["endpoint"], "/")),
		"subaccount_id": knownvalue.StringExact(rec.Values["subaccount_id"]),
	}

	resource.Test(t, rec.TestCase(
		resource.TestStep{
			Query:          true,
			GenerateConfig: true,
			Config: fmt.Sprintf(`
list "btpservice_ctms_route" "routes" {
  provider         = btpservice
  include_resource = true
  config {
    name           = "tf-ctms-query-route"
    source_node_id = %q
    target_node_id = %q
  }
}
`, strconv.FormatInt(sourceID, 10), strconv.FormatInt(targetID, 10)),
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_route.routes", 1),
				querycheck.ExpectIdentity("btpservice_ctms_route.routes", identity),
			},
		},
		resource.TestStep{
			Query: true,
			Config: `
list "btpservice_ctms_route" "routes" {
  provider = btpservice
  config { name = "tf-ctms-deliberately-missing-route" }
}
`,
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLength("btpservice_ctms_route.routes", 0),
			},
		},
	))
}
