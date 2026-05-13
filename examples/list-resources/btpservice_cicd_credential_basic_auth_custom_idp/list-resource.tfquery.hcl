# This feature requires Terraform v1.14.0 or later.
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block.
list "btpservice_cicd_credential_basic_auth_custom_idp" "<label_name>" {
  # (Required) Provider instance to use.
  provider = btpservice

  # (Optional) Return full resource objects instead of only identities.
  # include_resource = true
}

# Discover all Basic Auth (Custom IdP) credentials — identities only.
list "btpservice_cicd_credential_basic_auth_custom_idp" "all" {
  provider = btpservice
}

# Discover all Basic Auth (Custom IdP) credentials with full resource details.
list "btpservice_cicd_credential_basic_auth_custom_idp" "with_resource" {
  provider         = btpservice
  include_resource = true
}
