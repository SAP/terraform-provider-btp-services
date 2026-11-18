package ctmsroutes_test

import "github.com/hashicorp/terraform-plugin-testing/knownvalue"

// Expected route fields from TEST_DO_NOT_COMMIT/02_SETUP/terraform.tfstate.
func testDevelopmentToQualityRouteValues() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"id":             knownvalue.StringExact("192"),
		"name":           knownvalue.StringExact("Terraform_DEV_TO_TEST"),
		"description":    knownvalue.StringExact("Route from Development to Test Space"),
		"source_node_id": knownvalue.StringExact("183"),
		"target_node_id": knownvalue.StringExact("184"),
	}
}

func testTestToProductionRouteValues() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"id":             knownvalue.StringExact("204"),
		"name":           knownvalue.StringExact("Terraform_TEST_TO_PROD"),
		"description":    knownvalue.StringExact("Route from Test to Production Space"),
		"source_node_id": knownvalue.StringExact("184"),
		"target_node_id": knownvalue.StringExact("201"),
	}
}
