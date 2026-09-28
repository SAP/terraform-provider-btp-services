# Quick Start Guide

## Introduction

The Terraform provider for SAP BTP Services enables you to automate the provisioning, management, and configuration of resources across [SAP Business Technology Platform](https://account.hana.ondemand.com/) services.

The following services are currently supported:

| Service           | Description                                                      |
|-------------------|------------------------------------------------------------------|
| **CI/CD Service** | Manage credentials, repositories, jobs, and triggers as code     |

## Prerequisites

- Access to an [SAP BTP account](https://account.hana.ondemand.com/)
- Credentials for the service(s) you want to manage — see the service-specific sections below

## Authentication

Each service block in the provider requires its own credentials. The authentication mechanism depends on the service. We strongly recommend providing credentials via environment variables rather than hardcoding them in Terraform configuration files.

## Example configuration

```terraform
terraform {
  required_providers {
    btpservice = {
      source  = "SAP/btp-services"
      version = "<latest>" # Replace <latest> with the latest provider version available on the Terraform Registry.
    }
  }
}

provider "btpservice" {
  cicd {
    # Credentials are read from BTP_CICD_* environment variables
  }
}
```

## Service-specific setup

### CI/CD Service

The CI/CD service uses the **OAuth2 client credentials** flow. The four credentials the provider needs (`endpoint`, `token_url`, `client_id`, `client_secret`) all come from a **service key** that you create on a CI/CD service instance in your BTP subaccount. The steps below walk you through the entire process.

#### Step 1 — Entitle your subaccount

Before you can create a CI/CD service instance your subaccount must have an entitlement for the service.

1. Open the [SAP BTP Cockpit](https://cockpit.btp.cloud.sap) and navigate to your **global account**.
2. Go to **Entitlements → Subaccount Assignments**.
3. Select your subaccount, click **Configure Entitlements**, then **Add Service Plans**.
4. Search for **Continuous Integration & Delivery**, select the plan you want (e.g. `default` or `free`), and save.

> If the entitlement already appears under your subaccount you can skip this step.

#### Step 2 — Subscribe to the CI/CD application *(optional — for the UI)*

The CI/CD service offers both a web UI and an API. The Terraform provider only uses the API, but if you also want to manage pipelines through the browser you need a subscription.

1. In the BTP Cockpit, navigate to your **subaccount**.
2. Go to **Services → Service Marketplace** and find **Continuous Integration & Delivery**.
3. Click the tile and choose **Create** with application plan **`default`**.

#### Step 3 — Create a service instance

This step creates the API-accessible instance that backs the service key.

1. In your subaccount, go to **Services → Service Marketplace → Continuous Integration & Delivery**.
2. Click **Create**.
3. Choose the **`default`** (or `free`) plan, give the instance a name (e.g. `my-cicd-instance`), and click **Create**.

Alternatively, use the BTP CLI:

```bash
btp create services/instance \
  --offering-name continuous-integration-and-delivery \
  --plan-name default \
  --name my-cicd-instance \
  --subaccount <subaccount-id>
```

Or the Cloud Foundry CLI (if your subaccount has a CF environment):

```bash
cf create-service continuous-integration-and-delivery default my-cicd-instance
```

#### Step 4 — Create a service key

A service key generates the credential JSON you will use in the next step.

1. Open the service instance you created in Step 3 (**Services → Instances → my-cicd-instance**).
2. Click **Create Service Key** (or **Service Keys → Create**), give it a name (e.g. `terraform-key`), and confirm.

With the BTP CLI:

```bash
btp create services/key \
  --name terraform-key \
  --instance-name my-cicd-instance \
  --subaccount <subaccount-id>
```

With the CF CLI:

```bash
cf create-service-key my-cicd-instance terraform-key
```

#### Step 5 — Read the service key JSON

1. In the BTP Cockpit, open the service key you just created and click **View**.
2. You will see a JSON document similar to the following:

```json
{
  "api": "https://cicd.cfapps.<region>.hana.ondemand.com",
  "uaa": {
    "clientid": "sb-clone-xxxx!bXXXX|cicd-service!bXXXX",
    "clientsecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx=",
    "url": "https://<subaccount-id>.authentication.<region>.hana.ondemand.com",
    ...
  }
}
```

With the BTP CLI:

```bash
btp get services/key terraform-key \
  --instance-name my-cicd-instance \
  --subaccount <subaccount-id>
```

With the CF CLI:

```bash
cf service-key my-cicd-instance terraform-key
```

#### Step 6 — Map the service key fields to provider attributes

Use the following table to convert the service key JSON into the provider values:

| Provider attribute / environment variable              | Service key JSON field     | How to derive it                                |
|--------------------------------------------------------|----------------------------|-------------------------------------------------|
| `endpoint` / `BTP_CICD_ENDPOINT`                       | `api`                      | Copy the value directly                         |
| `client_id` / `BTP_CICD_CLIENT_ID`                     | `uaa.clientid`             | Copy the value directly                         |
| `client_secret` / `BTP_CICD_CLIENT_SECRET`             | `uaa.clientsecret`         | Copy the value directly                         |
| `token_url` / `BTP_CICD_TOKEN_URL`                     | `uaa.url`                  | Append `/oauth/token` to the value              |

For example, if `uaa.url` is `https://my-subaccount.authentication.eu10.hana.ondemand.com`, then `token_url` is `https://my-subaccount.authentication.eu10.hana.ondemand.com/oauth/token`.

#### Step 7 — Export the environment variables

We strongly recommend supplying credentials via environment variables rather than hardcoding them in Terraform files.

**Mac / Linux**

```bash
export BTP_CICD_ENDPOINT="https://cicd.cfapps.<region>.hana.ondemand.com"
export BTP_CICD_TOKEN_URL="https://<subaccount-id>.authentication.<region>.hana.ondemand.com/oauth/token"
export BTP_CICD_CLIENT_ID="sb-clone-xxxx!bXXXX|cicd-service!bXXXX"
export BTP_CICD_CLIENT_SECRET="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx="
```

**Windows (CMD)**

```shell
set BTP_CICD_ENDPOINT=https://cicd.cfapps.<region>.hana.ondemand.com
set BTP_CICD_TOKEN_URL=https://<subaccount-id>.authentication.<region>.hana.ondemand.com/oauth/token
set BTP_CICD_CLIENT_ID=sb-clone-xxxx!bXXXX|cicd-service!bXXXX
set BTP_CICD_CLIENT_SECRET=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx=
```

**Windows (PowerShell)**

```shell
$Env:BTP_CICD_ENDPOINT = "https://cicd.cfapps.<region>.hana.ondemand.com"
$Env:BTP_CICD_TOKEN_URL = "https://<subaccount-id>.authentication.<region>.hana.ondemand.com/oauth/token"
$Env:BTP_CICD_CLIENT_ID = "sb-clone-xxxx!bXXXX|cicd-service!bXXXX"
$Env:BTP_CICD_CLIENT_SECRET = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx="
```

The supported environment variables are:

| Environment Variable     | Description                        |
|--------------------------|------------------------------------|
| `BTP_CICD_ENDPOINT`      | CI/CD service base URL             |
| `BTP_CICD_TOKEN_URL`     | OAuth2 token endpoint              |
| `BTP_CICD_CLIENT_ID`     | OAuth2 client ID                   |
| `BTP_CICD_CLIENT_SECRET` | OAuth2 client secret *(sensitive)* |

## Documentation

Terraform Provider for SAP BTP Services [Documentation](https://registry.terraform.io/providers/SAP/btp-services/latest/docs)
