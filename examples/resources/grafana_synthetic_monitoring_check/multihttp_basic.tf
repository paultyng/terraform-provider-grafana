data "grafana_synthetic_monitoring_probes" "main" {}

resource "grafana_synthetic_monitoring_check" "multihttp" {
  job     = "MultiHTTP defaults"
  target  = "https://www.grafana-dev.com"
  enabled = false
  probes = [
    data.grafana_synthetic_monitoring_probes.main.probes.Atlanta,
  ]
  labels = {
    foo = "bar"
  }
  settings {
    multihttp {
      entries {
        request {
          method = "GET"
          url    = "https://www.grafana-dev.com"
        }
      }
    }
  }
}