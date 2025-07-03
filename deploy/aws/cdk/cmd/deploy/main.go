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
	Region      string `json:"region"`
	Account     string `json:"account"`
	Environment string `json:"environment"`
	AlertEmail  string `json:"alertEmail"`
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

	// Create VPC with proper configuration for websockets
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

	// Create ECR Repository for the websocket service
	repository := awsecr.NewRepository(stack, jsii.String("WebSocketRepository"), &awsecr.RepositoryProps{
		RepositoryName: jsii.String("websocket-service"),
		LifecycleRules: &[]*awsecr.LifecycleRule{
			{
				MaxImageCount: jsii.Number(10),
				TagPrefixList: jsii.Strings("latest"),
			},
		},
	})

	// Create ECS Cluster with container insights
	cluster := awsecs.NewCluster(stack, jsii.String("WebSocketCluster"), &awsecs.ClusterProps{
		Vpc: vpc,
		ClusterName: jsii.String("websocket-cluster"),
		ContainerInsights: jsii.Bool(true),
	})

	// Create CloudWatch Log Group
	logGroup := awslogs.NewLogGroup(stack, jsii.String("WebSocketLogGroup"), &awslogs.LogGroupProps{
		LogGroupName: jsii.String("/aws/ecs/websocket-service"),
		Retention: awslogs.RetentionDays_ONE_WEEK,
	})

	// Create Fargate Service with Application Load Balancer
	fargateService := awsecspatterns.NewApplicationLoadBalancedFargateService(stack, jsii.String("WebSocketService"), &awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
		Cluster: cluster,
		Cpu: jsii.Number(512),
		MemoryLimitMiB: jsii.Number(1024),
		DesiredCount: jsii.Number(1),
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
		Protocol: awselasticloadbalancingv2.ApplicationProtocol_HTTP,
	})

	// Configure health check for websocket service
	fargateService.TargetGroup().ConfigureHealthCheck(&awselasticloadbalancingv2.HealthCheck{
		Path: jsii.String("/health"),
		HealthyThresholdCount: jsii.Number(2),
		UnhealthyThresholdCount: jsii.Number(5),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Interval: awscdk.Duration_Seconds(jsii.Number(30)),
		Port: jsii.String("8080"),
		Protocol: awselasticloadbalancingv2.Protocol_HTTP,
	})

	// Configure load balancer for WebSocket support
	loadBalancer := fargateService.LoadBalancer()
	targetGroup := fargateService.TargetGroup()

	// Set target group attributes for WebSocket support
	targetGroup.SetAttribute(jsii.String("load_balancing.algorithm.type"), jsii.String("least_outstanding_requests"))
	targetGroup.SetAttribute(jsii.String("stickiness.enabled"), jsii.String("true"))
	targetGroup.SetAttribute(jsii.String("stickiness.type"), jsii.String("lb_cookie"))

	// Add WebSocket specific listener rule
	listener := fargateService.Listener()
	listener.AddAction(jsii.String("WebSocketAction"), &awselasticloadbalancingv2.AddApplicationActionProps{
		Action: awselasticloadbalancingv2.ListenerAction_Forward(&[]*awselasticloadbalancingv2.IApplicationTargetGroup{
			targetGroup,
		}),
		Conditions: &[]*awselasticloadbalancingv2.ListenerCondition{
			awselasticloadbalancingv2.ListenerCondition_PathPatterns(&[]*string{jsii.String("/ws")}),
		},
		Priority: jsii.Number(100),
	})

	// Output the important URLs and information
	awscdk.NewCfnOutput(stack, jsii.String("ServiceURL"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(jsii.String(""), &[]*string{
			jsii.String("http://"),
			loadBalancer.LoadBalancerDnsName(),
		}),
		Description: jsii.String("WebSocket Service HTTP URL"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("WebSocketURL"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(jsii.String(""), &[]*string{
			jsii.String("ws://"),
			loadBalancer.LoadBalancerDnsName(),
			jsii.String("/ws"),
		}),
		Description: jsii.String("WebSocket Connection URL (NO AUTHENTICATION REQUIRED)"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("HealthCheckURL"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(jsii.String(""), &[]*string{
			jsii.String("http://"),
			loadBalancer.LoadBalancerDnsName(),
			jsii.String("/health"),
		}),
		Description: jsii.String("Health Check Endpoint"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("MetricsURL"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(jsii.String(""), &[]*string{
			jsii.String("http://"),
			loadBalancer.LoadBalancerDnsName(),
			jsii.String("/metrics"),
		}),
		Description: jsii.String("Prometheus Metrics Endpoint"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ECRRepository"), &awscdk.CfnOutputProps{
		Value: repository.RepositoryUri(),
		Description: jsii.String("ECR Repository URI for Docker images"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	// Read configuration from cdk.json
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

	// Create the main WebSocket service stack
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