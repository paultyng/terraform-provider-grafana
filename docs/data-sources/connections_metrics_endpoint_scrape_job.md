---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "grafana_connections_metrics_endpoint_scrape_job Data Source - terraform-provider-grafana"
subcategory: "Connections"
description: |-
  
---

# grafana_connections_metrics_endpoint_scrape_job (Data Source)



## Example Usage

```terraform
data "grafana_connections_metrics_endpoint_scrape_job" "ds_test" {
  stack_id = "1"
  name     = "my-scrape-job"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the Metrics Endpoint Scrape Job. Part of the Terraform Resource ID.
- `stack_id` (String) The Stack ID of the Grafana Cloud instance. Part of the Terraform Resource ID.

### Read-Only

- `authentication_basic_password` (String, Sensitive) Password for basic authentication.
- `authentication_basic_username` (String) Username for basic authentication.
- `authentication_bearer_token` (String, Sensitive) Token for authentication bearer.
- `authentication_method` (String) Method to pass authentication credentials: basic or bearer.
- `enabled` (Boolean) Whether the metrics endpoint scrape job is enabled or not.
- `id` (String) The Terraform Resource ID. This has the format "{{ stack_id }}:{{ name }}".
- `scrape_interval_seconds` (Number) Frequency for scraping the metrics endpoint: 30, 60, or 120 seconds.
- `url` (String) The url to scrape metrics.
