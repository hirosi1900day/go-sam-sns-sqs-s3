package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandler(t *testing.T) {
	inputJson := readJsonFromFile(t, "../testdata/s3event.json")
	snsEvent := events.SNSEvent{
		Records: []events.SNSEventRecord{
			{
				SNS: events.SNSEntity{
					Message: string(inputJson)},
			},
		},
	}
	snsEventByte, err := json.Marshal(snsEvent)
	if err != nil {
		t.Errorf("error: %s", err)
	}

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				Body: string(snsEventByte),
			},
		},
	}

	if err := handler(sqsEvent); err != nil {
		t.Errorf("error: %s", err)
	}
}

func readJsonFromFile(t *testing.T, inputFile string) []byte {
	inputJson, err := os.ReadFile(inputFile)
	if err != nil {
		t.Errorf("could not open test file. details: %v", err)
	}

	return inputJson
}
