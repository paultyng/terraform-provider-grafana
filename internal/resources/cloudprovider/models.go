package cloudprovider

import (
	"context"
	"fmt"

	"github.com/grafana/terraform-provider-grafana/v3/internal/common/cloudproviderapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type awsCWScrapeJobTFModel struct {
	ID                   types.String `tfsdk:"id"`
	StackID              types.String `tfsdk:"stack_id"`
	Name                 types.String `tfsdk:"name"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	AWSAccountResourceID types.String `tfsdk:"aws_account_resource_id"`
	Regions              types.Set    `tfsdk:"regions"`
	ExportTags           types.Bool   `tfsdk:"export_tags"`
	DisabledReason       types.String `tfsdk:"disabled_reason"`
	// TODO(tristan): if the grafana provider is updated to use the Terraform v6 plugin protocol,
	// we can consider adding additional support to use Set Nested Attributes, instead of Blocks.
	// See https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes#nested-attribute-types
	Services         []awsCWScrapeJobServiceTFModel         `tfsdk:"service"`
	CustomNamespaces []awsCWScrapeJobCustomNamespaceTFModel `tfsdk:"custom_namespace"`
}
type awsCWScrapeJobServiceTFModel struct {
	Name                        types.String                     `tfsdk:"name"`
	Metrics                     []awsCWScrapeJobMetricTFModel    `tfsdk:"metric"`
	ScrapeIntervalSeconds       types.Int64                      `tfsdk:"scrape_interval_seconds"`
	ResourceDiscoveryTagFilters []awsCWScrapeJobTagFilterTFModel `tfsdk:"resource_discovery_tag_filter"`
	TagsToAddToMetrics          types.Set                        `tfsdk:"tags_to_add_to_metrics"`
}
type awsCWScrapeJobCustomNamespaceTFModel struct {
	Name                  types.String                  `tfsdk:"name"`
	Metrics               []awsCWScrapeJobMetricTFModel `tfsdk:"metric"`
	ScrapeIntervalSeconds types.Int64                   `tfsdk:"scrape_interval_seconds"`
}
type awsCWScrapeJobMetricTFModel struct {
	Name       types.String `tfsdk:"name"`
	Statistics types.Set    `tfsdk:"statistics"`
}
type awsCWScrapeJobTagFilterTFModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type awsCWScrapeJobNoDuplicateServiceNamesValidator struct{}

func (v awsCWScrapeJobNoDuplicateServiceNamesValidator) Description(ctx context.Context) string {
	return "No duplicate service names are allowed."
}

func (v awsCWScrapeJobNoDuplicateServiceNamesValidator) MarkdownDescription(ctx context.Context) string {
	return "No duplicate service names are allowed."
}

func (v awsCWScrapeJobNoDuplicateServiceNamesValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	seen := map[string]struct{}{}
	elems := make([]awsCWScrapeJobServiceTFModel, len(req.ConfigValue.Elements()))
	diags := req.ConfigValue.ElementsAs(ctx, &elems, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	for _, elem := range elems {
		name := elem.Name.ValueString()
		if _, ok := seen[name]; ok {
			resp.Diagnostics.AddError("Duplicate service name", fmt.Sprintf("Service name %q is duplicated.", name))
		}
		seen[name] = struct{}{}
	}
}

type awsCWScrapeJobNoDuplicateCustomNamespaceNamesValidator struct{}

func (v awsCWScrapeJobNoDuplicateCustomNamespaceNamesValidator) Description(ctx context.Context) string {
	return "No duplicate custom namespace names are allowed."
}

func (v awsCWScrapeJobNoDuplicateCustomNamespaceNamesValidator) MarkdownDescription(ctx context.Context) string {
	return "No duplicate custom namespace names are allowed."
}

func (v awsCWScrapeJobNoDuplicateCustomNamespaceNamesValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	seen := map[string]struct{}{}
	elems := make([]awsCWScrapeJobCustomNamespaceTFModel, len(req.ConfigValue.Elements()))
	diags := req.ConfigValue.ElementsAs(ctx, &elems, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	for _, elem := range elems {
		name := elem.Name.ValueString()
		if _, ok := seen[name]; ok {
			resp.Diagnostics.AddError("Duplicate custom namespace name", fmt.Sprintf("Custom namespace name %q is duplicated.", name))
		}
		seen[name] = struct{}{}
	}
}

type awsCWScrapeJobNoDuplicateMetricNamesValidator struct{}

func (v awsCWScrapeJobNoDuplicateMetricNamesValidator) Description(ctx context.Context) string {
	return "Metric names must be unique (case-insensitive) within the same service or custom namespace."
}

func (v awsCWScrapeJobNoDuplicateMetricNamesValidator) MarkdownDescription(ctx context.Context) string {
	return "Metric names must be unique (case-insensitive) within the same service or custom namespace."
}

func (v awsCWScrapeJobNoDuplicateMetricNamesValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	seen := map[string]struct{}{}
	elems := make([]awsCWScrapeJobMetricTFModel, len(req.ConfigValue.Elements()))
	diags := req.ConfigValue.ElementsAs(ctx, &elems, false)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	for _, elem := range elems {
		name := elem.Name.ValueString()
		if _, ok := seen[name]; ok {
			resp.Diagnostics.AddError("Duplicate metric name for service or custom namespace", fmt.Sprintf("Metric name %q is duplicated within the service or custom namespace.", name))
		}
		seen[name] = struct{}{}
	}
}

// convertScrapeJobTFModelToClientModel converts a awsCWScrapeJobTFModel instance to a cloudproviderapi.AWSCloudWatchScrapeJob instance.
// A special converter is needed because the TFModel uses special Terraform types that build upon their underlying Go types for
// supporting Terraform's state management/dependency analysis of the resource and its data.
func convertScrapeJobTFModelToClientModel(ctx context.Context, tfData awsCWScrapeJobTFModel) (*cloudproviderapi.AWSCloudWatchScrapeJob, diag.Diagnostics) {
	conversionDiags := diag.Diagnostics{}
	converted := cloudproviderapi.AWSCloudWatchScrapeJob{
		Name:                 tfData.Name.ValueString(),
		Enabled:              tfData.Enabled.ValueBool(),
		AWSAccountResourceID: tfData.AWSAccountResourceID.ValueString(),
		ExportTags:           tfData.ExportTags.ValueBool(),
	}

	diags := tfData.Regions.ElementsAs(ctx, &converted.Regions, false)
	conversionDiags.Append(diags...)
	if conversionDiags.HasError() {
		return nil, conversionDiags
	}

	converted.Services = make([]cloudproviderapi.AWSCloudWatchService, len(tfData.Services))
	for i, service := range tfData.Services {
		converted.Services[i] = cloudproviderapi.AWSCloudWatchService{
			Name:                  service.Name.ValueString(),
			ScrapeIntervalSeconds: service.ScrapeIntervalSeconds.ValueInt64(),
		}

		diags = service.TagsToAddToMetrics.ElementsAs(ctx, &converted.Services[i].TagsToAddToMetrics, false)
		conversionDiags.Append(diags...)
		if conversionDiags.HasError() {
			return nil, conversionDiags
		}

		converted.Services[i].Metrics = make([]cloudproviderapi.AWSCloudWatchMetric, len(service.Metrics))
		for j, metric := range service.Metrics {
			converted.Services[i].Metrics[j] = cloudproviderapi.AWSCloudWatchMetric{
				Name: metric.Name.ValueString(),
			}

			diags = metric.Statistics.ElementsAs(ctx, &converted.Services[i].Metrics[j].Statistics, false)
			conversionDiags.Append(diags...)
			if conversionDiags.HasError() {
				return nil, conversionDiags
			}
		}

		converted.Services[i].ResourceDiscoveryTagFilters = make([]cloudproviderapi.AWSCloudWatchTagFilter, len(service.ResourceDiscoveryTagFilters))
		for j, tagFilter := range service.ResourceDiscoveryTagFilters {
			converted.Services[i].ResourceDiscoveryTagFilters[j] = cloudproviderapi.AWSCloudWatchTagFilter{
				Key:   tagFilter.Key.ValueString(),
				Value: tagFilter.Value.ValueString(),
			}
		}
	}

	converted.CustomNamespaces = make([]cloudproviderapi.AWSCloudWatchCustomNamespace, len(tfData.CustomNamespaces))
	for i, customNamespace := range tfData.CustomNamespaces {
		converted.CustomNamespaces[i] = cloudproviderapi.AWSCloudWatchCustomNamespace{
			Name:                  customNamespace.Name.ValueString(),
			ScrapeIntervalSeconds: customNamespace.ScrapeIntervalSeconds.ValueInt64(),
		}

		converted.CustomNamespaces[i].Metrics = make([]cloudproviderapi.AWSCloudWatchMetric, len(customNamespace.Metrics))
		for j, metric := range customNamespace.Metrics {
			converted.CustomNamespaces[i].Metrics[j] = cloudproviderapi.AWSCloudWatchMetric{
				Name: metric.Name.ValueString(),
			}

			diags = metric.Statistics.ElementsAs(ctx, &converted.CustomNamespaces[i].Metrics[j].Statistics, false)
			conversionDiags.Append(diags...)
			if conversionDiags.HasError() {
				return nil, conversionDiags
			}
		}
	}

	return &converted, conversionDiags
}

// convertScrapeJobClientModelToTFModel converts a cloudproviderapi.AWSCloudWatchScrapeJob instance to a awsCWScrapeJobTFModel instance.
// A special converter is needed because the TFModel uses special Terraform types that build upon their underlying Go types for
// supporting Terraform's state management/dependency analysis of the resource and its data.
func convertScrapeJobClientModelToTFModel(ctx context.Context, stackID string, scrapeJobData cloudproviderapi.AWSCloudWatchScrapeJob) (*awsCWScrapeJobTFModel, diag.Diagnostics) {
	conversionDiags := diag.Diagnostics{}
	converted := &awsCWScrapeJobTFModel{
		ID:                   types.StringValue(resourceAWSCloudWatchScrapeJobTerraformID.Make(stackID, scrapeJobData.Name)),
		StackID:              types.StringValue(stackID),
		Name:                 types.StringValue(scrapeJobData.Name),
		Enabled:              types.BoolValue(scrapeJobData.Enabled),
		AWSAccountResourceID: types.StringValue(scrapeJobData.AWSAccountResourceID),
		ExportTags:           types.BoolValue(scrapeJobData.ExportTags),
		DisabledReason:       types.StringValue(scrapeJobData.DisabledReason),
	}

	regions, diags := types.SetValueFrom(ctx, basetypes.StringType{}, &scrapeJobData.Regions)
	conversionDiags.Append(diags...)
	if conversionDiags.HasError() {
		return nil, conversionDiags
	}
	converted.Regions = regions

	for i, serviceData := range scrapeJobData.Services {
		service := awsCWScrapeJobServiceTFModel{
			Name:                  types.StringValue(serviceData.Name),
			ScrapeIntervalSeconds: types.Int64Value(serviceData.ScrapeIntervalSeconds),
		}

		metricsData := make([]awsCWScrapeJobMetricTFModel, len(serviceData.Metrics))
		for j, metricData := range serviceData.Metrics {
			metricsData[j] = awsCWScrapeJobMetricTFModel{
				Name: types.StringValue(metricData.Name),
			}
			statistics, diags := types.SetValueFrom(ctx, basetypes.StringType{}, &scrapeJobData.Services[i].Metrics[j].Statistics)
			conversionDiags.Append(diags...)
			if conversionDiags.HasError() {
				return nil, conversionDiags
			}
			metricsData[j].Statistics = statistics
		}
		service.Metrics = metricsData

		tagFiltersData := make([]awsCWScrapeJobTagFilterTFModel, len(serviceData.ResourceDiscoveryTagFilters))
		for j, tagFilterData := range serviceData.ResourceDiscoveryTagFilters {
			tagFiltersData[j] = awsCWScrapeJobTagFilterTFModel{
				Key:   types.StringValue(tagFilterData.Key),
				Value: types.StringValue(tagFilterData.Value),
			}
		}
		service.ResourceDiscoveryTagFilters = tagFiltersData

		tagsToAdd, diags := types.SetValueFrom(ctx, basetypes.StringType{}, &scrapeJobData.Services[i].TagsToAddToMetrics)
		conversionDiags.Append(diags...)
		if conversionDiags.HasError() {
			return nil, conversionDiags
		}
		service.TagsToAddToMetrics = tagsToAdd

		converted.Services = append(converted.Services, service)
	}

	for i, customNamespaceData := range scrapeJobData.CustomNamespaces {
		customNamespace := awsCWScrapeJobCustomNamespaceTFModel{
			Name:                  types.StringValue(customNamespaceData.Name),
			ScrapeIntervalSeconds: types.Int64Value(customNamespaceData.ScrapeIntervalSeconds),
		}

		metricsData := make([]awsCWScrapeJobMetricTFModel, len(customNamespaceData.Metrics))
		for j, metricData := range customNamespaceData.Metrics {
			metricsData[j] = awsCWScrapeJobMetricTFModel{
				Name: types.StringValue(metricData.Name),
			}
			statistics, diags := types.SetValueFrom(ctx, basetypes.StringType{}, &scrapeJobData.CustomNamespaces[i].Metrics[j].Statistics)
			conversionDiags.Append(diags...)
			if conversionDiags.HasError() {
				return nil, conversionDiags
			}
			metricsData[j].Statistics = statistics
		}
		customNamespace.Metrics = metricsData

		converted.CustomNamespaces = append(converted.CustomNamespaces, customNamespace)
	}

	return converted, conversionDiags
}

// TestAWSCloudWatchscrapeJobData is only temporarily exported here until
// we have the resource handlers talking to the real API.
// TODO(tristan): move this to test package and unexport
// once we're using the actual API for interactions.
const TestStackID = "test-stack-id"

var TestAWSCloudWatchScrapeJobData = cloudproviderapi.AWSCloudWatchScrapeJob{
	Name:                 "test-scrape-job",
	Enabled:              true,
	AWSAccountResourceID: "1",
	Regions:              []string{"us-east-1", "us-east-2", "us-west-1"},
	ExportTags:           true,
	DisabledReason:       "",
	Services: []cloudproviderapi.AWSCloudWatchService{
		{
			Name: "AWS/EC2",
			Metrics: []cloudproviderapi.AWSCloudWatchMetric{
				{
					Name:       "CPUUtilization",
					Statistics: []string{"Average"},
				},
				{
					Name:       "StatusCheckFailed",
					Statistics: []string{"Maximum"},
				},
			},
			ResourceDiscoveryTagFilters: []cloudproviderapi.AWSCloudWatchTagFilter{
				{
					Key:   "k8s.io/cluster-autoscaler/enabled",
					Value: "true",
				},
			},
			TagsToAddToMetrics: []string{"eks:cluster-name"},
		},
	},
	CustomNamespaces: []cloudproviderapi.AWSCloudWatchCustomNamespace{
		{
			Name:                  "CoolApp",
			ScrapeIntervalSeconds: 300,
			Metrics: []cloudproviderapi.AWSCloudWatchMetric{
				{
					Name:       "CoolMetric",
					Statistics: []string{"Maximum", "Sum"},
				},
			},
		},
	},
}