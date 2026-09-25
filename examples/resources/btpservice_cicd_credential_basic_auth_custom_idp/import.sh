# terraform import btpservice_cicd_credential_basic_auth_custom_idp.<resource_name> <credential_name>

terraform import btpservice_cicd_credential_basic_auth_custom_idp.example my-custom-idp-user

# terraform import using id attribute in import block

import {
  to = btpservice_cicd_credential_basic_auth_custom_idp.<resource_name>
  id = "<credential_name>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
  to = btpservice_cicd_credential_basic_auth_custom_idp.<resource_name>
  identity = {
    id = "<credential_name>"
  }
}
