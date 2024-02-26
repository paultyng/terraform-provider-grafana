package cloud

import (
	"context"
	"time"

	"github.com/grafana/grafana-com-public-clients/go/gcom"
	"github.com/grafana/terraform-provider-grafana/internal/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var ResourceAccessPolicyTokenID = common.NewTFIDWithLegacySeparator("grafana_cloud_access_policy_token", "/", "region", "tokenId")

func ResourceAccessPolicyToken() *schema.Resource {
	return &schema.Resource{

		Description: `
* [Official documentation](https://grafana.com/docs/grafana-cloud/account-management/authentication-and-permissions/access-policies/)
* [API documentation](https://grafana.com/docs/grafana-cloud/developer-resources/api-reference/cloud-api/#create-a-token)
`,

		CreateContext: CreateCloudAccessPolicyToken,
		UpdateContext: UpdateCloudAccessPolicyToken,
		DeleteContext: DeleteCloudAccessPolicyToken,
		ReadContext:   ReadCloudAccessPolicyToken,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the access policy for which to create a token.",
			},
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Region of the access policy. Should be set to the same region as the access policy. Use the region list API to get the list of available regions: https://grafana.com/docs/grafana-cloud/developer-resources/api-reference/cloud-api/#list-regions.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the access policy token.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Display name of the access policy token. Defaults to the name.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" && old == d.Get("name").(string) {
						return true
					}
					return false
				},
			},
			"expires_at": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "Expiration date of the access policy token. Does not expire by default.",
				ValidateFunc: validation.IsRFC3339Time,
			},

			// Computed
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the access policy token.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the access policy token.",
			},
		},
	}
}

func CreateCloudAccessPolicyToken(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*common.Client).GrafanaCloudAPI
	region := d.Get("region").(string)

	tokenInput := gcom.PostTokensRequest{
		AccessPolicyId: d.Get("access_policy_id").(string),
		Name:           d.Get("name").(string),
		DisplayName:    common.Ref(d.Get("display_name").(string)),
	}

	if v, ok := d.GetOk("expires_at"); ok {
		expiresAt, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		tokenInput.ExpiresAt = &expiresAt
	}

	req := client.TokensAPI.PostTokens(ctx).Region(region).XRequestId(ClientRequestID()).PostTokensRequest(tokenInput)
	result, _, err := req.Execute()
	if err != nil {
		return apiError(err)
	}

	d.SetId(ResourceAccessPolicyTokenID.Make(region, result.Id))
	d.Set("token", result.Token)

	return ReadCloudAccessPolicyToken(ctx, d, meta)
}

func UpdateCloudAccessPolicyToken(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*common.Client).GrafanaCloudAPI

	split, err := ResourceAccessPolicyTokenID.Split(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	region, id := split[0], split[1]

	displayName := d.Get("display_name").(string)
	if displayName == "" {
		displayName = d.Get("name").(string)
	}

	req := client.TokensAPI.PostToken(ctx, id).Region(region).XRequestId(ClientRequestID()).PostTokenRequest(gcom.PostTokenRequest{
		DisplayName: &displayName,
	})
	if _, _, err := req.Execute(); err != nil {
		return apiError(err)
	}

	return ReadCloudAccessPolicyToken(ctx, d, meta)
}

func ReadCloudAccessPolicyToken(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*common.Client).GrafanaCloudAPI

	split, err := ResourceAccessPolicyTokenID.Split(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	region, id := split[0], split[1]

	result, _, err := client.TokensAPI.GetToken(ctx, id).Region(region).Execute()
	if err, shouldReturn := common.CheckReadError("policy token", d, err); shouldReturn {
		return err
	}

	d.Set("access_policy_id", result.AccessPolicyId)
	d.Set("region", region)
	d.Set("name", result.Name)
	d.Set("display_name", result.DisplayName)
	d.Set("created_at", result.CreatedAt.Format(time.RFC3339))
	if result.ExpiresAt != nil {
		d.Set("expires_at", result.ExpiresAt.Format(time.RFC3339))
	}
	if result.UpdatedAt != nil {
		d.Set("updated_at", result.UpdatedAt.Format(time.RFC3339))
	}
	d.SetId(ResourceAccessPolicyTokenID.Make(region, result.Id))

	return nil
}

func DeleteCloudAccessPolicyToken(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*common.Client).GrafanaCloudAPI

	split, err := ResourceAccessPolicyTokenID.Split(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	region, id := split[0], split[1]

	_, _, err = client.TokensAPI.DeleteToken(ctx, id).Region(region).XRequestId(ClientRequestID()).Execute()
	return apiError(err)
}
