// btpservices/provider/cicd/settings/resource_allowed_spaces_test.go

package cicdsettings_test

import (
	"context"
	"regexp"
	"testing"

	res "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	cicdsettings "github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/settings"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestResourceCicdAllowedSpaces(t *testing.T) {
	t.Parallel()

	t.Run("happy path - create and update", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_allowed_spaces")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_allowed_spaces" "test" {
  allowed_spaces = [
    {
      space_guid = "f9a880bc-20c3-4f71-a51f-882ccdd369c6"
      comment    = "Team Beta space"
    },
  ]
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						// resource.TestCheckResourceAttr("btpservice_cicd_allowed_spaces.test", "id", "allowed_spaces"),
						resource.TestCheckResourceAttrSet("btpservice_cicd_allowed_spaces.test", "allowed_spaces.#"),
					),
				},
				{
					// Step 2: Add a second space
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_allowed_spaces" "test" {
  allowed_spaces = [
    {
      space_guid = "a2bcf2b8-6eda-5b8a-0b7c-8512bb82060f"
      comment    = "Team Alpha space"
    },
	{
      space_guid = "f9a880bc-20c3-4f71-a51f-882ccdd369c6"
      comment    = "space id for integration-test-cicd-service"
    },
  ]
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_allowed_spaces.test", "allowed_spaces.#"),
					),
				},
				{
					// Step 3: Import — any string accepted as the singleton has no real ID.
					// ImportStateVerify is omitted because the resource has no id attribute.
					ResourceName: "btpservice_cicd_allowed_spaces.test",
					ImportState:  true,
				},
			},
		})
	})

	t.Run("error - missing allowed_spaces", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_allowed_spaces" "test" {
}
`,
					ExpectError: regexp.MustCompile(`The argument "allowed_spaces" is required`),
				},
			},
		})
	})

	t.Run("error - missing space_guid", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_allowed_spaces" "test" {
  allowed_spaces = [
    {
      comment = "missing guid"
    },
  ]
}
`,
					ExpectError: regexp.MustCompile(`"space_guid" is required`),
				},
			},
		})
	})

	t.Run("error - missing comment", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_allowed_spaces" "test" {
  allowed_spaces = [
    {
      space_guid = "a2bcf2b8-6eda-5b8a-0b7c-8512bb82060f"
    },
  ]
}
`,
					ExpectError: regexp.MustCompile(`"comment" is required`),
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		t.Parallel()

		r := cicdsettings.NewAllowedSpacesResource().(res.ResourceWithConfigure)
		resp := &res.ConfigureResponse{}
		req := res.ConfigureRequest{ProviderData: struct{}{}}

		r.Configure(context.Background(), req, resp)

		if !resp.Diagnostics.HasError() {
			t.Error("expected error for invalid provider data type")
		}
	})
}
