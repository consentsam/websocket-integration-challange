package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsaps"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsgrafana"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type MonitoringStackProps struct {
	awscdk.StackProps
	Config DeploymentConfig
}

func NewMonitoringStack(scope constructs.Construct, id string, props *MonitoringStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create Amazon Managed Service for Prometheus (AMP) Workspace
	prometheusWorkspace := awsaps.NewWorkspace(stack, jsii.String("PrometheusWorkspace"), &awsaps.WorkspaceProps{
		Alias: jsii.String("websocket-monitoring"),
		LoggingConfiguration: &awsaps.LoggingConfiguration{
			LogLevel: jsii.String("info"),
		},
	})

	// Create IAM role for Grafana
	grafanaRole := awsiam.NewRole(stack, jsii.String("GrafanaRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("grafana.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonPrometheusQueryAccess")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("CloudWatchReadOnlyAccess")),
		},
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"GrafanaWorkspacePolicy": awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
				Statements: &[]awsiam.PolicyStatement{
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Effect: awsiam.Effect_ALLOW,
						Actions: jsii.Strings(
							"aps:QueryMetrics",
							"aps:GetSeries",
							"aps:GetLabels",
							"aps:GetMetricMetadata",
						),
						Resources: &[]*string{prometheusWorkspace.WorkspaceArn()},
					}),
				},
			}),
		},
	})

	// Create Amazon Managed Grafana Workspace
	grafanaWorkspace := awsgrafana.NewWorkspace(stack, jsii.String("GrafanaWorkspace"), &awsgrafana.WorkspaceProps{
		AccountAccessType: awsgrafana.AccountAccessType_CURRENT_ACCOUNT,
		AuthenticationProviders: &[]awsgrafana.AuthenticationProviderTypes{
			awsgrafana.AuthenticationProviderTypes_AWS_SSO,
		},
		DataSources: &[]awsgrafana.DataSourceType{
			awsgrafana.DataSourceType_PROMETHEUS,
			awsgrafana.DataSourceType_CLOUDWATCH,
		},
		Description: jsii.String("Grafana workspace for WebSocket service monitoring"),
		Name: jsii.String("websocket-dashboards"),
		NotificationDestinations: &[]awsgrafana.NotificationDestinationType{
			awsgrafana.NotificationDestinationType_SNS,
		},
		OrganizationRoleName: jsii.String("ADMIN"),
		Role: grafanaRole,
	})

	// Output monitoring endpoints
	awscdk.NewCfnOutput(stack, jsii.String("PrometheusWorkspaceId"), &awscdk.CfnOutputProps{
		Value:       prometheusWorkspace.WorkspaceId(),
		Description: jsii.String("Amazon Managed Prometheus Workspace ID"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("PrometheusEndpoint"), &awscdk.CfnOutputProps{
		Value:       prometheusWorkspace.WorkspacePrometheusEndpoint(),
		Description: jsii.String("Prometheus Query Endpoint"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("GrafanaWorkspaceId"), &awscdk.CfnOutputProps{
		Value:       grafanaWorkspace.WorkspaceId(),
		Description: jsii.String("Amazon Managed Grafana Workspace ID"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("GrafanaURL"), &awscdk.CfnOutputProps{
		Value:       grafanaWorkspace.WorkspaceEndpoint(),
		Description: jsii.String("Grafana Dashboard URL"),
	})

	return stack
} 