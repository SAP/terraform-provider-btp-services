![Golang](https://img.shields.io/badge/Go-1.26-informational)
[![Go Report Card](https://goreportcard.com/badge/github.com/SAP/terraform-provider-btp-services)](https://goreportcard.com/report/github.com/SAP/terraform-provider-btp-services)
[![REUSE status](https://api.reuse.software/badge/github.com/SAP/terraform-provider-btp-services)](https://api.reuse.software/info/github.com/SAP/terraform-provider-btp-services)

# Terraform Provider for SAP BTP Services

## About This Project

The Terraform provider for SAP BTP Services enables dedicated management of services available within [SAP Business Technology Platform](https://www.sap.com/products/technology-platform.html) via [Terraform](https://terraform.io/).

You will find the detailed information about the provider in the official [Terraform Registry](https://registry.terraform.io/providers/SAP/btp-services).

You find usage examples in the [examples folder](./examples/) of this repository.

## Usage of the Provider

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    btpservice = {
      source = "sap/btp-services"
    }
  }
}

provider "btpservice" {
  cicd {
    endpoint      = "https://cicd-service-url.cfapps.us10.hana.ondemand.com"
    token_url     = "https://your-subaccount.authentication.us10.hana.ondemand.com/oauth/token"
    client_id     = "sb-clone-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx!b12345|cicd-service!b6789"
    client_secret = "your-client-secret-value-here="
  }
}
```

## Developing & Contributing to the Provider

Refer to the [developer documentation](DEVELOPER.md) for instructions on how to build and develop the provider locally.

## Support, Feedback, Contributing

❓ - If you have a *question*, open a [GitHub Discussion](https://github.com/SAP/terraform-provider-btp-services/discussions/) or ask in the [SAP Community](https://answers.sap.com/questions/ask.html).

🐞 - If you find a bug, please open a [bug report](https://github.com/SAP/terraform-provider-btp-services/issues/new?labels=bug%2Cneeds-triage&template=bug_report.yml&title=%5BBUG%5D).

💡 - If you have a feature idea, please open a [feature request](https://github.com/SAP/terraform-provider-btp-services/issues/new?labels=enhancement%2Cneeds-triage&template=feature_request.yml&title=%5BFEATURE%5D).

For more information about how to contribute, the project structure, and additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

> **Note**: We take Terraform's security and our users' trust seriously. If you believe you have found a security issue in the Terraform provider for SAP BTP Services, please responsibly disclose it. You find more details on the process in [our security policy](https://github.com/SAP/terraform-provider-btp-services/security/policy).

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/SAP/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright 2026 SAP SE or an SAP affiliate company and terraform-provider-btp-services contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/SAP/terraform-provider-btp-services).

## OpenTofu Compatibility

The Terraform Provider for SAP BTP Services supports [OpenTofu](https://opentofu.org/) under the following conditions:
1. **Drop-In Replacement**: The provider can be used with [OpenTofu CLI](https://opentofu.org/docs/cli/) as a direct replacement for [HashiCorp Terraform CLI](https://developer.hashicorp.com/terraform/cli) without modifications.
2. **Feature Limitations**: The provider does not support OpenTofu specific features or functions outside the standard Terraform functionality.
3. **Issue Reporting**: Any issues reported for the Terraform Provider for SAP BTP Services will only be addressed if they are reproducible using the Terraform CLI.
