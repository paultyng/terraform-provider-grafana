terraform {
  required_providers {
    grafana = {
      source = "registry.terraform.io/grafana/grafana"
    }
  }
}

provider "grafana" {
  url = "https://elainetest.grafana.net/"
}

resource "grafana_slo" "sample" {
  name        = "Terraform - Import Test Name"
  description = "Terraform - Import Test Description"
  query {
    freeformquery = "sum(rate(apiserver_request_total{code!=\"500\"}[$__rate_interval])) / sum(rate(apiserver_request_total[$__rate_interval]))"
    type          = "freeform"
  }
  objectives {
    objective_value  = 0.995
    objective_window = "30d"
  }
}

output "sample_slo" {
  value = grafana_slo.sample
}
