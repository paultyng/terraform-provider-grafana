resource "grafana_oncall_outgoing_webhook" "test-acc-outgoing_webhook" {
  provider = grafana.oncall
  name     = "my outgoing webhook"
  webhook  = "https://example.com/"
}