// btpservices/provider/cicd/jobs/resource_job_test.go

package cicdjobs_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/cicd/utils"
	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/tfutils"
)

func TestResourceCicdJob(t *testing.T) {
	t.Parallel()

	t.Run("happy path - cf-env pipeline", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_job_cf_env")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-cf-env"
  description          = "CF env test job"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  active               = true
  build_retention_days = 28
  max_builds_to_keep   = 10

  pipeline_parameters = <<-YAML
    configurationSource: job_parameter
    cfEnvConfiguration:
      stages:
        build:
          buildTool: mta
          buildToolVersion: MBTJ21N24
  YAML
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "name", "tf-test-cf-env"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "description", "CF env test job"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "pipeline", "cf-env"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "pipeline_version", "3.0"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "active", "true"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "branch", "main"),
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "pipeline_parameters"),
					),
				},
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-cf-env"
  description          = "CF env test job - updated"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  active               = false
  build_retention_days = 14
  max_builds_to_keep   = 5

  pipeline_parameters = <<-YAML
    configurationSource: job_parameter
    cfEnvConfiguration:
      stages:
        build:
          buildTool: mta
          buildToolVersion: MBTJ21N24
  YAML
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "description", "CF env test job - updated"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "active", "false"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "build_retention_days", "14"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "max_builds_to_keep", "5"),
					),
				},
				{
					ResourceName:            "btpservice_cicd_job.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"pipeline_parameters"},
				},
			},
		})
	})

	t.Run("happy path - kyma-cnb pipeline", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_job_kyma_cnb")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-kyma-cnb"
  description          = "Kyma CNB test job"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "kyma-cnb"
  pipeline_version     = "1.0"
  active               = true
  build_retention_days = 7
  max_builds_to_keep   = 50

  pipeline_parameters = <<-YAML
    configurationSource: job_parameter
    kymaCnbConfiguration:
      common:
        chartPath: helm/chart
        images:
          - name: app-image
            path: .
      stages:
        build:
          npm:
            toolVersion: N22
          cnb:
            containerRegistry:
              url: https://docker.io/myuser
              credential: d794d687-3053-4cba-a942-88e6b13ef035
  YAML
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "name", "tf-test-kyma-cnb"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "pipeline", "kyma-cnb"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "pipeline_version", "1.0"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "active", "true"),
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "pipeline_parameters"),
					),
				},
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-kyma-cnb"
  description          = "Kyma CNB test job - updated"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "kyma-cnb"
  pipeline_version     = "1.0"
  active               = false
  build_retention_days = 14
  max_builds_to_keep   = 20

  pipeline_parameters = <<-YAML
    configurationSource: job_parameter
    kymaCnbConfiguration:
      common:
        chartPath: helm/chart
        images:
          - name: app-image
            path: .
      stages:
        build:
          npm:
            toolVersion: N22
          cnb:
            containerRegistry:
              url: https://docker.io/myuser
              credential: d794d687-3053-4cba-a942-88e6b13ef035
  YAML
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "description", "Kyma CNB test job - updated"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "active", "false"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "build_retention_days", "14"),
					),
				},
				{
					ResourceName:            "btpservice_cicd_job.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"pipeline_parameters"},
				},
			},
		})
	})

	t.Run("happy path - cpi pipeline", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_job_cpi")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-cpi"
  description          = "CPI test job"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "cpi"
  pipeline_version     = "2.0"
  active               = true
  build_retention_days = 14
  max_builds_to_keep   = 20

  pipeline_parameters = <<-YAML
    configurationSource: job_parameter
    cpiConfiguration:
      common:
        integrationFlowId: MyIntegrationFlow
        apiPlanServiceKey: b9a0ba8a-8933-446f-b32e-7ff64ca2e5ce
      stages:
        upload:
          integrationFlowName: My Integration Flow
          packageId: com.example.mypackage
          active: true
  YAML
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "name", "tf-test-cpi"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "pipeline", "cpi"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "pipeline_version", "2.0"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "active", "true"),
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "pipeline_parameters"),
					),
				},
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-cpi"
  description          = "CPI test job - updated"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "cpi"
  pipeline_version     = "2.0"
  active               = false
  build_retention_days = 7
  max_builds_to_keep   = 10

  pipeline_parameters = <<-YAML
    configurationSource: job_parameter
    cpiConfiguration:
      common:
        integrationFlowId: MyIntegrationFlow
        apiPlanServiceKey: b9a0ba8a-8933-446f-b32e-7ff64ca2e5ce
      stages:
        upload:
          integrationFlowName: My Integration Flow
          packageId: com.example.mypackage
          active: true
  YAML
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "description", "CPI test job - updated"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "active", "false"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "build_retention_days", "7"),
					),
				},
				{
					ResourceName:            "btpservice_cicd_job.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"pipeline_parameters"},
				},
			},
		})
	})

	t.Run("happy path - sap-ui5-abap-fes pipeline", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_job_ui5_abap_fes")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-ui5-abap-fes"
  description          = "UI5 ABAP FES test job"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "sap-ui5-abap-fes"
  pipeline_version     = "1.0"
  active               = true
  build_retention_days = 14
  max_builds_to_keep   = 20

  pipeline_parameters = <<-YAML
    configurationSource: job_parameter
    sapUi5FesConfiguration:
      build:
        buildToolVersion: N24
  YAML
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "id"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "name", "tf-test-ui5-abap-fes"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "pipeline", "sap-ui5-abap-fes"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "pipeline_version", "1.0"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "active", "true"),
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "pipeline_parameters"),
					),
				},
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-ui5-abap-fes"
  description          = "UI5 ABAP FES test job - updated"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "sap-ui5-abap-fes"
  pipeline_version     = "1.0"
  active               = false
  build_retention_days = 7
  max_builds_to_keep   = 10

  pipeline_parameters = <<-YAML
    configurationSource: job_parameter
    sapUi5FesConfiguration:
      build:
        buildToolVersion: N24
  YAML
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "description", "UI5 ABAP FES test job - updated"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "active", "false"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "build_retention_days", "7"),
					),
				},
				{
					ResourceName:            "btpservice_cicd_job.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"pipeline_parameters"},
				},
			},
		})
	})

	t.Run("happy path - with notification_configuration", func(t *testing.T) {
		t.Parallel()

		rec, creds := utils.SetupVCR(t, "../fixtures/resource_job_with_ans")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(creds, rec),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-job-ans"
  description          = "Job with ANS notifications"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  active               = true
  build_retention_days = 28
  max_builds_to_keep   = 10

  pipeline_parameters = <<-YAML
    configurationSource: source_repository
  YAML

  notification_configuration = {
    ans = {
      active        = true
      credential_id = "b9a0ba8a-8933-446f-b32e-7ff64ca2e5ce"
      custom_tag    = "my-team"
    }
  }
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "id"),
						resource.TestCheckResourceAttrSet("btpservice_cicd_job.test", "pipeline_parameters"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "notification_configuration.ans.active", "true"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "notification_configuration.ans.credential_id", "b9a0ba8a-8933-446f-b32e-7ff64ca2e5ce"),
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "notification_configuration.ans.custom_tag", "my-team"),
					),
				},
				{
					Config: utils.HCLProviderBlock(creds) + `
resource "btpservice_cicd_job" "test" {
  name                 = "tf-test-job-ans"
  description          = "Job with ANS notifications"
  repository_id        = "4126058b-997f-45b3-9379-4a948b96949f"
  branch               = "main"
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  active               = true
  build_retention_days = 28
  max_builds_to_keep   = 10

  pipeline_parameters = <<-YAML
    configurationSource: source_repository
  YAML

  notification_configuration = {
    ans = {
      active        = false
      credential_id = "b9a0ba8a-8933-446f-b32e-7ff64ca2e5ce"
      custom_tag    = "my-team"
    }
  }
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("btpservice_cicd_job.test", "notification_configuration.ans.active", "false"),
					),
				},
				{
					ResourceName:            "btpservice_cicd_job.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"pipeline_parameters"},
				},
			},
		})
	})

	t.Run("error - missing name", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_job" "test" {
  active               = true
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  pipeline_parameters  = "{}"
  build_retention_days = 28
  max_builds_to_keep   = 15
  branch               = "main"
  repository_id        = "fda133cb-9dae-4d8e-a64d-82fa105f7b2c"
}
`,
					ExpectError: regexp.MustCompile(`The argument "name" is required`),
				},
			},
		})
	})

	t.Run("error - missing active", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_job" "test" {
  name                 = "test-job"
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  pipeline_parameters  = "{}"
  build_retention_days = 28
  max_builds_to_keep   = 15
  branch               = "main"
  repository_id        = "fda133cb-9dae-4d8e-a64d-82fa105f7b2c"
}
`,
					ExpectError: regexp.MustCompile(`The argument "active" is required`),
				},
			},
		})
	})

	t.Run("error - missing pipeline", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_job" "test" {
  name                 = "test-job"
  active               = true
  pipeline_version     = "3.0"
  pipeline_parameters  = "{}"
  build_retention_days = 28
  max_builds_to_keep   = 15
  branch               = "main"
  repository_id        = "fda133cb-9dae-4d8e-a64d-82fa105f7b2c"
}
`,
					ExpectError: regexp.MustCompile(`The argument "pipeline" is required`),
				},
			},
		})
	})

	t.Run("error - invalid pipeline value", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_job" "test" {
  name                 = "test-job"
  active               = true
  pipeline             = "unknown-pipeline"
  pipeline_version     = "3.0"
  pipeline_parameters  = "{}"
  build_retention_days = 28
  max_builds_to_keep   = 15
  branch               = "main"
  repository_id        = "fda133cb-9dae-4d8e-a64d-82fa105f7b2c"
}
`,
					ExpectError: regexp.MustCompile(`(?i)value must be one of`),
				},
			},
		})
	})

	t.Run("error - missing pipeline_version", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_job" "test" {
  name                 = "test-job"
  active               = true
  pipeline             = "cf-env"
  pipeline_parameters  = "{}"
  build_retention_days = 28
  max_builds_to_keep   = 15
  branch               = "main"
  repository_id        = "fda133cb-9dae-4d8e-a64d-82fa105f7b2c"
}
`,
					ExpectError: regexp.MustCompile(`The argument "pipeline_version" is required`),
				},
			},
		})
	})

	t.Run("error - missing pipeline_parameters", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_job" "test" {
  name                 = "test-job"
  active               = true
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  build_retention_days = 28
  max_builds_to_keep   = 15
  branch               = "main"
  repository_id        = "fda133cb-9dae-4d8e-a64d-82fa105f7b2c"
}
`,
					ExpectError: regexp.MustCompile(`The argument "pipeline_parameters" is required`),
				},
			},
		})
	})

	t.Run("error - invalid YAML in pipeline_parameters", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_job" "test" {
  name                 = "test-job"
  active               = true
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  pipeline_parameters  = "key: [invalid yaml"
  build_retention_days = 28
  max_builds_to_keep   = 15
  branch               = "main"
  repository_id        = "fda133cb-9dae-4d8e-a64d-82fa105f7b2c"
}
`,
					ExpectError: regexp.MustCompile(`(?i)invalid yaml`),
				},
			},
		})
	})

	t.Run("error - missing build_retention_days", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_job" "test" {
  name                 = "test-job"
  active               = true
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  pipeline_parameters  = "{}"
  max_builds_to_keep   = 15
  branch               = "main"
  repository_id        = "fda133cb-9dae-4d8e-a64d-82fa105f7b2c"
}
`,
					ExpectError: regexp.MustCompile(`The argument "build_retention_days" is required`),
				},
			},
		})
	})

	t.Run("error - missing max_builds_to_keep", func(t *testing.T) {
		t.Parallel()
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: utils.GetTestProviders(utils.Redacted, nil),
			Steps: []resource.TestStep{
				{
					Config: utils.HCLProviderBlock(utils.Redacted) + `
resource "btpservice_cicd_job" "test" {
  name                 = "test-job"
  active               = true
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  pipeline_parameters  = "{}"
  build_retention_days = 28
  branch               = "main"
  repository_id        = "fda133cb-9dae-4d8e-a64d-82fa105f7b2c"
}
`,
					ExpectError: regexp.MustCompile(`The argument "max_builds_to_keep" is required`),
				},
			},
		})
	})
}
