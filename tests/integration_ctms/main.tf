# =============================================================================
# Step 1 — All Nodes
# =============================================================================
resource "btpservice_ctms_node" "nodes_dev_int" {
  provider             = btpservice
  auto_import          = false
  description          = "Development Space Integration"
  forward_mode         = "AUTO"
  name                 = "terraform_dev_int"
  notification_enabled = true
  targets = [
    {
      content_type     = "MTA"
      destination_name = "ctms-deployer-dev"
      import_options = {
        strategy = "default"
      }
    },
  ]
  upload_allowed = true
  virtual        = false
}

resource "btpservice_ctms_node" "nodes_test_int" {
  provider             = btpservice
  auto_import          = false
  description          = "Test Space Integration"
  forward_mode         = "MANUAL"
  name                 = "terraform_test_int"
  notification_enabled = true
  targets = [
    {
      content_type     = "MTA"
      destination_name = "ctms-deployer-test"
      import_options = {
        strategy = "default"
      }
    },
  ]
  upload_allowed = true
  virtual        = false
}

resource "btpservice_ctms_node" "nodes_prod_int" {
  provider             = btpservice
  auto_import          = false
  description          = "Production Space Integration"
  forward_mode         = "AUTO"
  name                 = "terraform_prod_int"
  notification_enabled = true
  targets = [
    {
      content_type     = "MTA"
      destination_name = "ctms-deployer-prod"
      import_options = {
        strategy = "default"
      }
    },
  ]
  upload_allowed = false
  virtual        = false
}

# =============================================================================
# Step 1 — All Routes
# =============================================================================
resource "btpservice_ctms_route" "routes_dev_test_int" {
  provider       = btpservice
  description    = "Route from Development to Test Integration"
  name           = "terraform_dev_to_test_int"
  source_node_id = btpservice_ctms_node.nodes_dev_int.id
  target_node_id = btpservice_ctms_node.nodes_test_int.id
}

resource "btpservice_ctms_route" "routes_test_prod_int" {
  provider       = btpservice
  description    = "Route from Test to Production Integration"
  name           = "terraform_test_to_prod_int"
  source_node_id = btpservice_ctms_node.nodes_test_int.id
  target_node_id = btpservice_ctms_node.nodes_prod_int.id
}
