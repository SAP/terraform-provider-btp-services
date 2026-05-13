# This feature requires Terraform v1.14.0 or later.
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block.
list "btpservice_cicd_credential_service_key" "<label_name>" {
  # (Required) Provider instance to use.
  provider = btpservice

  # (Optional) Return full resource objects instead of only identities.
  # include_resource = true
}

# Discover all Service Key credentials — identities only.
list "btpservice_cicd_credential_service_key" "all" {
  provider = btpservice
}

# Discover all Service Key credentials with full resource details.
list "btpservice_cicd_credential_service_key" "with_resource" {
  provider         = btpservice
  include_resource = true
}
