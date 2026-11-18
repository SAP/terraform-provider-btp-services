package ctmsnodes_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/testutils"
)

func TestDataSourceCTMSNodes(t *testing.T) {
	rec := testutils.SetupVCR(t, "datasource_nodes")
	config := testDataSourceCTMSNodesConfig()

	resource.Test(
		t,
		rec.TestCase(
			resource.TestStep{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.btpservice_ctms_nodes.all", "values.#"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.btpservice_ctms_nodes.all", tfjsonpath.New("values"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"name":                 knownvalue.StringExact("Terraform_dev"),
							"description":          knownvalue.StringExact("Development Space"),
							"forward_mode":         knownvalue.StringExact("AUTO"),
							"upload_allowed":       knownvalue.Bool(true),
							"notification_enabled": knownvalue.Bool(true),
							"auto_import":          knownvalue.Bool(false),
							"virtual":              knownvalue.Bool(false),
							"controlled_by":        knownvalue.Null(),
							"job_id":               knownvalue.Null(),
							"targets": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"content_type":     knownvalue.StringExact("MTA"),
									"destination_name": knownvalue.StringExact("ctms-deployer-dev"),
									"import_options": knownvalue.ObjectExact(map[string]knownvalue.Check{
										"strategy": knownvalue.StringExact("default"),
									}),
								}),
							}),
						}),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"name":                 knownvalue.StringExact("Terraform_test"),
							"description":          knownvalue.StringExact("Test Space"),
							"forward_mode":         knownvalue.StringExact("MANUAL"),
							"upload_allowed":       knownvalue.Bool(true),
							"notification_enabled": knownvalue.Bool(true),
							"auto_import":          knownvalue.Bool(false),
							"virtual":              knownvalue.Bool(false),
							"controlled_by":        knownvalue.Null(),
							"job_id":               knownvalue.Null(),
							"targets": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"content_type":     knownvalue.StringExact("MTA"),
									"destination_name": knownvalue.StringExact("ctms-deployer-test"),
									"import_options": knownvalue.ObjectExact(map[string]knownvalue.Check{
										"strategy": knownvalue.StringExact("default"),
									}),
								}),
							}),
						}),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"name":                 knownvalue.StringExact("Terraform_prod"),
							"description":          knownvalue.StringExact("Production Space"),
							"forward_mode":         knownvalue.StringExact("AUTO"),
							"upload_allowed":       knownvalue.Bool(false),
							"notification_enabled": knownvalue.Bool(true),
							"auto_import":          knownvalue.Bool(false),
							"virtual":              knownvalue.Bool(false),
							"controlled_by":        knownvalue.Null(),
							"job_id":               knownvalue.Null(),
							"targets": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"content_type":     knownvalue.StringExact("MTA"),
									"destination_name": knownvalue.StringExact("ctms-deployer-prod"),
									"import_options": knownvalue.ObjectExact(map[string]knownvalue.Check{
										"strategy": knownvalue.StringExact("default"),
									}),
								}),
							}),
						}),
					})),
				},
			},
			resource.TestStep{
				Config:   config,
				PlanOnly: true,
			},
		),
	)
}

func testDataSourceCTMSNodesConfig() string {
	return `
data "btpservice_ctms_nodes" "all" {}
`
}
