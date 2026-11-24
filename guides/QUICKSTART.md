# Quick Start Guide

## Introduction

The Terraform provider for SAP BTP Services enables you to automate the management and configuration of resources across [SAP Business Technology Platform](https://account.hana.ondemand.com/) services.

The setup follows the same flow for each service: assign an entitlement, create a service instance and service key, then configure the corresponding provider block. Repeat the flow for each service you want to manage. If you already have an instance and service key, start with the credential mapping in Step 4.

## Prerequisites

- Access to an SAP BTP subaccount and permissions to manage entitlements, service instances, and service keys
- Terraform installed; use version 1.14 or later for the examples in this guide (required for cTMS)
- A service plan available in your subaccount and region

## Step 1 — Choose your service and assign an entitlement

Use the table below to identify the service and its provider configuration. The shared steps cover both services; service-specific values and exceptions are listed here and in the credential mapping.

| Service | Service Marketplace name | Provider block | Environment variable prefix | Managed resources |
|---------|--------------------------|----------------|-----------------------------|-------------------|
| SAP Continuous Integration and Delivery (CI/CD) | Continuous Integration & Delivery | `cicd` | `BTP_CICD_` | Credentials, repositories, jobs, and triggers |
| SAP Cloud Transport Management (cTMS) | Cloud Transport Management | `ctms` | `BTP_CTMS_` | Nodes and routes |

1. Open the [SAP BTP Cockpit](https://cockpit.btp.cloud.sap) and navigate to your **global account**.
2. Go to **Entitlements → Subaccount Assignments**.
3. Select your subaccount, click **Configure Entitlements**, then **Add Service Plans**.
4. Find your service, select the service plan you need, and save.

If the required entitlement is already assigned, continue to Step 2. Plan names and availability depend on the service and region; use the service instance plan offered for your subaccount.

The provider uses the service API. If you also need the service's web UI, follow that service's instructions for application subscriptions and role assignments.

## Step 2 — Create a service instance

Create an instance of the selected service using one of the following options. Use a descriptive name, such as `my-cicd-instance` or `my-ctms-instance`.

### Option A — BTP Cockpit

1. In your subaccount, go to **Services → Service Marketplace** and open the service from the table above.
2. Click **Create** and choose a **service instance**, where prompted.
3. Select the service plan and runtime environment supported by the service, enter an instance name, and complete the creation dialog.
4. Wait until the instance is ready.

### Option B — Cloud Foundry CLI

If you use a Cloud Foundry environment, target the appropriate organization and space. Run `cf marketplace` to find the offering and plan names, then replace the placeholders below:

```bash
cf create-service <service-offering> <service-plan> <instance-name>
```

### Option C — SAP BTP Terraform Provider

You can also provision the instance using the [SAP BTP Terraform Provider](https://registry.terraform.io/providers/SAP/btp/latest/docs). Configure that provider separately with access to your subaccount, then use its [`btp_subaccount_service_instance`](https://registry.terraform.io/providers/SAP/btp/latest/docs/resources/subaccount_service_instance) resource:

```terraform

resource "btp_subaccount_service_instance" "service" {
  subaccount_id         = "<subaccount-id>"
  service_offering_name = "<service-offering>"
  serviceplan_name      = "<service-plan>"
  name                  = "<instance-name>"
}
```

Use the technical offering and plan names from your subaccount's service marketplace. For example, CI/CD uses offering `continuous-integration-and-delivery` with an available plan such as `default` or `free`.

## Step 3 — Create and read a service key

A service key provides the credential JSON needed to authenticate the provider. Create a key on each service instance you want to use.

### Option A — BTP Cockpit

1. Go to **Services → Instances and Subscriptions** and open your service instance.
2. Under **Service Keys**, choose **Create**, enter a name such as `terraform-key`, and confirm. Supply any service-specific parameters requested by the creation dialog.
3. Open the new service key and view its JSON credentials.

### Option B — Cloud Foundry CLI

```bash
cf create-service-key <instance-name> terraform-key
cf service-key <instance-name> terraform-key
```

### Option C — SAP BTP Terraform Provider

Use the [`btp_subaccount_service_binding`](https://registry.terraform.io/providers/SAP/btp/latest/docs/resources/subaccount_service_binding) resource with the instance from Step 2:

```terraform
resource "btp_subaccount_service_binding" "service" {
  subaccount_id       = "<subaccount-id>"
  service_instance_id = btp_subaccount_service_instance.service.id
  name                = "terraform-key"
}
```

The binding exposes the service key JSON in its `credentials` attribute. Apply this provisioning configuration before using the credentials to configure the SAP BTP Services provider; cTMS requires its configuration values to be known before making API requests. The provisioning state contains these credentials and should be stored securely.

## Step 4 — Map credentials to provider settings

CI/CD and cTMS both use the **OAuth2 client credentials** flow. Each service block needs credentials from its own service key. The relevant fields have the following structure (other fields are omitted):

```json
{
  "api": "https://<service-api-host>",
  "uaa": {
    "clientid": "<client-id>",
    "clientsecret": "<client-secret>",
    "url": "https://<authentication-host>"
  }
}
```

| Provider attribute | Service key value | CI/CD environment variable | cTMS environment variable |
|--------------------|-------------------|----------------------------|---------------------------|
| `endpoint` | `api` | `BTP_CICD_ENDPOINT` | `BTP_CTMS_ENDPOINT` |
| `client_id` | `uaa.clientid` | `BTP_CICD_CLIENT_ID` | `BTP_CTMS_CLIENT_ID` |
| `client_secret` | `uaa.clientsecret` | `BTP_CICD_CLIENT_SECRET` | `BTP_CTMS_CLIENT_SECRET` |
| `token_url` | `uaa.url` with `/oauth/token` appended | `BTP_CICD_TOKEN_URL` | `BTP_CTMS_TOKEN_URL` |

For example, `uaa.url` value `https://my-subaccount.authentication.eu10.hana.ondemand.com` becomes token URL `https://my-subaccount.authentication.eu10.hana.ondemand.com/oauth/token`.

For **cTMS**, set `endpoint` to the service root **without `/v2`**; the provider adds the API version.
Both the cTMS service endpoint and token URL must use HTTPS. Redirects to HTTP are rejected by the default cTMS client.

As support for more services is added, use the same setup flow and consult the [provider documentation](https://registry.terraform.io/providers/SAP/btp-services/latest/docs) for each new service's block name, authentication requirements, and credential mapping.

## Step 5 — Supply credentials

We recommend supplying credentials through environment variables rather than hardcoding them in Terraform files. Set all four variables for each service you enable, using the mapping above.

### macOS / Linux

For CI/CD:

```bash
export BTP_CICD_ENDPOINT="<api>"
export BTP_CICD_TOKEN_URL="<uaa.url>/oauth/token"
export BTP_CICD_CLIENT_ID="<uaa.clientid>"
export BTP_CICD_CLIENT_SECRET="<uaa.clientsecret>"
```

For cTMS, use values from your cTMS service key:

```bash
export BTP_CTMS_ENDPOINT="<api-without-/v2>"
export BTP_CTMS_TOKEN_URL="<uaa.url>/oauth/token"
export BTP_CTMS_CLIENT_ID="<uaa.clientid>"
export BTP_CTMS_CLIENT_SECRET="<uaa.clientsecret>"
```

### Windows

In PowerShell, use `$Env:NAME = "value"`; in CMD, use `set "NAME=value"`. For example:

```powershell
$Env:BTP_CTMS_ENDPOINT = "<api-without-/v2>"
```

```shell
set "BTP_CTMS_ENDPOINT=<api-without-/v2>"
```

Set the remaining variables using the same syntax and the names in Step 4.

## Step 6 — Configure the provider and run Terraform

Create a Terraform configuration with a block for each service you want to use. This example enables both CI/CD and cTMS; remove any service block you do not need. An empty block enables the service and reads its credentials from the environment. Explicit attribute values override environment variables.

```terraform
terraform {
  required_version = ">= 1.14.0"

  required_providers {
    btpservice = {
      source  = "SAP/btp-services"
      version = "<latest>" # Replace with a published version from the Terraform Registry.
    }
  }
}

provider "btpservice" {
  cicd {
    # Credentials are read from BTP_CICD_* environment variables.
  }

  ctms {
    # Credentials are read from BTP_CTMS_* environment variables.
  }
}
```

Add the resources or data sources you need from the [examples](../examples/) or the [provider documentation](https://registry.terraform.io/providers/SAP/btp-services/latest/docs). For a first read of existing service configuration, use either or both of these data sources, keeping only those for your enabled services:

```terraform
data "btpservice_cicd_repositories" "all" {}

data "btpservice_ctms_nodes" "all" {}
```

Run Terraform from the directory containing your configuration, in the environment where you set the credentials:

```bash
terraform init
terraform plan
terraform apply
```

Review the plan before confirming the apply. Repeat the same setup for another service instance when you expand your configuration.
