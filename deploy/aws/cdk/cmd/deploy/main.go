package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DeploymentConfig struct {
	Region          string `json:"region"`
	Account         string `json:"account"`
	DomainName      string `json:"domainName"`
	SubdomainPrefix string `json:"subdomainPrefix"`
	AlertEmail      string `json:"alertEmail"`
	Environment     string `json:"environment"`
}

type WebSocketServiceStackProps struct {
	awscdk.StackProps
	Config DeploymentConfig
}

func NewWebSocketServiceStack(scope constructs.Construct, id string, props *WebSocketServiceStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create VPC
	vpc := awsec2.NewVpc(stack, jsii.String("WebSocketVPC"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2),
		NatGateways: jsii.Number(1),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// Create ECR Repository
	repository := awsecr.NewRepository(stack, jsii.String("WebSocketRepository"), &awsecr.RepositoryProps{
		RepositoryName: jsii.String("websocket-service"),
		LifecycleRules: &[]*awsecr.LifecycleRule{
			{
				MaxImageCount: jsii.Number(10),
				TagPrefixList: jsii.Strings("latest"),
			},
		},
	})

	// Create ECS Cluster
	cluster := awsecs.NewCluster(stack, jsii.String("WebSocketCluster"), &awsecs.ClusterProps{
		Vpc: vpc,
		ClusterName: jsii.String("websocket-cluster"),
		ContainerInsights: jsii.Bool(true),
	})

	// Create Log Group
	logGroup := awslogs.NewLogGroup(stack, jsii.String("WebSocketLogGroup"), &awslogs.LogGroupProps{
		LogGroupName: jsii.String("/aws/ecs/websocket-service"),
		Retention: awslogs.RetentionDays_ONE_WEEK,
	})

	// Note: Simplified deployment without SSL for MVP

	// Create Fargate Service with Application Load Balancer (HTTP only for MVP)
	fargateService := awsecspatterns.NewApplicationLoadBalancedFargateService(stack, jsii.String("WebSocketService"), &awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
		Cluster: cluster,
		Cpu: jsii.Number(512),
		MemoryLimitMiB: jsii.Number(1024),
		DesiredCount: jsii.Number(1), // Start with 1 container for MVP
		TaskImageOptions: &awsecspatterns.ApplicationLoadBalancedTaskImageOptions{
			Image: awsecs.ContainerImage_FromEcrRepository(repository, jsii.String("latest")),
			ContainerPort: jsii.Number(8080),
			LogDriver: awsecs.LogDrivers_AwsLogs(&awsecs.AwsLogDriverProps{
				StreamPrefix: jsii.String("websocket"),
				LogGroup: logGroup,
			}),
			Environment: &map[string]*string{
				"ENVIRONMENT": jsii.String(props.Config.Environment),
				"AWS_REGION": jsii.String(props.Config.Region),
			},
		},
		PublicLoadBalancer: jsii.Bool(true),
		Protocol: awselasticloadbalancingv2.ApplicationProtocol_HTTP, // HTTP only for MVP
	})

	// Configure health check
	fargateService.TargetGroup().ConfigureHealthCheck(&awselasticloadbalancingv2.HealthCheck{
		Path: jsii.String("/health"),
		HealthyThresholdCount: jsii.Number(2),
		UnhealthyThresholdCount: jsii.Number(5),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Interval: awscdk.Duration_Seconds(jsii.Number(30)),
	})

	// Note: Auto-scaling disabled for MVP deployment

	// Create SNS Topic for Alarms
	alertTopic := awssns.NewTopic(stack, jsii.String("WebSocketAlerts"), &awssns.TopicProps{
		DisplayName: jsii.String("WebSocket Service Alerts"),
	})

	// Subscribe email to SNS topic
	awssns.NewSubscription(stack, jsii.String("EmailSubscription"), &awssns.SubscriptionProps{
		Topic: alertTopic,
		Protocol: awssns.SubscriptionProtocol_EMAIL,
		Endpoint: jsii.String(props.Config.AlertEmail),
	})

	// Create CloudWatch Alarms
	awscloudwatch.NewAlarm(stack, jsii.String("HighCpuAlarm"), &awscloudwatch.AlarmProps{
		AlarmName: jsii.String("WebSocket-High-CPU"),
		Metric: fargateService.Service().MetricCpuUtilization(&awsecs.ServiceMetricOptions{}),
		Threshold: jsii.Number(80),
		EvaluationPeriods: jsii.Number(2),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
		AlarmDescription: jsii.String("High CPU utilization in WebSocket service"),
	})

	awscloudwatch.NewAlarm(stack, jsii.String("HighMemoryAlarm"), &awscloudwatch.AlarmProps{
		AlarmName: jsii.String("WebSocket-High-Memory"),
		Metric: fargateService.Service().MetricMemoryUtilization(&awsecs.ServiceMetricOptions{}),
		Threshold: jsii.Number(80),
		EvaluationPeriods: jsii.Number(2),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
		AlarmDescription: jsii.String("High memory utilization in WebSocket service"),
	})

	// Output important values (HTTP URLs for MVP)
	awscdk.NewCfnOutput(stack, jsii.String("ServiceURL"), &awscdk.CfnOutputProps{
		Value: fargateService.LoadBalancer().LoadBalancerDnsName(),
		Description: jsii.String("WebSocket Service URL (HTTP)"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("WebSocketURL"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(jsii.String(""), &[]*string{
			jsii.String("ws://"),
			fargateService.LoadBalancer().LoadBalancerDnsName(),
			jsii.String("/ws"),
		}),
		Description: jsii.String("WebSocket Connection URL (WS)"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ECRRepository"), &awscdk.CfnOutputProps{
		Value: repository.RepositoryUri(),
		Description: jsii.String("ECR Repository URI"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	// Read configuration
	configData, err := os.ReadFile("../cdk.json")
	if err != nil {
		fmt.Printf("Error reading cdk.json: %v\n", os.Stderr)
		os.Exit(1)
	}

	var cdkConfig struct {
		DeploymentConfig DeploymentConfig `json:"deploymentConfig"`
	}

	if err := json.Unmarshal(configData, &cdkConfig); err != nil {
		fmt.Printf("Error parsing cdk.json: %v\n", os.Stderr)
		os.Exit(1)
	}

	app := awscdk.NewApp(nil)

	// Create monitoring stack first
	NewMonitoringStack(app, "WebSocketMonitoringStack", &MonitoringStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(cdkConfig.DeploymentConfig.Account),
				Region:  jsii.String(cdkConfig.DeploymentConfig.Region),
			},
		},
		Config: cdkConfig.DeploymentConfig,
	})

	// Create main service stack
	NewWebSocketServiceStack(app, "WebSocketServiceStack", &WebSocketServiceStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(cdkConfig.DeploymentConfig.Account),
				Region:  jsii.String(cdkConfig.DeploymentConfig.Region),
			},
		},
		Config: cdkConfig.DeploymentConfig,
	})

	app.Synth(nil)
} 