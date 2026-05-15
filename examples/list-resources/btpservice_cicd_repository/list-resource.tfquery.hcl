# This feature requires Terraform v1.14.0 or later (Stable as of 2026).
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block.
list "btpservice_cicd_repository" "<label_name>" {
  # (Required) Provider instance to use.
  provider = btpservice

  # (Optional) Return full resource objects instead of only identities.
  # include_resource = true
}

# List block to discover all repositories in the SAP BTP CI/CD service.
# Returns only the resource identities by default.
list "btpservice_cicd_repository" "all" {
  provider = btpservice
}

# List block to discover all repositories with full resource details.
# Setting include_resource = true returns full resource objects (e.g., clone_url, name).
list "btpservice_cicd_repository" "with_resource" {
  provider         = btpservice
  include_resource = true
}

