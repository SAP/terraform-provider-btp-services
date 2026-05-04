# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for SAP BTP (Business Technology Platform) services. The first service targeted is the CI/CD Service (credentials management). Provider type name: `btpservice` (resources are named `btpservice_cicd_credential_basic_auth` etc.).

Module path: `github.com/SAP/terraform-provider-sap-btp-services`

## Essential Commands

Build and development:
- `make fmt` - Format code with gofmt
- `make fix` - Run go fix to update code to newer Go versions
- `make lint` - Run golangci-lint (must pass before commits)
- `make build` - Compile the provider
- `make install` - Build and install to `$GOBIN` for local Terraform dev override
- `make generate` - Generate documentation from code annotations and templates

**CRITICAL: After every code change, always run in order:**
1. `make lint` - Check for linting issues
2. `make fix` - Apply automatic fixes
3. `make build` - Verify compilation

Testing:
- `make test` - Run unit tests with coverage (all tags, parallel=4)
- `make testacc` - Run acceptance tests (requires `TF_ACC=1`, long-running, needs live BTP credentials)
- `go test -v -run TestResourceCredentialBasicAuth ./btpservices/provider/cicd/credentials/` - Run specific resource test
- `go test -v -run TestDataSourceCredential ./btpservices/provider/cicd/credentials/` - Run specific datasource test

Development setup:
- Configure Terraform CLI dev override in `~/.terraformrc` (Mac/Linux) or `%APPDATA%/terraform.rc` (Windows):
  ```hcl
  provider_installation {
    dev_overrides {
      "sap/btp-services" = "/path/to/go/bin"
    }
    direct {}
  }
  ```
- Do NOT run `terraform init` when using dev overrides
- Verify setup: `cd examples/provider/ && terraform validate`

Pre-commit hooks (via Lefthook):
- `make lefthook` - Install Lefthook and register the pre-commit hooks
- Hooks run automatically on commit: `go fmt`, `golangci-lint --fix`, `terraform fmt`
- Install once after cloning: `make lefthook`

## Folder Layout

```
btpservices/provider/
  provider.go                                           # package btpservicesprovider — provider schema + Configure()
  provider_test.go
  testutil/vcr.go                                       # package testutil — generic VCR helper
  cicd/
    service_package.go                                  # package cicd — registers resources/datasources
    cicdtest/testhelper.go                              # package cicdtest — SetupVCR() for CI/CD tests
    fixtures/                                           # VCR cassettes (YAML) shared by all cicd tests
    credentials/
      resource_credential_basic_auth.go                 # CRUD resource
      resource_credential_basic_auth_test.go
      datasource_credential.go                          # single credential data source
      datasource_credential_test.go
      datasource_credentials.go                         # list credentials data source
      datasource_credentials_test.go
      types.go                                          # model structs + valueFrom() + toRequest()

internal/
  shared/
    provider_clients.go                                 # package shared — ProviderClients struct
  cicd/
    client/
      client.go                                         # HTTP transport + OAuth2 token caching
      client_config.go
      facade.go                                         # CicdClientFacade interface
      facade_credentials.go                             # credential CRUD methods
      facade_credentials_test.go
      facade_test.go
    models/
      credential.go                                     # API request/response structs
      errors.go
      errors_test.go

docs/                                                   # GENERATED — never edit manually
examples/                                               # example Terraform configs
```

## Architecture

### Adding a new service
1. Create `internal/<svc>/client/` and `internal/<svc>/models/` for HTTP transport + models
2. Create `btpservices/provider/<svc>/` for provider wiring (service_package.go + resources/datasources)
3. Add a field to `internal/shared/provider_clients.go`
4. Register `ServicePackage{}` in `btpservices/provider/provider.go::servicePackages()`

### Auth pattern
OAuth2 client_credentials flow. The client fetches and caches the token internally.
Config fields: `endpoint`, `token_url`, `client_id`, `client_secret`.
Env vars: `SAPBTP_CICD_ENDPOINT`, `SAPBTP_CICD_TOKEN_URL`, `SAPBTP_CICD_CLIENT_ID`, `SAPBTP_CICD_CLIENT_SECRET`.

### CI/CD API
- All endpoints use `/v2/` prefix (e.g. `GET /v2/credentials`)
- POST returns 201 with no body — must GET by name after create to obtain ID
- PUT/PATCH return 204 with no body; PATCH is a merge-patch (name IS mutable)
- DELETE returns 204 with no body
- Password is **never** returned on read — preserve from prior state
- List response: `{ "_embedded": { "credentials": [...] } }`
- BasicAuth credential shape: `{ "name", "description", "basic": { "username", "password" } }`

## Documentation Generation

- **NEVER** manually edit files in `docs/` — they are generated
- Modify code comments (schema `MarkdownDescription` fields) and `templates/` instead
- Run `make generate` to regenerate docs

## Development Workflow

1. Use an existing resource/datasource as a template
2. Implement schema with proper types, validators, descriptions
3. Add CRUD logic delegating to `internal/<svc>/client/`
4. Write tests in `*_test.go` with VCR cassettes in `fixtures/`
5. **MANDATORY after every change:** `make lint` → `make fix` → `make build`
6. `make test`
7. `make generate`
8. `make install` + `cd examples/provider/ && terraform validate`

**Testing:**
- Uses `terraform-plugin-testing` framework
- VCR (go-vcr v3.2.1) recordings in `btpservices/provider/cicd/fixtures/` reduce live API dependency
- `IsUnitTest: true` on all test cases — run without `TF_ACC=1`
- `ImportStateVerifyIgnore: []string{"password"}` — API never returns password on read
- Test naming: `TestResource<Name>` or `TestDataSource<Name>`
- Include import state verification in tests

## Commit Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat: add resource for cicd pipeline`
- `fix: handle nil pointer in credential read`
- `docs: update examples for basic auth credential`
- `refactor!: breaking change to schema`
- `feat(sapbtp_cicd): scoped feature addition`

## Common Pitfalls

1. **Package declarations**: Each Go file has exactly one `package` declaration. Never duplicate it.

2. **Dev overrides**: Do NOT run `terraform init` — it will error.

3. **Test credentials**: If acceptance tests fail, ensure `SAPBTP_CICD_*` env vars are set and cassettes exist.

4. **Generated docs**: Changes to `docs/*.md` will be overwritten. Update code comments and run `make generate`.

5. **Error handling**: Always return diagnostics via `resp.Diagnostics.Append()` — never panic.

6. **Schema stability**: Keep attribute names stable across versions. Use deprecation warnings for schema changes.

## Security Considerations

- No hardcoded credentials — use environment variables
- Mark sensitive attributes with `Sensitive: true` in schema
- Redact sensitive data in logs and VCR cassettes
- Keep dependencies updated for security patches
