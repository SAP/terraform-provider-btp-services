# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for SAP BTP (Business Technology Platform) services. It is currently a work in progress. The first service targeted is the CI/CD Service.

## Status

The repository is in early setup — no Go module, provider scaffold, or tests exist yet. When the project is built out, update this file with build commands, test commands, and architecture details.

## Expected Architecture (standard Terraform provider pattern)

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
- `make test` - Run unit tests with coverage (tagged tests included)
- `make testacc` - Run acceptance tests (requires `TF_ACC=1`, long-running, needs live BTP credentials)
- `go test -v -run TestResourceSubaccountServiceInstance ./btp/provider/` - Run specific test

Development setup:
- Configure Terraform CLI dev override in `~/.terraformrc` (Mac/Linux) or `%APPDATA%/terraform.rc` (Windows):
  ```hcl
  provider_installation {
    dev_overrides {
      "sap/btp" = "/path/to/go/bin"
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

Once scaffolded, the common commands will likely be:

```bash
go build ./...          # Build
go test ./...           # Run all tests
make testacc            # Run acceptance tests (requires BTP credentials)
golangci-lint run       # Lint
```

**Testing:**
- Uses `terraform-plugin-testing` framework
- VCR (go-vcr) recordings in `fixtures/` reduce live API dependency
- Test naming: `TestResource<Name>` or `TestDataSource<Name>`
- Include import state verification in tests

## Documentation Generation

- **NEVER** manually edit files in `docs/` - they are generated
- Modify code comments (especially schema MarkdownDescription fields) and `templates/` instead
- Run `make generate` to regenerate docs
- Generated docs power the Terraform Registry documentation

## Development Workflow

1. Start with similar existing resource/datasource/list_resource/function/action as template
2. Implement schema with proper types, validators, descriptions
3. Add CRUD/List logic delegating to `internal/btpcli`
4. Write tests in `*_test.go` with VCR fixtures
5. **MANDATORY after every change:**
   - `make lint` - Fix any linting issues
   - `make fix` - Apply automatic fixes
   - `make build` - Verify compilation succeeds
6. Test: `make test`
7. Generate docs: `make generate`
8. Install locally: `make install`
9. Verify with example: `cd examples/provider/ && terraform validate`

## Commit Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat: add resource for subaccount subscription`
- `fix: handle nil pointer in service instance read`
- `docs: update examples for trust configuration`
- `refactor!: breaking change to schema`
- `feat(btp_subaccount): scoped feature addition`

## Common Pitfalls

1. **Package declarations**: Each Go file has exactly ONE `package` declaration. When editing existing files, preserve the existing package line - never duplicate it.

2. **Dev overrides**: When using local dev overrides, do NOT run `terraform init` - it's unnecessary and will error.

3. **Test failures**: If acceptance tests fail, ensure:
   - `BTP_USERNAME` and `BTP_PASSWORD` env vars are set
   - VCR fixtures exist or test is marked for live API calls
   - Timeout is sufficient for long-running operations

4. **Generated docs**: Changes to `docs/*.md` will be overwritten. Update code comments and run `make generate`.

5. **Error handling**: Always return diagnostics via `resp.Diagnostics.Append()` - never panic in provider code.

6. **Schema stability**: Keep attribute names stable across versions. Use deprecation warnings for schema changes.

## Testing Strategy

- Unit tests: Fast, use VCR recordings where possible
- Acceptance tests: Slower, may require live BTP account
- Integration tests: In `tests/integration-test/` folder, full Terraform scenarios
- Regression tests: In `tests/regression-test/` and `regression-test/` folders

VCR setup in tests:
```go
rec, user := setupVCR(t, "fixtures/resource_subaccount_service_instance.wo_parameters")
defer stopQuietly(rec)
```

## Security Considerations

- No hardcoded credentials - use environment variables
- Mark sensitive attributes with `Sensitive: true` in schema
- Redact sensitive data in logs and VCR recordings
- Keep dependencies updated for security patches
