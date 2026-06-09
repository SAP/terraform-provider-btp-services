# Delete a completed (or failed) build and its logs.
# The build must exist; otherwise the action returns an error.
resource "terraform_data" "delete_build" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.btpservice_cicd_delete_build.example]
    }
  }
}

action "btpservice_cicd_delete_build" "example" {
  config {
    job   = "my-pipeline-job"
    build = 7
  }
}
