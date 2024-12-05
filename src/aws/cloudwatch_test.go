package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCloudwatchLogsClient struct {
	mockGetLogEvents func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)
}

func (m mockCloudwatchLogsClient) GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return m.mockGetLogEvents(ctx, params, optFns...)
}

func TestRetrieveLogs(t *testing.T) {
	input := LogDetails{
		logGroupName:  "test-group",
		logStreamName: "test-stream/test-container/07cc583696bd44e0be450bff7314ddaf",
	}

	events := []types.OutputLogEvent{
		{
			IngestionTime: aws.Int64(0),
			Message:       aws.String("beans have been detected in the system"),
			Timestamp:     aws.Int64(0),
		},
		{
			IngestionTime: aws.Int64(1),
			Message:       aws.String("beans have been removed from the system"),
			Timestamp:     aws.Int64(1),
		},
	}

	positiveTests := []struct {
		name     string
		input    LogDetails
		client   mockCloudwatchLogsClient
		expected []types.OutputLogEvent
	}{
		{
			// This is a regression test to ensure the function signature remains the same
			name:  "given a valid LogDetails input, return the events of the log stream",
			input: input,
			client: mockCloudwatchLogsClient{
				mockGetLogEvents: func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
					return &cloudwatchlogs.GetLogEventsOutput{Events: events}, nil
				},
			},
			expected: events,
		},
	}

	for _, tc := range positiveTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := RetrieveLogs(context.TODO(), tc.client, tc.input)

			t.Logf("result: %v", result)
			t.Logf("expected: %v", tc.expected)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}

	negativeTests := []struct {
		name     string
		input    LogDetails
		client   mockCloudwatchLogsClient
		expected []types.OutputLogEvent
	}{
		{
			name:  "when the underlying cloudwatch client experiences an error, return it in the function ",
			input: input,
			client: mockCloudwatchLogsClient{
				mockGetLogEvents: func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
					return &cloudwatchlogs.GetLogEventsOutput{}, errors.New("generic cloudwatch error")
				},
			},
			expected: []types.OutputLogEvent{},
		},
	}

	for _, tc := range negativeTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := RetrieveLogs(context.TODO(), tc.client, tc.input)

			t.Logf("result: %v", result)
			t.Logf("expected: %v", tc.expected)
			require.Error(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
