package aws

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// internal interface for ssm
type ssmAPI interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// ECS Task Configuration
type TaskRunnerConfiguration struct {
	Cluster           string   `json:"cluster"`
	Command           []string `json:"command"`
	SecurityGroupIds  []string `json:"securityGroupIds"`
	SubnetIds         []string `json:"subnetIds"`
	TaskDefinitionArn string   `json:"taskDefinitionArn"`
}

// RetrieveConfiguration retrieves the configuration from the SSM parameter store
func RetrieveConfiguration(ctx context.Context, ssmAPI ssmAPI, parameterName string) (*TaskRunnerConfiguration, error) {
	res, err := ssmAPI.GetParameter(ctx, &ssm.GetParameterInput{
		Name: &parameterName,
	})
	if err != nil {
		return nil, err
	}

	value := &TaskRunnerConfiguration{}
	err = json.Unmarshal([]byte(*res.Parameter.Value), value)
	if err != nil {
		return nil, err
	}

	return value, nil
}
