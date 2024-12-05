package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGetParameter func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)

func (m mockGetParameter) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	return m(ctx, params, optFns...)
}

func TestRetrieveConfiguration(t *testing.T) {
	// JSON strings that would be returned from the SSM parameter store
	testTaskConfig1 := `{"cluster": "test-cluster","command": ["echo", "hello"],"subnetIds": ["subnet-123456"],"securityGroupIds": ["sg-123456"],"taskDefinitionArn": "arn:aws:ecs:us-west-2:123456789012:task-definition/test-task-1"}`
	testTaskConfig2 := `{"cluster": "test-cluster","command": ["echo", "byebye"],"subnetIds": ["subnet-654321"],"securityGroupIds": ["sg-654321"],"taskDefinitionArn": "arn:aws:ecs:us-west-2:123456789012:task-definition/test-task-2"}`

	mockedClient1 := mockGetParameter(func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
		return &ssm.GetParameterOutput{
			Parameter: &types.Parameter{
				Value: aws.String(testTaskConfig1),
			},
		}, nil
	})
	mockedClient2 := mockGetParameter(func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
		return &ssm.GetParameterOutput{
			Parameter: &types.Parameter{
				Value: aws.String(testTaskConfig2),
			},
		}, nil
	})

	// Running 2 different test cases to validate that the JSON payload unmarshals into the TaskRunnerConfiguration struct
	tests := []struct {
		name     string
		input    string
		client   ssmAPI
		expected *TaskRunnerConfiguration
	}{
		{
			name:   "given the parameter name test-parameter-name-1, it should return the configuration",
			input:  "test-parameter-name-1",
			client: mockedClient1,
			expected: &TaskRunnerConfiguration{
				Cluster:           "test-cluster",
				Command:           []string{"echo", "hello"},
				SecurityGroupIds:  []string{"sg-123456"},
				SubnetIds:         []string{"subnet-123456"},
				TaskDefinitionArn: "arn:aws:ecs:us-west-2:123456789012:task-definition/test-task-1",
			},
		},
		{
			name:   "given the parameter name test-parameter-name-2, it should return the configuration",
			input:  "test-parameter-name-2",
			client: mockedClient2,
			expected: &TaskRunnerConfiguration{
				Cluster:           "test-cluster",
				Command:           []string{"echo", "byebye"},
				SecurityGroupIds:  []string{"sg-654321"},
				SubnetIds:         []string{"subnet-654321"},
				TaskDefinitionArn: "arn:aws:ecs:us-west-2:123456789012:task-definition/test-task-2",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := RetrieveConfiguration(context.TODO(), tc.client, tc.input)

			t.Logf("result: %v", result)
			t.Logf("expected: %v", tc.expected)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
