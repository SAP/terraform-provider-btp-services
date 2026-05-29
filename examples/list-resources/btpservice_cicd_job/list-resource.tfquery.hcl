# This feature requires Terraform v1.14.0 or later (Stable as of 2026).
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block.
list "btpservice_cicd_job" "<label_name>" {
  # (Required) Provider instance to use.
  provider = btpservice

  # (Optional) Return full resource objects instead of only identities.
  # include_resource = true

  # (Optional) Filter attributes must be specified inside a config {} block.
  # config {
  #   # Filter by pipeline type. One of: cpi, cf-env, kyma-cnb, sap-ui5-abap-fes
  #   pipeline = "cf-env"
  #
  #   # Filter by repository ID.
  #   repository_id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  # }
}

# List all jobs — returns only identities by default.
list "btpservice_cicd_job" "all" {
  provider = btpservice
}

# List all jobs with full resource details.
list "btpservice_cicd_job" "all_with_resource" {
  provider         = btpservice
  include_resource = true
}

# List only cf-env jobs with full resource details.
list "btpservice_cicd_job" "cf_env" {
  provider         = btpservice
  include_resource = true
  config {
    pipeline = "cf-env"
  }
}

# List all jobs for a specific repository.
list "btpservice_cicd_job" "by_repository" {
  provider = btpservice
  config {
    repository_id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
