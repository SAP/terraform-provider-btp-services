# This feature requires Terraform v1.14.0 or later (Stable as of 2026).
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block.
list "btpservice_cicd_build_trigger" "<label_name>" {
  # (Required) Provider instance to use.
  provider = btpservice

  # (Optional) Return full resource objects instead of only identities.
  # include_resource = true

  config {
    # (Required) Name or ID of the CI/CD job whose triggers to list.
    job = "<job_name_or_id>"
  }
}

# List block to discover all build triggers for a job.
# Returns only the resource identities by default.
list "btpservice_cicd_build_trigger" "all" {
  provider = btpservice
  config {
    job = "my-pipeline-job"
  }
}

# List block to discover all build triggers with full resource details.
# Setting include_resource = true returns full resource objects (e.g., type, timer schedule).
list "btpservice_cicd_build_trigger" "with_resource" {
  provider         = btpservice
  include_resource = true
  config {
    job = "my-pipeline-job"
  }
}
