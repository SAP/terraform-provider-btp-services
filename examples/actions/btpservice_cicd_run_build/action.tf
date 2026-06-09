# Minimal — trigger a build with no guards.
# The job runs on whatever branch and commit is configured in the job definition.
resource "terraform_data" "trigger_build" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.btpservice_cicd_run_build.minimal]
    }
  }
}

action "btpservice_cicd_run_build" "minimal" {
  config {
    job = "my-pipeline-job"
  }
}

# ---------------------------------------------------------------------------
# Full — use the job's ETag to guard against triggering a stale configuration,
# specify the commit, and pass runtime parameter overrides.
# ---------------------------------------------------------------------------
data "btpservice_cicd_job" "pipeline" {
  name = "my-pipeline-job"
}

resource "terraform_data" "trigger_build_full" {
  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.btpservice_cicd_run_build.full]
    }
  }
}

action "btpservice_cicd_run_build" "full" {
  config {
    job                = data.btpservice_cicd_job.pipeline.name
    job_etag           = data.btpservice_cicd_job.pipeline.etag
    commit_to_be_built = "main"

    parameters = [
      {
        name       = "addon.yml"
        value      = "enabled: true\nversion: 1.2.3\n"
        visibility = "RESTRICTED"
      }
    ]
  }
}
