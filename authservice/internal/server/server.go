package server

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type Server struct {
	dynamoClient *dynamodb.Client
	sesClient    *ses.Client
	// snsClient    *sns.Client
	// cfg          aws.Config
}

func InitialiseServer() (*Server, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)
	sesclient := ses.NewFromConfig(cfg)

	server := &Server{
		dynamoClient: client,
		sesClient:    sesclient,
	}

	return server, nil

}
