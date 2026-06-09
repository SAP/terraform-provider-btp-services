# Abort a running build.
# The build must exist and be in a running state; otherwise the action returns an error.
resource "terraform_data" "abort_build" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.btpservice_cicd_abort_build.example]
    }
  }
}

action "btpservice_cicd_abort_build" "example" {
  config {
    job   = "my-pipeline-job"
    build = 42
  }
}
