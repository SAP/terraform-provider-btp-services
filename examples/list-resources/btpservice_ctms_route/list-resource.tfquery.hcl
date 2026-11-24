list "btpservice_ctms_route" "example" {
  provider         = btpservice
  include_resource = true

  config {
    name = "DEVELOPMENT_TO_QUALITY"
  }
}
