# terraform import btpservice_cicd_credential_service_key.<resource_name> <id>

terraform import btpservice_cicd_credential_service_key.example dd005d8b-1fee-4e6b-b6ff-cb9a197b7fe0

import {
  to = btpservice_cicd_credential_service_key.example
  id = "dd005d8b-1fee-4e6b-b6ff-cb9a197b7fe0"
}

import {
  to = btpservice_cicd_credential_service_key.example
  identity = {
    id = "dd005d8b-1fee-4e6b-b6ff-cb9a197b7fe0"
  }
}
