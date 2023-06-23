package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/balintalfoldy/go-serverless/pkg/validators"
)

var (
	ErrorFailedToFetchRecord     = "Failed to fetch record"
	ErrorFailedToUnmarshalRecord = "Failed to unmarshal record"
	ErrorInvalidUserData         = "Invalid user data"
	ErrorInvalidEmail            = "Invalid email"
	ErrorCouldNotMarshalItem     = "Could not marshal item"
	ErrorCouldNotDeleteItem      = "Could not delete item"
	ErrorCouldNotPutItem         = "Could not put item"
	ErrorUserAlreadyExists       = "User already exists"
	ErrorUserDoesNotExist        = "User does not exist"
)

type User struct {
	Email     string `dynamodbav:"email"`
	FirstName string `dynamodbav:"firstName"`
	LastName  string `dynamodbav:"lastName"`
}

func FetchUser(email string, tableName string, ddbclient *dynamodb.Client) (*User, error) {

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	}

	result, err := ddbclient.GetItem(context.TODO(), input)
	if err != nil {
		fmt.Printf("Error: %s - %s", ErrorFailedToFetchRecord, err.Error())
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	item := new(User)
	err = attributevalue.UnmarshalMap(result.Item, item)
	if err != nil {
		fmt.Printf("Error: %s - %s", ErrorFailedToUnmarshalRecord, err.Error())
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}
	return item, nil

}

func FetchUsers(tableName string, ddbclient *dynamodb.Client) (*[]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	result, err := ddbclient.Scan(context.TODO(), input)
	if err != nil {
		fmt.Printf("Error: %s - %s", ErrorFailedToFetchRecord, err.Error())
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	item := new([]User)
	err = attributevalue.UnmarshalListOfMaps(result.Items, item)
	if err != nil {
		fmt.Printf("Error: %s - %s", ErrorFailedToUnmarshalRecord, err.Error())
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}
	return item, nil
}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, ddbclient *dynamodb.Client) (*User, error) {

	var u User

	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}
	if !validators.IsEmailValid(u.Email) {
		return nil, errors.New(ErrorInvalidEmail)
	}
	currentUser, _ := FetchUser(u.Email, tableName, ddbclient)
	if currentUser != nil && len(currentUser.Email) != 0 {
		return nil, errors.New(ErrorUserAlreadyExists)
	}

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		fmt.Printf("Error: %s - %s", ErrorCouldNotMarshalItem, err.Error())
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	}

	_, err = ddbclient.PutItem(context.TODO(), input)
	if err != nil {
		fmt.Printf("Error: %s - %s", ErrorCouldNotPutItem, err.Error())
		return nil, errors.New(ErrorCouldNotPutItem)
	}

	return &u, nil

}

func UpdateUser(req events.APIGatewayProxyRequest, tableName string, ddbclient *dynamodb.Client) (*User, error) {

	var u User

	email := req.QueryStringParameters["email"]

	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}
	currentUser, _ := FetchUser(email, tableName, ddbclient)
	if currentUser == nil || len(currentUser.Email) == 0 {
		return nil, errors.New(ErrorUserDoesNotExist)
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":fn": &types.AttributeValueMemberS{Value: u.FirstName},
			":ln": &types.AttributeValueMemberS{Value: u.LastName},
		},
		ExpressionAttributeNames: map[string]string{
			"#fn": "firstName",
			"#ln": "lastName",
		},
		UpdateExpression: aws.String(fmt.Sprintf("set #fn = :fn, #ln = :ln")),
	}
	_, err := ddbclient.UpdateItem(context.TODO(), input)
	if err != nil {
		fmt.Printf("Error: %s - %s", ErrorCouldNotPutItem, err.Error())
		return nil, errors.New(ErrorCouldNotPutItem)
	}
	return &u, nil

}

func DeleteUser(req events.APIGatewayProxyRequest, tableName string, ddbclient *dynamodb.Client) error {

	email := req.QueryStringParameters["email"]
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	}
	_, err := ddbclient.DeleteItem(context.TODO(), input)
	if err != nil {
		fmt.Printf("Error: %s - %s", ErrorCouldNotDeleteItem, err.Error())
		return errors.New(ErrorCouldNotDeleteItem)
	}

	return nil
}
