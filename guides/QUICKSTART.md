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

The CI/CD service uses the **OAuth2 client credentials** flow. The credentials are available in the service key of a CI/CD service instance in your BTP subaccount. For detailed steps on how to set up the service and retrieve the credentials, refer to the [SAP Help Portal — Initial Setup of SAP Continuous Integration and Delivery](https://help.sap.com/docs/continuous-integration-and-delivery/sap-continuous-integration-and-delivery/initial-setup).

The following environment variables are supported:

| Environment Variable     | Description                        |
|--------------------------|------------------------------------|
| `BTP_CICD_ENDPOINT`      | CI/CD service base URL             |
| `BTP_CICD_TOKEN_URL`     | OAuth2 token endpoint              |
| `BTP_CICD_CLIENT_ID`     | OAuth2 client ID                   |
| `BTP_CICD_CLIENT_SECRET` | OAuth2 client secret *(sensitive)* |

#### Mac / Linux

```bash
export BTP_CICD_ENDPOINT="https://cicd-service-url.cfapps.<region>.hana.ondemand.com"
export BTP_CICD_TOKEN_URL="https://<subaccount>.authentication.<region>.hana.ondemand.com/oauth/token"
export BTP_CICD_CLIENT_ID="sb-clone-xxxx!bXXXX|cicd-service!bXXXX"
export BTP_CICD_CLIENT_SECRET="your-client-secret"
```

#### Windows (CMD)

```shell
set BTP_CICD_ENDPOINT=https://cicd-service-url.cfapps.<region>.hana.ondemand.com
set BTP_CICD_TOKEN_URL=https://<subaccount>.authentication.<region>.hana.ondemand.com/oauth/token
set BTP_CICD_CLIENT_ID=sb-clone-xxxx!bXXXX|cicd-service!bXXXX
set BTP_CICD_CLIENT_SECRET=your-client-secret
```

#### Windows (PowerShell)

```shell
$Env:BTP_CICD_ENDPOINT = "https://cicd-service-url.cfapps.<region>.hana.ondemand.com"
$Env:BTP_CICD_TOKEN_URL = "https://<subaccount>.authentication.<region>.hana.ondemand.com/oauth/token"
$Env:BTP_CICD_CLIENT_ID = "sb-clone-xxxx!bXXXX|cicd-service!bXXXX"
$Env:BTP_CICD_CLIENT_SECRET = "your-client-secret"
```

## Documentation

Terraform Provider for SAP BTP Services [Documentation](https://registry.terraform.io/providers/SAP/btp-services/latest/docs)
