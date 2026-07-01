# Import

## Overview

In general Terraform supports the *import* of resources into the Terraform state. You find the official documentation on how to achieve this [here](https://developer.hashicorp.com/terraform/cli/import).

The Terraform provider for SAP BTP Services supports the import of resources as well. [The documentation](https://registry.terraform.io/providers/SAP/btp-services/latest/docs) of the Terraform provider for SAP BTP Services provides the necessary information on how to import a resource and which keys to use on the level of each resource.

To get a quick overview of the resources and if they support the import functionality, you can refer to the [Resource Overview](#resource-overview) section in this document.

## Resource Overview

The following list provides an overview of the resources and their support for the import functionality (state: 01.01.2026)

| Resource                                                   | Import Support
|---                                                         |---
| btpservice_cicd_allowed_spaces                             | No
| btpservice_cicd_credential_basic_auth                      | No
| btpservice_cicd_credential_basic_auth_custom_idp           | No
| btpservice_cicd_credential_cert_based_auth_custom_idp      | No
| btpservice_cicd_credential_cloud_connector                 | No
| btpservice_cicd_credential_container_registry              | No
| btpservice_cicd_credential_kubernetes_config               | No
| btpservice_cicd_credential_secret_text                     | No
| btpservice_cicd_credential_service_key                     | No
| btpservice_cicd_credential_webhook_secret                  | No
| btpservice_cicd_job                                        | Yes
| btpservice_cicd_repository                                 | Yes
| btpservice_cicd_trigger                                    | Yes
