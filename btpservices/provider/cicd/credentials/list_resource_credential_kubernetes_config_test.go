// btpservices/provider/cicd/credentials/list_resource_credential_kubernetes_config_test.go

package cicdcredentials_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	res "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/cicdtest"
	cicdcredentials "github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/cicd/credentials"
	"github.com/SAP/terraform-provider-sap-btp-services/btpservices/provider/testutil"
)

func TestListResourceCicdCredentialKubernetesConfig(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		rec, creds := cicdtest.SetupVCR(t, "../fixtures/list_resource_credential_kubernetes_config")
		defer testutil.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: cicdtest.GetTestProviders(creds, rec),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query: true,
					Config: cicdtest.HCLProviderBlock(creds) + `
list "btpservice_cicd_credential_kubernetes_config" "test" {
  provider = "btpservice"
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_credential_kubernetes_config.test", 1),
					},
				},
				{
					Query: true,
					Config: cicdtest.HCLProviderBlock(creds) + `
list "btpservice_cicd_credential_kubernetes_config" "test" {
  provider         = "btpservice"
  include_resource = true
}
`,
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLengthAtLeast("btpservice_cicd_credential_kubernetes_config.test", 1),
						// Verify full resource data is populated
						querycheck.ExpectResourceKnownValues(
							"btpservice_cicd_credential_kubernetes_config.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"id": knownvalue.StringExact("f9425a2e-aba6-4425-b983-ec7c87d0867b"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("my-kubernetes-config-test-list-resource"),
								},
								{
									Path:       tfjsonpath.New("description"),
									KnownValue: knownvalue.NotNull(),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		t.Parallel()

		r := cicdcredentials.NewKubernetesConfigListResource().(list.ListResourceWithConfigure)
		resp := &res.ConfigureResponse{}
		req := res.ConfigureRequest{
			ProviderData: struct{}{},
		}

		r.Configure(context.Background(), req, resp)

		if !resp.Diagnostics.HasError() {
			t.Error("Expected error for invalid provider data type")
		}
	})
}
