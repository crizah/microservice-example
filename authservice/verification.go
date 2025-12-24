package authservice

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func generateToken() (string, error) {
	b := make([]byte, 128)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GenerateVerificationLink() (string, error) {
	// generate token
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	// generate link using token
	link := "https://example.com/auth/verify-email?token=" + token

	// a user might try to request multiple email verifications.
	// Since we have a generous expiration time, we can reuse the previously generated token
	// in the database and send the same link each time the user requests a new email.
	// If the expiration time is up, we can generate a new token and send the new link to the user.

	return link, nil

}

func CreateSNSTopic(topicName string, cfg aws.Config) (string, error) {
	snsClient := sns.NewFromConfig(cfg)

	output, err := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})

	if err != nil {
		return "", err
	}

	return *output.TopicArn, nil
}

func SendSNS(cfg aws.Config, snsTopicARN string, link string) error {
	snsClient := sns.NewFromConfig(cfg)

	_, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: aws.String(snsTopicARN),
		Message:  aws.String(fmt.Sprintf("Your verification link is %s", link)),
		Subject:  aws.String("New Message"),
	})

	return err

}
