package authservice

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func Dynamoclient() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := dynamodb.NewFromConfig(cfg)
	return client

}

func PutIntoPassTable(username string, salt string, hash string, client *dynamodb.Client) error {
	// username partition key

	_, err := client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Pass"),
		Item: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{Value: username},
			"salt":     &types.AttributeValueMemberS{Value: salt},
			"hash":     &types.AttributeValueMemberS{Value: hash},
		},
	})

	return err

}

func PutIntoUsersTable(username string, email string, client *dynamodb.Client) error {
	// username primary key
	// make email partition key

	timestamp := time.Now().Format(time.RFC3339)
	_, err := client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Users"),
		Item: map[string]types.AttributeValue{
			"username":  &types.AttributeValueMemberS{Value: username},
			"email":     &types.AttributeValueMemberS{Value: email},
			"createdAt": &types.AttributeValueMemberS{Value: timestamp},
			"verified":  &types.AttributeValueMemberS{Value: aws.FalseTernary.String()},
		},
	})

	return err

}

func CheckEmailExists(email string, client *dynamodb.Client) (bool, error) {
	result, err := client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{
				Value: email,
			},
		},
	})

	if err != nil {
		return false, err
	}

	if result.Item != nil {
		return true, nil
	}

	return false, nil
}

func UpdateUserVerification(username string, client *dynamodb.Client) error {
	_, err := client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Users"),
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{
				Value: username,
			},
		},
		UpdateExpression:          aws.String("SET verified = :v"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":v": &types.AttributeValueMemberS{Value: aws.TrueTernary.String()}},
	})

	return err
}
