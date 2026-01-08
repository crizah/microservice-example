package server

import (
	"context"

	"errors"
	"fmt"
	"math/rand"

	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (s *Server) PutIntoPassTable(username string, salt string, hash string, ch chan<- error) error {
	// username partition key
	// hash is reserved keyword

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

func (s *Server) UpdatePasswordInTable(username string, salt string, hash string) error {
	_, err := s.dynamoClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Pass"),
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{Value: username},
		},
		UpdateExpression: aws.String("SET salt = :salt, #hash = :hash"),
		ExpressionAttributeNames: map[string]string{
			"#hash": "hash", // hash is a reserved keyword
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":salt": &types.AttributeValueMemberS{Value: salt},
			":hash": &types.AttributeValueMemberS{Value: hash},
		},
	})

	return err

}

func (s *Server) PutintoEmailTokenTable(username string, token string, link string) error {

	now := time.Now().Unix()

	// username is pk
	// token in gsi called TokenIndex

	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	_, err := s.dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Email-token"),
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

func (s *Server) PutIntoUsersTable(username string, email string, ch chan<- error) error {
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
			// "ARN":       &types.AttributeValueMemberS{Value: arn},
		},
	})

	return err

}

func GeneratePasswordResetToken() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *Server) PutIntoPasswordTokenTable(token string, username string) error {
	now := time.Now().Unix()

	// username is pk

	expiresAt := time.Now().Add(10 * time.Minute).Unix() // expires in 10mins

	_, err := s.dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Password-reset-token"),
		Item: map[string]types.AttributeValue{
			"username":  &types.AttributeValueMemberS{Value: username},
			"token":     &types.AttributeValueMemberS{Value: token},
			"createdAt": &types.AttributeValueMemberS{Value: strconv.FormatInt(now, 10)},
			"expiresAt": &types.AttributeValueMemberS{Value: strconv.FormatInt(expiresAt, 10)},
		},
	})

	return err

}

func (s *Server) VerifyPasswordResetToken(username string, token string) (bool, error) {
	// query db with username
	result, err := s.dynamoClient.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("Password-reset-Token"),
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
		return false, err
	}

	// check if token expired
	valid, err := s.CheckTokenExpired(username, result.Item)
	if err != nil {
		return false, err
	}

	if !valid {
		return false, errors.New("token expired")
	}

	// verify if token is same
	got_token := ""

	if tok, ok := result.Item["token"].(*types.AttributeValueMemberS); ok {
		got_token = tok.Value

	}

	if token != got_token {
		return false, nil
	}

	return true, nil

}

func (s *Server) QueryWithEmail(email string) ([]map[string]types.AttributeValue, error) {
	// query users table with email
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

func (s *Server) QueryPasswordTable(username string) (string, string, error) {

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

func (s *Server) CheckUserVerified(username string) (bool, error) {

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

func (s *Server) GetUsernameWithToken(token string) (string, error) {
	// queries email verification token table
	// gets username from token

	result, err := s.dynamoClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("Email-token"),
		IndexName:              aws.String("TokenIndex"),
		KeyConditionExpression: aws.String("token = :token"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":token": &types.AttributeValueMemberS{Value: token},
		},
	})

	if err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return "", nil
	}

	user, ok := result.Items[0]["username"].(*types.AttributeValueMemberS)
	if !ok {
		return "", nil
	}

	return user.Value, nil
}

func (s *Server) DeleteToken(username string, table string) error {
	_, err := s.dynamoClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{Value: username},
		},
	})
	return err
}

func (s *Server) QueryEmailTokenTableWithUsername(username string) (map[string]types.AttributeValue, error) {
	result, err := s.dynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("Email-token"),
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{Value: username},
		},
	})
	if err != nil {
		return nil, err

	}

	return result.Item, nil

}

func (s *Server) CheckTokenExpired(username string, Item map[string]types.AttributeValue) (bool, error) {
	// true if exists and not expired

	cA, ok := Item["createdAt"].(*types.AttributeValueMemberS)

	if !ok {
		return true, nil
	}

	createdAt, err := strconv.ParseInt(cA.Value, 10, 64)
	if err != nil {
		return true, err
	}

	eA, ok := Item["expiresAt"].(*types.AttributeValueMemberS)

	if !ok {
		return true, nil
	}

	expiredAt, err := strconv.ParseInt(eA.Value, 10, 64)
	if err != nil {
		return true, err
	}
	now := time.Now().Unix()

	if now < createdAt || now > expiredAt {
		// token is expired
		return true, nil
	}

	return false, nil

}
func (s *Server) GetTokenLinkWithUsername(username string) (string, error) {
	// using username, returns link if token valid, else generate new token and link and return new link
	Item, err := s.QueryEmailTokenTableWithUsername(username)
	if err != nil {
		return "", err
	}

	expired, err := s.CheckTokenExpired(username, Item)
	if err != nil {
		return "", err
	}
	if expired {
		// token is expired
		// generate new token
		link, err := s.GenerateVerificationLink(username) // this uses puinto token table,
		if err != nil {
			return "", err
		}
		return link, nil
	}

	// token is valid
	link, ok := Item["link"].(*types.AttributeValueMemberS)
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

func (s *Server) CreateSession(username string) (string, error) {
	sessionToken, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix() // 7 days

	// sessionToken is pk
	// username gpi

	_, err = s.dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Sessions"),
		Item: map[string]types.AttributeValue{
			"sessionToken": &types.AttributeValueMemberS{Value: sessionToken},
			"username":     &types.AttributeValueMemberS{Value: username},
			"createdAt":    &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
			"expiresAt":    &types.AttributeValueMemberN{Value: strconv.FormatInt(expiresAt, 10)},
		},
	})

	return sessionToken, err
}

func (s *Server) GetSessionUser(sessionToken string) (string, error) {
	result, err := s.dynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("Sessions"),
		Key: map[string]types.AttributeValue{
			"sessionToken": &types.AttributeValueMemberS{Value: sessionToken},
		},
	})

	if err != nil {
		return "", err
	}

	if result.Item == nil {
		return "", fmt.Errorf("session not found")
	}

	// Check expiration
	expiresAt := int64(0)
	if n, ok := result.Item["expiresAt"].(*types.AttributeValueMemberN); ok {
		expiresAt, _ = strconv.ParseInt(n.Value, 10, 64)
	}

	if time.Now().Unix() > expiresAt {
		// Delete expired session
		s.DeleteSession(sessionToken)
		return "", fmt.Errorf("session expired")
	}

	// Get username
	username := ""
	if u, ok := result.Item["username"].(*types.AttributeValueMemberS); ok {
		username = u.Value
	}

	return username, nil
}

func (s *Server) DeleteSession(sessionToken string) error {
	_, err := s.dynamoClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String("Sessions"),
		Key: map[string]types.AttributeValue{
			"sessionToken": &types.AttributeValueMemberS{Value: sessionToken},
		},
	})
	return err
}
