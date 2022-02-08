package grafana

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceDashboards() *schema.Resource {
	return &schema.Resource{
		Description: `
Datasource for retrieving all dashboards. Specify list of folder IDs to search in for dashboards.

* [Official documentation](https://grafana.com/docs/grafana/latest/dashboards/)
* [Folder/Dashboard Search HTTP API](https://grafana.com/docs/grafana/latest/http_api/folder_dashboard_search/)
* [Dashboard HTTP API](https://grafana.com/docs/grafana/latest/http_api/dashboard/)
`,
		ReadContext: dataSourceReadDashboards,
		Schema: map[string]*schema.Schema{
			"folder_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Numerical IDs of Grafana folders containing dashboards. Specify to filter for dashboards by folder (eg. `[0]` for General folder), or leave blank to get all dashboards in all folders.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of string Grafana dashboard tags to search for, eg. `[\"prod\"]`. Used only as search input, i.e., attribute value will remain unchanged.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"dashboards": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"title": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"uid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"folder_title": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceReadDashboards(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client).gapi
	var diags diag.Diagnostics
	params := map[string]string{
		"limit": "5000",
		"type":  "dash-db",
	}

	// add tags and folder IDs from attributes to dashboard search parameters
	resourceId := "dashboards"
	for thisParamKey, thisListAttributeName := range map[string]string{
		"folderIds": "folder_ids",
		"tag":       "tags",
	} {
		if list, ok := d.GetOk(thisListAttributeName); ok {
			listJSON, err := json.Marshal(list.([]interface{}))
			if err != nil {
				return diag.FromErr(err)
			}
			params[thisParamKey] = string(listJSON)
			resourceId += fmt.Sprintf("-%s", thisListAttributeName)
		}
	}

	results, err := client.FolderDashboardSearch(params)
	if err != nil {
		return diag.FromErr(err)
	}

	dashboards := make([]map[string]interface{}, len(results))
	for i, result := range results {
		dashboards[i] = map[string]interface{}{
			"title":        result.Title,
			"uid":          result.UID,
			"folder_title": result.FolderTitle,
		}
	}

	d.SetId(resourceId)
	if err := d.Set("dashboards", dashboards); err != nil {
		return diag.Errorf("error setting dashboards attribute: %s", err)
	}

	return diags
}
