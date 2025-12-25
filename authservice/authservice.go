package authservice

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func enablePostCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) handlePasswordReset(w http.ResponseWriter, r *http.Request) {
	// user clicks forgot password on frontend
	// sends email to backend
	// frontend turns into enter token page
	// gets the data
	if r.Method != http.MethodPost {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get email

	// check if user exists

	// create a token and send token to users email
	// expiration time 1min

	// send token to email

	// send response to frontend that token was sent

	// user enters token on frontend frontend verifies it
	// frontend changes to new password page

	// gets data of new password
	// generate new salt and hash
	// update passwords table

	// send response of password reset

}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// gets the data
	if r.Method != http.MethodPost {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	// log in using email OR username
	if r.Method != http.MethodPost {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	type result struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var res result

	// put result into res

	var ch bool
	if res.Email == "" && res.Username != "" {
		// check username given
		ch = true
	} else if res.Email != "" && res.Username == "" {
		// !check email given
		ch = false
	}

	// client := s.dynamoClient

	// checks if email exists, if not, say username or password incorrect

	if !ch {
		Items, err := s.QueryWithEmail(res.Email)
		if err != nil {
			http.Error(w, "error checking email", http.StatusInternalServerError)
			return
		}

		if len(Items) == 0 {
			http.Error(w, "username or password incorrect", http.StatusUnauthorized)
			return
		}

	}

	if ch {
		Item, err := s.QueryWithUsername(res.Username)
		if err != nil {
			http.Error(w, "error checking username", http.StatusInternalServerError)
			return
		}
		if Item == nil {
			http.Error(w, "username or password incorrect", http.StatusUnauthorized)
			return
		}

	}

	// check if user verified

	verified, err := s.CheckUserVerified(res.Email, res.Username, ch)
	if err != nil {
		http.Error(w, "error checking verification", http.StatusInternalServerError)
		return
	}

	if !verified {
		http.Error(w, "user not verified", http.StatusUnauthorized)
		// send verification email again?
		return
	}

	// get salt and hash from PASS table
	salt, hash, err := s.QueryPasswordTable(res.Username, res.Email, ch)

	if err != nil {
		http.Error(w, "error querying passwords table", http.StatusInternalServerError)
		return
	}

	// verify password
	verified, err = VerifyPass(res.Password, salt, hash)
	if err != nil {
		http.Error(w, "error verifying password", http.StatusInternalServerError)
		return
	}

	if !verified {
		http.Error(w, "username or password incorrect", http.StatusUnauthorized)
		return
	}

	// genreate a session token and send back

}

func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) {

	enablePostCors(w)
	// gets the data
	if r.Method == http.MethodOptions {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	type result struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var res result

	// get the response here

	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if res.Username == "" || res.Email == "" || res.Password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// client := s.dynamoClient
	// cfg := s.cfg

	// queries the DB to see if email or username exists

	Item1, err := s.QueryWithEmail(res.Email)
	if err != nil {
		http.Error(w, "error checking email", http.StatusInternalServerError)
		return
	}
	if len(Item1) != 0 {
		http.Error(w, "user already exists for username or email", http.StatusConflict)
		return
	}

	Item2, err := s.QueryWithUsername(res.Username)

	if err != nil {
		http.Error(w, "error checking username", http.StatusConflict)
		return
	}

	if Item2 != nil {
		http.Error(w, "user already exists for username or email", http.StatusConflict)
		return
	}

	// generates hashed password

	salt, hash, err := HashPassword(res.Password)
	if err != nil {
		http.Error(w, "error hashing pass", http.StatusInternalServerError)
		return
	}
	// create SNS topic for verification

	ARN, err := s.CreateSNSTopic(res.Username)
	if err != nil {
		http.Error(w, "error creating sns topic", http.StatusInternalServerError)
		return
	}

	// stores user info in USER table and password in PASS table

	err = s.PutIntoUsersTable(res.Username, res.Email, ARN)
	if err != nil {
		http.Error(w, "error storing user info", http.StatusInternalServerError)
		return
	}

	err = s.PutIntoPassTable(res.Username, salt, hash)
	if err != nil {
		http.Error(w, "error storing password info", http.StatusInternalServerError)
		return
	}

	// send verification email

	link, err := s.GenerateVerificationLink(res.Username)
	if err != nil {
		http.Error(w, "error generating verification link", http.StatusInternalServerError)
		return
	}

	err = s.SendSNS(ARN, link)
	if err != nil {
		http.Error(w, "error sending verification link", http.StatusInternalServerError)
		return
	}

	// signed up
	// send response of successful sign up and verification link
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account created successfully",
		"link":    link,
	})

}

func (s *Server) handleResend(w http.ResponseWriter, r *http.Request) {
	// gets reponse as either verified or request to resend link
	enablePostCors(w)

	if r.Method == http.MethodOptions {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get response
	type result struct {
		Username string `json:"username"`
		// Resend   bool   `json:"resend"`
		Email string `json:"email"`
	}

	var res result

	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// query token table to see if token for user exists and not expired

	link, err := s.QueryTokenTable(res.Username) // true means token valid
	if err != nil {
		http.Error(w, "error querying token table", http.StatusInternalServerError)
		return
	}

	// get sns arn of user
	Item, err := s.QueryWithUsername(res.Username)
	if err != nil {
		http.Error(w, "error querying user table", http.StatusInternalServerError)
		return
	}

	arn, ok := Item["ARN"].(*types.AttributeValueMemberS)
	if !ok {
		http.Error(w, "error getting sns arn", http.StatusInternalServerError)
		return
	}

	// send link
	err = s.SendSNS(arn.Value, link)
	if err != nil {
		http.Error(w, "error sending sns", http.StatusInternalServerError)
		return
	}

	// send response link resent
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Verification link resent successfully",
		"link":    link,
	})

	// // if resend is false
	// // check if user verified
	// verified, err := s.CheckUserVerified
	// // set user verified
	// err = s.SetUserVerified(res.Username)
	// if err != nil {
	// 	http.Error(w, "error setting user verified", http.StatusInternalServerError)
	// 	return
	// }

	// // send response user verified
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(map[string]string{
	// 	"message": "User verified successfully",
	// })

}
