package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	s3PutEvent = "ObjectCreated:Put"
)

var (
	dynamoDBClient *dynamodb.Client
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}
	dynamoDBClient = dynamodb.NewFromConfig(cfg)
}

func main() {
	lambda.Start(handler)
}

func handler(sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		if err := processMessage(message); err != nil {
			log.Printf("Error processing message: %v", err)
		}
	}
	return nil
}

func processMessage(e events.SQSMessage) error {
	fileName, err := extractFileName(e)
	if err != nil {
		return err
	}

	if fileName == "" {
		return nil // No action needed if no fileName
	}

	if err := createItemInDynamoDB(fileName); err != nil {
		return err
	}

	log.Printf("File processed: %s", fileName)
	return nil
}

func extractFileName(e events.SQSMessage) (string, error) {
	var snsEvent events.SNSEntity
	if err := json.Unmarshal([]byte(e.Body), &snsEvent); err != nil {
		return "", err
	}

	if !strings.Contains(snsEvent.Message, s3PutEvent) {
		return "", nil // Ignore non-S3 Put events
	}

	var s3Event events.S3Event
	if err := json.Unmarshal([]byte(snsEvent.Message), &s3Event); err != nil {
		return "", err
	}

	key, err := url.QueryUnescape(s3Event.Records[0].S3.Object.Key)
	if err != nil {
		return "", err
	}

	return filepath.Base(key), nil
}

func createItemInDynamoDB(fileName string) error {
	now := time.Now()
	id := strconv.FormatInt(now.Unix(), 10)
	item := map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: id},
		"name": &types.AttributeValueMemberS{Value: fileName},
	}

	tableName := os.Getenv("TABLE_NAME")
	_, err := dynamoDBClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      item,
	})
	if err != nil {
		return err
	}

	log.Println("Item successfully written to DynamoDB")
	return nil
}
