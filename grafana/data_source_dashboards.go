package grafana

import (
	"context"
	"encoding/json"

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
				Computed:    true,
				Description: `Numerical IDs of Grafana folders containing dashboards. Specify to filter for dashboards by folder (eg. [0] for General folder), or leave blank to get all dashboards in all folders.`,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: `List of string Grafana dashboard tags to search for, eg. ["prod"]. Used only as search input, i.e., attribute value will remain unchanged.`,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"dashboards": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Map of Grafana dashboard unique identifiers (list of string UIDs as values) to folder IDs (integers as keys).",
				Elem: &schema.Schema{
					Type:        schema.TypeList,
					Description: "List of string dashboard UIDs.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func dataSourceReadDashboards(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client).gapi
	params := map[string]string{
		"limit": "5000",
		"type":  "dash-db",
	}

	// add tags and folder IDs from attributes to dashboard search parameters
	for thisParamKey, thisListAttributeName := range map[string]string{
		"folderIds": "folder_ids",
		"tag":       "tags",
	} {
		list := d.Get(thisListAttributeName).([]interface{})
		if len(list) > 0 {
			listJSON, err := json.Marshal(list)
			if err != nil {
				return diag.FromErr(err)
			}
			params[thisParamKey] = string(listJSON)
		}
	}

	results, err := client.FolderDashboardSearch(params)
	if err != nil {
		return diag.FromErr(err)
	}

	// make list of string dashboard UIDs (as values) mapped to each int folder ID (as keys)
	dashboards := make(map[int64][]string, len(results))
	for _, thisResult := range results {
		thisFolderID := int64(thisResult.FolderID)
		dashboards[thisFolderID] = append(dashboards[thisFolderID], thisResult.UID)
	}

	// only set folder_ids if user did not specify, as re-ordered list may cause diff
	if _, ok := d.GetOk("folder_ids"); !ok {
		var folders []int64
		for thisFolderID := range dashboards {
			folders = append(folders, thisFolderID)
		}
		d.Set("folder_ids", folders)
	}
	d.Set("dashboards", dashboards)
	d.SetId("dashboards")

	return nil
}
