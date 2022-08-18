package grafana

import (
	"context"

	gapi "github.com/grafana/grafana-api-golang-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceNotificationPolicy() *schema.Resource {
	return &schema.Resource{
		Description: `
* [Official documentation](https://grafana.com/docs/grafana/latest/alerting/notifications/)
* [HTTP API](https://grafana.com/docs/grafana/next/developers/http_api/alerting_provisioning/#notification-policies)
`,

		CreateContext: createNotificationPolicy,
		ReadContext:   readNotificationPolicy,
		UpdateContext: updateNotificationPolicy,
		DeleteContext: deleteNotificationPolicy,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"contact_point": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The default contact point to route all unmatched notifications to.",
			},
			"group_by": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "A list of alert labels to group alerts into notifications by. Use the special label `...` to group alerts by all labels, effectively disabling grouping.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"group_wait": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Time to wait to buffer alerts of the same group before sending a notification. Default is 30 seconds.",
			},
			"group_interval": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Minimum time interval between two notifications for the same group. Default is 5 minutes.",
			},
			"repeat_interval": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Minimum time interval for re-sending a notification if an alert is still firing. Default is 4 hours.",
			},

			"policy": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Routing rules for specific label sets.",
				Elem:        policySchema(supportedPolicyTreeDepth),
			},
		},
	}
}

// The maximum depth of policy tree that the provider supports, as Terraform does not allow for infinitely recursive schemas.
// This can be increased without breaking backwards compatibility.
const supportedPolicyTreeDepth = 1

const PolicySingletonID = "policy"

func policySchema(depth uint) *schema.Resource {
	if depth == 0 {
		panic("there is no valid Terraform schema for a policy tree with depth 0")
	}

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"contact_point": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The contact point to route notifications that match this rule to.",
			},
			"group_by": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "A list of alert labels to group alerts into notifications by. Use the special label `...` to group alerts by all labels, effectively disabling grouping.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"matcher": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Describes which labels this rule should match. When multiple matchers are supplied, an alert must match ALL matchers to be accepted by this policy.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the label to match against.",
						},
						"match": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The operator to apply when matching values of the given label. Allowed operators are `=` for equality, `!=` for negated equality, `=~` for regex equality, and `!~` for negated regex equality.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The label value to match against.",
						},
					},
				},
			},
			"mute_timings": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of mute timing names to apply to alerts that match this policy.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"continue": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to continue matching subsequent rules if an alert matches the current rule. Otherwise, the rule will be 'consumed' by the first policy to match it.",
			},
			"group_interval": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Minimum time interval between two notifications for the same group. Default is 5 minutes.",
			},
			"repeat_interval": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Minimum time interval for re-sending a notification if an alert is still firing. Default is 4 hours.",
			},
		},
	}
}

func readNotificationPolicy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client).gapi

	npt, err := client.NotificationPolicyTree()
	if err != nil {
		return diag.FromErr(err)
	}

	packNotifPolicy(npt, data)
	data.SetId(PolicySingletonID)
	return nil
}

func createNotificationPolicy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client).gapi

	npt := unpackNotifPolicy(data)

	if err := client.SetNotificationPolicyTree(&npt); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(PolicySingletonID)
	return readNotificationPolicy(ctx, data, meta)
}

func updateNotificationPolicy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client).gapi

	npt := unpackNotifPolicy(data)

	if err := client.SetNotificationPolicyTree(&npt); err != nil {
		return diag.FromErr(err)
	}

	return readNotificationPolicy(ctx, data, meta)
}

func deleteNotificationPolicy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client).gapi

	if err := client.ResetNotificationPolicyTree(); err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}

func packNotifPolicy(npt gapi.NotificationPolicyTree, data *schema.ResourceData) {
	data.Set("contact_point", npt.Receiver)
	data.Set("group_by", npt.GroupBy)
	data.Set("group_wait", npt.GroupWait)
	data.Set("group_interval", npt.GroupInterval)
	data.Set("repeat_interval", npt.RepeatInterval)
}

func unpackNotifPolicy(data *schema.ResourceData) gapi.NotificationPolicyTree {
	groupBy := data.Get("group_by").([]interface{})
	groups := make([]string, 0, len(groupBy))
	for _, g := range groupBy {
		groups = append(groups, g.(string))
	}
	return gapi.NotificationPolicyTree{
		Receiver:       data.Get("contact_point").(string),
		GroupBy:        groups,
		GroupWait:      data.Get("group_wait").(string),
		GroupInterval:  data.Get("group_interval").(string),
		RepeatInterval: data.Get("repeat_interval").(string),
	}
}
