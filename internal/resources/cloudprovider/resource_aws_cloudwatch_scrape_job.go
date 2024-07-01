package cloudprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/grafana/terraform-provider-grafana/v3/internal/common"
	"github.com/grafana/terraform-provider-grafana/v3/internal/common/cloudproviderapi"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	resourceAWSCloudWatchScrapeJobTerraformName = "grafana_cloud_provider_aws_cloudwatch_scrape_job"
	resourceAWSCloudWatchScrapeJobTerraformID   = common.NewResourceID(common.StringIDField("stack_id"), common.StringIDField("job_name"))
)

type resourceAWSCloudWatchScrapeJob struct {
	client *cloudproviderapi.Client
}

func makeResourceAWSCloudWatchScrapeJob() *common.Resource {
	return common.NewResource(
		common.CategoryCloudProvider,
		"grafana_cloud_provider_aws_cloudwatch_scrape_job",
		resourceAWSCloudWatchScrapeJobTerraformID,
		&resourceAWSCloudWatchScrapeJob{},
	)
}

func (r *resourceAWSCloudWatchScrapeJob) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Configure is called multiple times (sometimes when ProviderData is not yet available), we only want to configure once
	if req.ProviderData == nil || r.client != nil {
		return
	}

	client, err := withClientForResource(req, resp)
	if err != nil {
		return
	}

	r.client = client
}

func (r *resourceAWSCloudWatchScrapeJob) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = resourceAWSCloudWatchScrapeJobTerraformName
}

func (r *resourceAWSCloudWatchScrapeJob) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The Terraform Resource ID. This has the format \"{{ stack_id }}:{{ job_name }}\".",
				Computed:    true,
			},
			"stack_id": schema.StringAttribute{
				Description: "The Stack ID of the Grafana Cloud instance. Part of the Terraform Resource ID.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the CloudWatch Scrape Job. Part of the Terraform Resource ID.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the CloudWatch Scrape Job is enabled or not.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"aws_account_resource_id": schema.StringAttribute{
				Description: "The ID assigned by the Grafana Cloud Provider API to an AWS Account resource that should be associated with this CloudWatch Scrape Job.",
				Required:    true,
			},
			"regions": schema.SetAttribute{
				Description: "A set of AWS region names that this CloudWatch Scrape Job applies to.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"service_configuration": schema.ListNestedBlock{
				Description: "One or more configuration blocks to dictate what this CloudWatch Scrape Job should scrape. Each block must have a distinct `name` attribute. When accessing this as an attribute reference, it is a list of objects.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					awsCWScrapeJobNoDuplicateServiceConfigNamesValidator{},
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the service to scrape. See https://grafana.com/docs/grafana-cloud/monitor-infrastructure/aws/cloudwatch-metrics/services/ for supported services, metrics, and their statistics.",
							Required:    true,
						},
						"scrape_interval_seconds": schema.Int64Attribute{
							Description: "The interval in seconds to scrape the service. See https://grafana.com/docs/grafana-cloud/monitor-infrastructure/aws/cloudwatch-metrics/services/ for supported scrape intervals.",
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(300),
						},
						"tags_to_add_to_metrics": schema.SetAttribute{
							Description: "A set of tags to add to all metrics exported by this scrape job, for use in PromQL queries.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"is_custom_namespace": schema.BoolAttribute{
							Description: "Whether the service name is a custom, user-generated metrics namespace, as opposed to a standard AWS service metrics namespace.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
					Blocks: map[string]schema.Block{
						"metric": schema.ListNestedBlock{
							Description: "One or more configuration blocks to configure metrics and their statistics to scrape. Each block must represent a distinct metric name. When accessing this as an attribute reference, it is a list of objects.",
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								awsCWScrapeJobNoDuplicateMetricNamesValidator{},
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The name of the metric to scrape.",
										Required:    true,
									},
									"statistics": schema.SetAttribute{
										Description: "A set of statistics to scrape.",
										Required:    true,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
										ElementType: types.StringType,
									},
								},
							},
						},
						"resource_discovery_tag_filter": schema.ListNestedBlock{
							Description: "One or more configuration blocks to configure tag filters applied to discovery of resource entities in the associated AWS account. When accessing this as an attribute reference, it is a list of objects.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Description: "The key of the tag filter.",
										Required:    true,
									},
									"value": schema.StringAttribute{
										Description: "The value of the tag filter.",
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceAWSCloudWatchScrapeJob) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Invalid ID: %s", req.ID))
		return
	}
	stackID := parts[0]
	jobName := parts[1]
	// TODO(tristan): use client to get AWS account so we only import a resource that exists
	resp.State.Set(ctx, &awsCWScrapeJobTFModel{
		ID:      types.StringValue(req.ID),
		StackID: types.StringValue(stackID),
		Name:    types.StringValue(jobName),
	})
}

func (r *resourceAWSCloudWatchScrapeJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data awsCWScrapeJobTFModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, &awsCWScrapeJobTFModel{
		ID:                         types.StringValue(resourceAWSCloudWatchScrapeJobTerraformID.Make(data.StackID.ValueString(), data.Name.ValueString())),
		StackID:                    data.StackID,
		Name:                       data.Name,
		Enabled:                    data.Enabled,
		AWSAccountResourceID:       data.AWSAccountResourceID,
		Regions:                    data.Regions,
		ServiceConfigurationBlocks: data.ServiceConfigurationBlocks,
	})
}

func (r *resourceAWSCloudWatchScrapeJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data awsCWScrapeJobTFModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, &awsCWScrapeJobTFModel{
		ID:                         types.StringValue(resourceAWSCloudWatchScrapeJobTerraformID.Make(data.StackID.ValueString(), data.Name.ValueString())),
		StackID:                    data.StackID,
		Name:                       data.Name,
		Enabled:                    data.Enabled,
		AWSAccountResourceID:       data.AWSAccountResourceID,
		Regions:                    data.Regions,
		ServiceConfigurationBlocks: data.ServiceConfigurationBlocks,
	})
}

func (r *resourceAWSCloudWatchScrapeJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var stateData awsCWScrapeJobTFModel
	diags := req.State.Get(ctx, &stateData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var configData awsCWScrapeJobTFModel
	diags = req.Config.Get(ctx, &configData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, &awsCWScrapeJobTFModel{
		ID:                         types.StringValue(resourceAWSCloudWatchScrapeJobTerraformID.Make(stateData.StackID.ValueString(), configData.Name.ValueString())),
		StackID:                    stateData.StackID,
		Name:                       configData.Name,
		Enabled:                    configData.Enabled,
		AWSAccountResourceID:       configData.AWSAccountResourceID,
		Regions:                    configData.Regions,
		ServiceConfigurationBlocks: configData.ServiceConfigurationBlocks,
	})
}

func (r *resourceAWSCloudWatchScrapeJob) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data awsCWScrapeJobTFModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, nil)
}