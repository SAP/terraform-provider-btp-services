resource "btpservice_ctms_node" "source" {
  name    = "DEVELOPMENT"
  virtual = true
}

resource "btpservice_ctms_node" "target" {
  name    = "QUALITY"
  virtual = true
}

resource "btpservice_ctms_route" "development_to_quality" {
  name           = "DEVELOPMENT_TO_QUALITY"
  source_node_id = btpservice_ctms_node.source.id
  target_node_id = btpservice_ctms_node.target.id
}
