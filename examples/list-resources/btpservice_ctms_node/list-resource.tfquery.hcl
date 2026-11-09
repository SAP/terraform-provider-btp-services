list "btpservice_ctms_node" "example" {
  provider         = btpservice
  include_resource = true

  config {
    name = "QUALITY"
  }
}
