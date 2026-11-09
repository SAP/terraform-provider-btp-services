resource "btpservice_ctms_node" "source" {
  name           = "DEVELOPMENT"
  description    = "Source node for the delivery pipeline"
  upload_allowed = true
  forward_mode   = "MANUAL"

  lifecycle {
    create_before_destroy = true
  }
}

resource "btpservice_ctms_node" "target" {
  name            = "QUALITY"
  import_disabled = true
  auto_import     = false

  targets = [{
    content_type     = "MTA"
    destination_name = "QUALITY_CF"
    import_options = {
      strategy = "default"
    }
  }]
}
