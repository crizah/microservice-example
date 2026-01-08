package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

const (
	SENDER     = "sd"
	EMAIL_BODY = `
	    <html>
            <body>
                <h2>Verify Your Email</h2>
                <p>Hi %s,</p>
                <p>Thanks for signing up! Please verify your email by clicking the link below:</p>
                <a href="%s" style="background-color: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">
                    Verify Email
                </a>
                <p>Or copy and paste this link in your browser:</p>
                <p>%s</p>
                <p>This link will expire in 24 hours.</p>
            </body>
        </html>`

	EMAIL_TEXTBODY = "Hi %s,\n\nPlease verify your email by clicking this link:\n%s\n\nThis link will expire in 24 hours."
	EMAIL_SUBJECT  = "Verification email"
	TOKEN_BODY     = `
	    <html>
            <body>
                <h2>Password reset token</h2>
                <p>Hi %s,</p>
                <p>This is your password reset token. </p>
                <p>%s</p>
                <p>This token will expire in 10 minutes.</p>
            </body>
        </html>`

	TOKEN_TEXTBODY = "Hi %s,\n\nThis is your token for password reset\n%s\n\nThis link will expire in 10 minutes."

	TOKEN_SUBJECT = "Password Reset"
)

func generateToken() (string, error) {

	b := make([]byte, 128)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *Server) GenerateVerificationLink(username string) (string, error) {
	// generate token
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	// generate link using token
	link := "https://localhost:3000/verify-email?token=" + token

	err = s.PutintoEmailTokenTable(username, token, link)
	if err != nil {
		return "", err
	}

	// a user might try to request multiple email verifications.
	// Since we have a generous expiration time, we can reuse the previously generated token
	// in the database and send the same link each time the user requests a new email.
	// If the expiration time is up, we can generate a new token and send the new link to the user.

	return link, nil

}

// func (s *Server) CreateSNSTopic(topicName string) (string, error) {

// 	output, err := s.snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
// 		Name: aws.String(topicName),
// 	})

// 	if err != nil {
// 		return "", err
// 	}

// 	return *output.TopicArn, nil
// }

// func (s *Server) SendSNS(snsTopicARN string, link string) error {

// 	_, err := s.snsClient.Publish(context.TODO(), &sns.PublishInput{
// 		TopicArn: aws.String(snsTopicARN),
// 		Message:  aws.String(fmt.Sprintf("Your verification link is %s", link)),
// 		Subject:  aws.String("New Message"),
// 	})

// 	return err

// }

func (s *Server) SendVerificationEmail(email string, username string, link string) error {
	htmlBoody := fmt.Sprintf(EMAIL_BODY, username, link, link)
	textBody := fmt.Sprintf(EMAIL_TEXTBODY, username, link)

	_, err := s.sesClient.SendEmail(context.Background(), &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{email},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(htmlBoody),
				},
				Text: &types.Content{
					Data: aws.String(textBody),
				},
			},
			Subject: &types.Content{
				Data: aws.String(EMAIL_SUBJECT),
			},
		},
		Source: aws.String(SENDER),
	})

	return err
}

func (s *Server) SendTokenEmail(email string, username string, token string) error {
	htmlBoody := fmt.Sprintf(TOKEN_BODY, username, token)
	textBody := fmt.Sprintf(TOKEN_TEXTBODY, username, token)

	_, err := s.sesClient.SendEmail(context.Background(), &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{email},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(htmlBoody),
				},
				Text: &types.Content{
					Data: aws.String(textBody),
				},
			},
			Subject: &types.Content{
				Data: aws.String(TOKEN_SUBJECT),
			},
		},
		Source: aws.String(SENDER),
	})

	return err
}
