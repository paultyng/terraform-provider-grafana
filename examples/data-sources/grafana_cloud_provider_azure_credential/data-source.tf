resource "grafana_cloud_provider_azure_credential" "test" {
  stack_id      = "1"
  name          = "test-name"
  client_id     = "my-client-id"
  client_secret = "my-client-secret"
  tenant_id     = "my-tenant-id"

  resource_tag_filter    {
    key   = "key-1"
    value = "value-1"
  }
  resource_tag_filter {
    key   = "key-2"
    value = "value-2"
  }
}


data "grafana_cloud_provider_azure_credential" "test" {
  stack_id      = grafana_cloud_provider_azure_credential.test.stack_id
  name          = grafana_cloud_provider_azure_credential.test.name
  client_id     = grafana_cloud_provider_azure_credential.test.client_id
  client_secret = grafana_cloud_provider_azure_credential.test.client_secret
  tenant_id     = grafana_cloud_provider_azure_credential.test.tenant_id
  resource_id   = grafana_cloud_provider_azure_credential.test.resource_id

  resource_tag_filter {
    key   = grafana_cloud_provider_azure_credential.test.resource_tag_filter[0].key
    value = grafana_cloud_provider_azure_credential.test.resource_tag_filter[0].value
  }

  resource_tag_filter {
      key   = grafana_cloud_provider_azure_credential.test.resource_tag_filter[1].key
      value = grafana_cloud_provider_azure_credential.test.resource_tag_filter[1].value
  }
}