# terraform import btpservice_cicd_job.<resource_name> <id>

terraform import btpservice_cicd_job.example pb091fd5-845b-4146-9bfc-d8cb74be04f8

# terraform import using id attribute in import block

import {
  to = btpservice_cicd_job.<resource_name>
  id = "<id>"
}

import {
  to =  btpservice_cicd_job.<resource_name>
  identity = {
   id = "<id>"
  }
}
