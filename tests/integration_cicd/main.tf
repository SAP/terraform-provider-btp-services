# =============================================================================
# Step 1 — Credentials (all nine types)
# =============================================================================

resource "btpservice_cicd_credential_basic_auth" "deploy" {
  name        = "it-deploy-user"
  description = "Integration test: CF deployment user"
  username    = var.deploy_username
  password    = var.deploy_password
}

resource "btpservice_cicd_credential_webhook_secret" "webhook" {
  name        = "it-webhook-secret"
  description = "Integration test: GitHub webhook secret"
  token       = var.github_webhook_secret
}

resource "btpservice_cicd_credential_cloud_connector" "cc" {
  name        = "it-cloud-connector"
  description = "Integration test: Cloud Connector location"
  location_id = "integration-test-location"
}

resource "btpservice_cicd_credential_container_registry" "registry" {
  name        = "it-container-registry"
  description = "Integration test: Container registry config"
  content = jsonencode({
    auths = {
      "registry.example.com" = {
        auth = base64encode("${var.deploy_username}:${var.deploy_password}")
      }
    }
  })
}

resource "btpservice_cicd_credential_kubernetes_config" "kubeconfig" {
  name        = "it-kubeconfig"
  description = "Integration test: Kubernetes config"
  content = jsonencode({
    apiVersion = "v1"
    kind       = "Config"
    clusters   = []
    users      = []
    contexts   = []
  })
}

resource "btpservice_cicd_credential_basic_auth_custom_idp" "cidp" {
  name        = "it-basic-auth-cidp"
  description = "Integration test: Basic auth for custom IdP"
  username    = var.deploy_username
  password    = var.deploy_password
  origin      = "integration-test-idp"
}

resource "btpservice_cicd_credential_cert_based_auth_custom_idp" "cert_cidp" {
  name          = "it-cert-cidp"
  description   = "Integration test: Cert-based auth for custom IdP"
  email_address = var.deploy_username
  hostname      = "integration-test.accounts.ondemand.com"
  origin        = "integration-test-idp_platform"
}

resource "btpservice_cicd_credential_secret_text" "api_token" {
  name        = "it-api-token"
  description = "Integration test: Secret text API token"
  text        = "integration-test-api-token-value"
}

resource "btpservice_cicd_credential_service_key" "service_key" {
  name        = "it-service-key"
  description = "Integration test: BTP service key"
  key = jsonencode({
    uri = "https://integration-test.cfapps.sap.hana.ondemand.com"
    uaa = {
      clientid     = "integration-test-client"
      clientsecret = "integration-test-secret"
      url          = "https://integration-test.authentication.sap.hana.ondemand.com"
    }
  })
}

# =============================================================================
# Step 2 — Repositories (public, private with clone cred, webhook-enabled)
# =============================================================================

resource "btpservice_cicd_repository" "public" {
  name      = "it-public-repo"
  clone_url = var.github_repo_url
}

resource "btpservice_cicd_repository" "private" {
  name                = "it-private-repo"
  clone_url           = var.github_repo_url
  clone_credential_id = btpservice_cicd_credential_basic_auth.deploy.id
}

resource "btpservice_cicd_repository" "webhook" {
  name      = "it-webhook-repo"
  clone_url = var.github_repo_url

  event_receiver = {
    active                      = true
    scm_type                    = "GITHUB"
    webhook_token_credential_id = btpservice_cicd_credential_webhook_secret.webhook.id
  }
}

# =============================================================================
# Step 3 — Jobs (CF env inline, source-repo, ANS) + Triggers
# =============================================================================

resource "btpservice_cicd_job" "cf_env" {
  name                 = "it-cf-env-job"
  description          = "Integration test: CF environment pipeline"
  repository_id        = btpservice_cicd_repository.public.id
  branch               = "main"
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  active               = true
  build_retention_days = 14
  max_builds_to_keep   = 5

  pipeline_parameters = <<-YAML
    configurationSource: job_parameter
    cfEnvConfiguration:
      stages:
        build:
          buildTool: mta
          buildToolVersion: MBTJ21N24
        release:
          cfDeploy:
            strategy: blue-green
            apiEndpoint: https://api.cf.us10.hana.ondemand.com
            org: integration-test-org
            space: integration-test-space
            credential: ${btpservice_cicd_credential_basic_auth.deploy.id}
  YAML
}

resource "btpservice_cicd_job" "source_repo" {
  name                 = "it-source-repo-job"
  description          = "Integration test: pipeline config read from source repo"
  repository_id        = btpservice_cicd_repository.webhook.id
  branch               = "main"
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  active               = true
  build_retention_days = 7
  max_builds_to_keep   = 3

  pipeline_parameters = <<-YAML
    configurationSource: source_repository
  YAML
}

resource "btpservice_cicd_job" "cf_with_ans" {
  name                 = "it-cf-ans-job"
  description          = "Integration test: CF pipeline with ANS notifications"
  repository_id        = btpservice_cicd_repository.public.id
  branch               = "main"
  pipeline             = "cf-env"
  pipeline_version     = "3.0"
  active               = true
  build_retention_days = 7
  max_builds_to_keep   = 3

  pipeline_parameters = <<-YAML
    configurationSource: source_repository
  YAML

  notification_configuration = {
    ans = {
      active        = true
      credential_id = btpservice_cicd_credential_service_key.service_key.id
      custom_tag    = "integration-test"
    }
  }
}

resource "btpservice_cicd_trigger" "nightly" {
  job  = btpservice_cicd_job.cf_env.name
  type = "timer"

  timer = {
    branch = "main"
    cron   = "0 2 * * *"
  }
}

resource "btpservice_cicd_trigger" "weekday_morning" {
  job  = btpservice_cicd_job.cf_env.name
  type = "timer"

  timer = {
    branch = "main"
    cron   = "0 9 * * 1-5"
  }

  depends_on = [btpservice_cicd_trigger.nightly]
}

# =============================================================================
# Step 4 — Data sources (read back everything created above)
# =============================================================================

data "btpservice_cicd_credential" "deploy" {
  name = btpservice_cicd_credential_basic_auth.deploy.name
}

data "btpservice_cicd_credentials" "all" {
  depends_on = [
    btpservice_cicd_credential_basic_auth.deploy,
    btpservice_cicd_credential_webhook_secret.webhook,
    btpservice_cicd_credential_cloud_connector.cc,
    btpservice_cicd_credential_container_registry.registry,
    btpservice_cicd_credential_kubernetes_config.kubeconfig,
    btpservice_cicd_credential_basic_auth_custom_idp.cidp,
    btpservice_cicd_credential_cert_based_auth_custom_idp.cert_cidp,
    btpservice_cicd_credential_secret_text.api_token,
    btpservice_cicd_credential_service_key.service_key,
  ]
}

data "btpservice_cicd_credential_usage" "deploy_all" {
  credential = btpservice_cicd_credential_basic_auth.deploy.name

  depends_on = [btpservice_cicd_job.cf_env, btpservice_cicd_repository.private]
}

data "btpservice_cicd_credential_usage" "deploy_jobs_only" {
  credential = btpservice_cicd_credential_basic_auth.deploy.name
  usertype   = "job"

  depends_on = [btpservice_cicd_job.cf_env]
}

data "btpservice_cicd_job_credentials" "cf_env" {
  job = btpservice_cicd_job.cf_env.name
}

data "btpservice_cicd_repository" "public" {
  name = btpservice_cicd_repository.public.name
}

data "btpservice_cicd_repositories" "all" {
  depends_on = [
    btpservice_cicd_repository.public,
    btpservice_cicd_repository.private,
    btpservice_cicd_repository.webhook,
  ]
}

data "btpservice_cicd_repository_jobs" "webhook_repo" {
  repository = btpservice_cicd_repository.webhook.name

  depends_on = [btpservice_cicd_job.source_repo]
}

data "btpservice_cicd_repository_event_receiver" "webhook" {
  repository = btpservice_cicd_repository.webhook.name
}

data "btpservice_cicd_repository_webhook_config" "webhook" {
  repository = btpservice_cicd_repository.webhook.name
}

data "btpservice_cicd_job" "cf_env" {
  name = btpservice_cicd_job.cf_env.name
}

data "btpservice_cicd_job" "source_repo_by_id" {
  id = btpservice_cicd_job.source_repo.id
}

data "btpservice_cicd_jobs" "all" {
  depends_on = [
    btpservice_cicd_job.cf_env,
    btpservice_cicd_job.source_repo,
    btpservice_cicd_job.cf_with_ans,
  ]
}

data "btpservice_cicd_trigger" "nightly" {
  job = btpservice_cicd_job.cf_env.name
  id  = btpservice_cicd_trigger.nightly.id
}

data "btpservice_cicd_triggers" "cf_env" {
  job = btpservice_cicd_job.cf_env.name

  depends_on = [btpservice_cicd_trigger.nightly, btpservice_cicd_trigger.weekday_morning]
}

# =============================================================================
# Outputs
# =============================================================================

output "credential_deploy_id" {
  value = btpservice_cicd_credential_basic_auth.deploy.id
}

output "credentials_total" {
  value = length(data.btpservice_cicd_credentials.all.values)
}

output "repository_public_id" {
  value = btpservice_cicd_repository.public.id
}

output "repository_webhook_receiver_id" {
  value = btpservice_cicd_repository.webhook.event_receiver.webhook_id
}

output "repository_webhook_delivery_url" {
  value = data.btpservice_cicd_repository_webhook_config.webhook.webhook_uri
}

output "job_cf_env_id" {
  value = btpservice_cicd_job.cf_env.id
}

output "job_cf_env_etag" {
  value = data.btpservice_cicd_job.cf_env.etag
}

output "jobs_total" {
  value = length(data.btpservice_cicd_jobs.all.values)
}

output "trigger_nightly_id" {
  value = btpservice_cicd_trigger.nightly.id
}

output "triggers_cf_env_total" {
  value = length(data.btpservice_cicd_triggers.cf_env.values)
}

output "deploy_credential_usage_count" {
  value = length(data.btpservice_cicd_credential_usage.deploy_all.usages)
}
