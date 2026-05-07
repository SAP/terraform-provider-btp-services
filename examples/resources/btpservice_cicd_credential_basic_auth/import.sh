# terraform import btpservice_cicd_credential_basic_auth.<resource_name> <id>

terraform import btpservice_cicd_credential_basic_auth.parent dd005d8b-1fee-4e6b-b6ff-cb9a197b7fe0

# terraform import using id attribute in import block

import {
  to = btpservice_cicd_credential_basic_auth.<resource_name>
  id = "<id>"
}

import {
  to =  btpservice_cicd_credential_basic_auth.<resource_name>
  identity = {
   id = "<id>"
  }
}