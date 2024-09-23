package connections

import (
	"github.com/grafana/terraform-provider-grafana/v3/internal/common/connectionsapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type metricsEndpointScrapeJobTFModel struct {
	ID                          types.String `tfsdk:"id"`
	StackID                     types.String `tfsdk:"stack_id"`
	Name                        types.String `tfsdk:"name"`
	Enabled                     types.Bool   `tfsdk:"enabled"`
	AuthenticationMethod        types.String `tfsdk:"authentication_method"`
	AuthenticationBearerToken   types.String `tfsdk:"authentication_bearer_token"`
	AuthenticationBasicUsername types.String `tfsdk:"authentication_basic_username"`
	AuthenticationBasicPassword types.String `tfsdk:"authentication_basic_password"`
	URL                         types.String `tfsdk:"url"`
	ScrapeIntervalSeconds       types.Int64  `tfsdk:"scrape_interval_seconds"`
}

// convertJobTFModelToClientModel converts a metricsEndpointScrapeJobTFModel instance to a connectionsapi.MetricsEndpointScrapeJob instance.
// A special converter is needed because the TFModel uses special Terraform types that build upon their underlying Go types for
// supporting Terraform's state management/dependency analysis of the resource and its data.
func convertJobTFModelToClientModel(tfData metricsEndpointScrapeJobTFModel) connectionsapi.MetricsEndpointScrapeJob {

	converted := connectionsapi.MetricsEndpointScrapeJob{
		Name:                        tfData.Name.ValueString(),
		Enabled:                     tfData.Enabled.ValueBool(),
		AuthenticationMethod:        tfData.AuthenticationMethod.ValueString(),
		AuthenticationBearerToken:   tfData.AuthenticationBearerToken.ValueString(),
		AuthenticationBasicUsername: tfData.AuthenticationBasicUsername.ValueString(),
		AuthenticationBasicPassword: tfData.AuthenticationBasicPassword.ValueString(),
		URL:                         tfData.URL.ValueString(),
		ScrapeIntervalSeconds:       tfData.ScrapeIntervalSeconds.ValueInt64(),
	}

	// TODO: I didn't think this needed to be a pointer
	return converted
}

// convertClientModelToTFModel converts a connectionsapi.MetricsEndpointScrapeJob instance to a metricsEndpointScrapeJobTFModel instance.
// A special converter is needed because the TFModel uses special Terraform types that build upon their underlying Go types for
// supporting Terraform's state management/dependency analysis of the resource and its data.
func convertClientModelToTFModel(stackID string, scrapeJobData connectionsapi.MetricsEndpointScrapeJob) *metricsEndpointScrapeJobTFModel {
	converted := &metricsEndpointScrapeJobTFModel{
		ID:                          types.StringValue(resourceMetricsEndpointScrapeJobTerraformID.Make(stackID, scrapeJobData.Name)),
		StackID:                     types.StringValue(stackID),
		Name:                        types.StringValue(scrapeJobData.Name),
		Enabled:                     types.BoolValue(scrapeJobData.Enabled),
		AuthenticationMethod:        types.StringValue(scrapeJobData.AuthenticationMethod),
		AuthenticationBearerToken:   types.StringValue(scrapeJobData.AuthenticationBearerToken),
		AuthenticationBasicUsername: types.StringValue(scrapeJobData.AuthenticationBasicUsername),
		AuthenticationBasicPassword: types.StringValue(scrapeJobData.AuthenticationBasicPassword),
		URL:                         types.StringValue(scrapeJobData.URL),
		ScrapeIntervalSeconds:       types.Int64Value(scrapeJobData.ScrapeIntervalSeconds),
	}

	return converted
}
