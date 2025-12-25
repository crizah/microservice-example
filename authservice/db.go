package authservice

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Server struct {
	dynamoClient *dynamodb.Client
	cfg          aws.Config
}

func initialiseServer() (*Server, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)

	server := &Server{
		dynamoClient: client,
		cfg:          cfg,
	}

	return server, nil

}

func (s *Server) PutIntoPassTable(username string, salt string, hash string) error {
	// username partition key

	_, err := s.dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Pass"),
		Item: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{Value: username},
			"salt":     &types.AttributeValueMemberS{Value: salt},
			"hash":     &types.AttributeValueMemberS{Value: hash},
		},
	})

	return err

}

func (s *Server) PutintoTokenTable(username string, token string, link string) error {

	// username is pk

	now := time.Now().Unix()
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	_, err := s.dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Token"),
		Item: map[string]types.AttributeValue{
			"username":  &types.AttributeValueMemberS{Value: username},
			"token":     &types.AttributeValueMemberS{Value: token},
			"link":      &types.AttributeValueMemberS{Value: link},
			"createdAt": &types.AttributeValueMemberS{Value: strconv.FormatInt(now, 10)},
			"expiresAt": &types.AttributeValueMemberS{Value: strconv.FormatInt(expiresAt, 10)},
		},
	})

	return err

}

func (s *Server) PutIntoUsersTable(username string, email string, arn string) error {
	// username primary key
	// email is gsi called EmailIndex

	timestamp := time.Now().Format(time.RFC3339)
	_, err := s.dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
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

func (s *Server) QueryWithEmail(email string) ([]map[string]types.AttributeValue, error) {
	result, err := s.dynamoClient.Query(context.Background(), &dynamodb.QueryInput{
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

func (s *Server) QueryWithUsername(username string) (map[string]types.AttributeValue, error) {
	result, err := s.dynamoClient.GetItem(context.Background(), &dynamodb.GetItemInput{
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

func (s *Server) UpdateUserVerification(username string) error {
	_, err := s.dynamoClient.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Users"),
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{
				Value: username,
			},
		},
		UpdateExpression:          aws.String("SET verified = :v"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":v": &types.AttributeValueMemberBOOL{Value: true}},
	})

	return err
}

func (s *Server) QueryPasswordTable(username string, email string, check bool) (string, string, error) {
	// check username given
	if check {
		// query using username
		Item, err := s.QueryWithUsername(username)
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
	Items, err := s.QueryWithEmail(email)
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

func (s *Server) CheckUserVerified(email string, username string, check bool) (bool, error) {
	// check username given
	if check {
		// check by username
		result, err := s.dynamoClient.GetItem(context.Background(), &dynamodb.GetItemInput{
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
	result, err := s.dynamoClient.Query(context.Background(), &dynamodb.QueryInput{
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

func (s *Server) QueryTokenTable(username string) (string, error) {
	result, err := s.dynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("Token"),
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{Value: username},
		},
	})
	if err != nil {
		return "", err
	}

	cA, ok := result.Item["createdAt"].(*types.AttributeValueMemberS)

	if !ok {
		return "", nil
	}

	createdAt, err := strconv.ParseInt(cA.Value, 10, 64)
	if err != nil {
		return "", err
	}

	eA, ok := result.Item["expiresAt"].(*types.AttributeValueMemberS)

	if !ok {
		return "", nil
	}

	expiredAt, err := strconv.ParseInt(eA.Value, 10, 64)
	if err != nil {
		return "", err
	}
	now := time.Now().Unix()

	if now < createdAt || now > expiredAt {
		// token is expired
		// generate new token
		link, err := s.GenerateVerificationLink(username) // this uses puinto token table,
		if err != nil {
			return "", err
		}
		return link, nil
	}

	// token is valid
	link, ok := result.Item["link"].(*types.AttributeValueMemberS)
	if !ok {
		return "", nil
	}
	return link.Value, nil

}

func (s *Server) SetUserVerified(username string) error {
	_, err := s.dynamoClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Users"),
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{Value: username},
		},
		UpdateExpression: aws.String("SET verified = :v"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v": &types.AttributeValueMemberBOOL{Value: true},
		},
	})

	return err
}
