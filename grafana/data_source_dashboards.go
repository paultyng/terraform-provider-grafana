package grafana

import (
	"context"
	"encoding/json"
	"math/rand"

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
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Map of Grafana dashboard unique identifiers (list of string UIDs as values) to folder UIDs (strings as keys), eg. `{\"folderuid1\" = [\"cIBgcSjkk\"]}`.",
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
	var diags diag.Diagnostics
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

	// make list of string dashboard UIDs (as values) mapped to each string folder UID (as keys)
	dashboards := make(map[string][]string, len(results))
	for _, result := range results {
		dashboards[result.FolderUID] = append(dashboards[result.FolderUID], result.UID)
	}

	d.SetId(RandomString(12))
	d.Set("dashboards", dashboards)

	return diags
}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
