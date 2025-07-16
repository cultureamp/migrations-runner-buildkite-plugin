package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type cloudwatchLogsClientAPI interface {
	GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)
}

type LogDetails struct {
	logGroupName  string
	logStreamName string
}

func RetrieveLogs(ctx context.Context, cloudwatchLogsClientAPI cloudwatchLogsClientAPI, loggingDetails LogDetails) ([]types.OutputLogEvent, error) {
	response, err := cloudwatchLogsClientAPI.GetLogEvents(ctx, &cloudwatchlogs.GetLogEventsInput{
		LogStreamName: &loggingDetails.logStreamName,
		LogGroupName:  &loggingDetails.logGroupName,
		StartFromHead: aws.Bool(true),
	})
	if err != nil {
		return []types.OutputLogEvent{}, err
	}

	return response.Events, nil
}
