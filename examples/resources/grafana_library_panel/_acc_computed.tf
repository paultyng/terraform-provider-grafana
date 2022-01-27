# Creating a library panel with a random uid.
# We'd like to ensure that using a computed configuration works.

resource "grafana_library_panel" "test" {
  name          = "computed"
  folder_id     = 0
  model_json    = jsonencode({
    title       = "computed"
    type        = "dash-db",
    id          = 12,
    version     = 35
  })
}

resource "grafana_library_panel" "test-computed" {
  name          = "computed-uid"
  folder_id     = 0
  model_json    = jsonencode({
    title       = "computed-uid"
    description = "test computed UID",
    tags        = ["${grafana_library_panel.test.uid}"],
    type        = "dash-db",
  })
}
