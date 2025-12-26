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

func (s *Server) GenerateVerificationLink(username string) (string, error) {
	// generate token
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	// generate link using token
	link := "https://yourfrontend.com/verify-email?token=" + token

	err = s.PutintoTokenTable(username, token, link)
	if err != nil {
		return "", err
	}

	// a user might try to request multiple email verifications.
	// Since we have a generous expiration time, we can reuse the previously generated token
	// in the database and send the same link each time the user requests a new email.
	// If the expiration time is up, we can generate a new token and send the new link to the user.

	return link, nil

}

func (s *Server) CreateSNSTopic(topicName string) (string, error) {
	snsClient := sns.NewFromConfig(s.cfg)

	output, err := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})

	if err != nil {
		return "", err
	}

	return *output.TopicArn, nil
}

func (s *Server) SendSNS(snsTopicARN string, link string) error {
	snsClient := sns.NewFromConfig(s.cfg)

	_, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: aws.String(snsTopicARN),
		Message:  aws.String(fmt.Sprintf("Your verification link is %s", link)),
		Subject:  aws.String("New Message"),
	})

	return err

}

// func (s *Server) SendVerificationEmail(email, username, link string) error {
//     htmlBody := fmt.Sprintf(`
//         <html>
//             <body>
//                 <h2>Verify Your Email</h2>
//                 <p>Hi %s,</p>
//                 <p>Thanks for signing up! Please verify your email by clicking the link below:</p>
//                 <a href="%s" style="background-color: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">
//                     Verify Email
//                 </a>
//                 <p>Or copy and paste this link in your browser:</p>
//                 <p>%s</p>
//                 <p>This link will expire in 24 hours.</p>
//             </body>
//         </html>
//     `, username, link, link)

//     textBody := fmt.Sprintf("Hi %s,\n\nPlease verify your email by clicking this link:\n%s\n\nThis link will expire in 24 hours.", username, link)

//     input := &ses.SendEmailInput{
//         Source: aws.String("noreply@yourdomain.com"),
//         Destination: &ses.Destination{
//             ToAddresses: []string{email},
//         },
//         Message: &ses.Message{
//             Subject: &ses.Content{
//                 Data: aws.String("Verify Your Email"),
//             },
//             Body: &ses.Body{
//                 Html: &ses.Content{
//                     Data: aws.String(htmlBody),
//                 },
//                 Text: &ses.Content{
//                     Data: aws.String(textBody),
//                 },
//             },
//         },
//     }

//     _, err := s.sesClient.SendEmail(context.TODO(), input)
//     return err
// }
