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

	// Create VPC with minimal configuration
	vpc := awsec2.NewVpc(stack, jsii.String("WebSocketVPC"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2),
		NatGateways: jsii.Number(1),
	})

	// Create ECR Repository
	repository := awsecr.NewRepository(stack, jsii.String("WebSocketRepository"), &awsecr.RepositoryProps{
		RepositoryName: jsii.String("websocket-service"),
		LifecycleRules: &[]*awsecr.LifecycleRule{
			{
				MaxImageCount: jsii.Number(5),
			},
		},
	})

	// Create ECS Cluster
	cluster := awsecs.NewCluster(stack, jsii.String("WebSocketCluster"), &awsecs.ClusterProps{
		Vpc: vpc,
		ClusterName: jsii.String("websocket-cluster"),
	})

	// Create Log Group
	logGroup := awslogs.NewLogGroup(stack, jsii.String("WebSocketLogGroup"), &awslogs.LogGroupProps{
		LogGroupName: jsii.String("/aws/ecs/websocket-service"),
		Retention: awslogs.RetentionDays_ONE_WEEK,
	})

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

	// Output important values (HTTP URLs for MVP)
	awscdk.NewCfnOutput(stack, jsii.String("ServiceURL"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(jsii.String(""), &[]*string{
			jsii.String("http://"),
			fargateService.LoadBalancer().LoadBalancerDnsName(),
		}),
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

	awscdk.NewCfnOutput(stack, jsii.String("MetricsURL"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(jsii.String(""), &[]*string{
			jsii.String("http://"),
			fargateService.LoadBalancer().LoadBalancerDnsName(),
			jsii.String("/metrics"),
		}),
		Description: jsii.String("Prometheus Metrics Endpoint"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ECRRepository"), &awscdk.CfnOutputProps{
		Value: repository.RepositoryUri(),
		Description: jsii.String("ECR Repository URI"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	// Read configuration - when run from cmd/deploy, cdk.json is two levels up
	configData, err := os.ReadFile("../../cdk.json")
	if err != nil {
		fmt.Printf("Error reading cdk.json: %v\n", err)
		os.Exit(1)
	}

	var cdkConfig struct {
		DeploymentConfig DeploymentConfig `json:"deploymentConfig"`
	}

	if err := json.Unmarshal(configData, &cdkConfig); err != nil {
		fmt.Printf("Error parsing cdk.json: %v\n", err)
		os.Exit(1)
	}

	app := awscdk.NewApp(nil)

	// Create main service stack only for MVP
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