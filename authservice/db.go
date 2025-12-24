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

func PutIntoUsersTable(username string, email string, arn string, client *dynamodb.Client) error {
	// username primary key
	// email is gsi called EmailIndex

	timestamp := time.Now().Format(time.RFC3339)
	_, err := client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Users"),
		Item: map[string]types.AttributeValue{
			"username":  &types.AttributeValueMemberS{Value: username},
			"email":     &types.AttributeValueMemberS{Value: email},
			"createdAt": &types.AttributeValueMemberS{Value: timestamp},
			"verified":  &types.AttributeValueMemberBOOL{Value: false},
			"ARN":       &types.AttributeValueMemberS{Value: arn},
		},
	})

	return err

}

func QueryWithEmail(email string, client *dynamodb.Client) ([]map[string]types.AttributeValue, error) {
	result, err := client.Query(context.Background(), &dynamodb.QueryInput{
		TableName:              aws.String("Users"),
		IndexName:              aws.String("EmailIndex"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{
				Value: email,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return result.Items, nil

}

func QueryWithUsername(username string, client *dynamodb.Client) (map[string]types.AttributeValue, error) {
	result, err := client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{
				Value: username,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return result.Item, nil

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

func QueryPasswordTable(username string, email string, client *dynamodb.Client, check bool) (string, string, error) {
	// check username given
	if check {
		// query using username
		Item, err := QueryWithUsername(username, client)
		if err != nil {
			return "", "", err
		}
		if Item == nil {
			return "", "", nil
		}
		salt, ok := Item["salt"].(*types.AttributeValueMemberS)
		if !ok {
			return "", "", nil
		}
		hash, ok := Item["hash"].(*types.AttributeValueMemberS)
		if !ok {
			return "", "", nil
		}
		return salt.Value, hash.Value, nil
	}

	// query using email
	Items, err := QueryWithEmail(email, client)
	if err != nil {
		return "", "", err
	}
	if len(Items) == 0 {
		return "", "", nil
	}
	salt, ok := Items[0]["salt"].(*types.AttributeValueMemberS)
	if !ok {
		return "", "", nil
	}
	hash, ok := Items[0]["hash"].(*types.AttributeValueMemberS)
	if !ok {
		return "", "", nil
	}
	return salt.Value, hash.Value, nil

}

func UserVerified(email string, username string, client *dynamodb.Client, check bool) (bool, error) {
	// check username given
	if check {
		// check by username
		result, err := client.GetItem(context.Background(), &dynamodb.GetItemInput{
			TableName: aws.String("Users"),
			Key: map[string]types.AttributeValue{
				"username": &types.AttributeValueMemberS{
					Value: username,
				},
			},
		})

		if err != nil {
			return false, err
		}

		if result.Item == nil {
			return false, nil
		}

		verifiedAttr, ok := result.Item["verified"].(*types.AttributeValueMemberBOOL)
		if !ok {
			return false, nil
		}

		return verifiedAttr.Value, nil
	}

	// check by email
	result, err := client.Query(context.Background(), &dynamodb.QueryInput{
		TableName:              aws.String("Users"),
		IndexName:              aws.String("EmailIndex"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{
				Value: email,
			},
		},
	})

	if err != nil {
		return false, err
	}

	if len(result.Items) == 0 {
		return false, nil
	}

	verifiedAttr, ok := result.Items[0]["verified"].(*types.AttributeValueMemberBOOL)
	if !ok {
		return false, nil
	}

	return verifiedAttr.Value, nil
}
