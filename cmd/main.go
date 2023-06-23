package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/balintalfoldy/go-serverless/pkg/handlers"
)

var ddbclient *dynamodb.Client

var tableName string

func LambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		return handlers.GetUser(req, tableName, ddbclient)
	case "POST":
		return handlers.CreateUser(req, tableName, ddbclient)
	case "PUT":
		return handlers.UpdateUser(req, tableName, ddbclient)
	case "DELETE":
		return handlers.DeleteUser(req, tableName, ddbclient)
	default:
		return handlers.UnhandledMethod()
	}
}

func main() {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Fatal(err)
	}
	ddbclient = dynamodb.NewFromConfig(cfg)

	tableName = os.Getenv("DB_NAME")

	lambda.Start(LambdaHandler)
}
