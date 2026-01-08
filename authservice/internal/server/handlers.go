package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func enablePostCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) HandlePasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	// user clicks forgot password on frontend
	// sends email to backend
	// frontend turns into enter token page page
	// gets the data

	enablePostCors(w)

	if r.Method != http.MethodPost {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	// gets email

	type response struct {
		Email string `json:"email"`
	}
	var res response

	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// check if user exists

	Items, err := s.QueryWithEmail(res.Email)
	if err != nil {
		http.Error(w, "error querying users table", http.StatusInternalServerError)
		return

	}
	if len(Items) == 0 {
		// send response that email has been sent if user exists
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "If an account exists with this email, a reset code has been sent",
		})
		return
	}

	// get username
	username := ""
	if u, ok := Items[0]["Username"].(*types.AttributeValueMemberS); ok {
		username = u.Value

	}

	// create a token and send token to users email
	// expiration time 10min

	token := GeneratePasswordResetToken()
	err = s.PutIntoPasswordTokenTable(token, username)
	if err != nil {
		http.Error(w, "error putting into table", http.StatusInternalServerError)
		return
	}

	// send token to email

	err = s.SendTokenEmail(res.Email, username, token)
	if err != nil {
		http.Error(w, "error sending email", http.StatusInternalServerError)
		return
	}

	// send response to frontend that token was sent
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "If an account exists with this email, a reset code has been sent",
	})
	// user enters token on frontend

}
func (s *Server) HandlePasswordReset(w http.ResponseWriter, r *http.Request) {

	enablePostCors(w)
	// sends user entererd token for verification as well as new password
	if r.Method != http.MethodPost {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	// gets data

	type request struct {
		Email       string `json:"email"`
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}

	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Email == "" || req.Token == "" || req.NewPassword == "" {
		http.Error(w, "Email, token, and new password are required", http.StatusBadRequest)
		return
	}

	// get username
	Items, err := s.QueryWithEmail(req.Email)
	if err != nil || len(Items) == 0 {
		http.Error(w, "Invalid or expired reset token", http.StatusUnauthorized)
		return
	}

	username := ""
	if u, ok := Items[0]["username"].(*types.AttributeValueMemberS); ok {
		username = u.Value
	}

	// first verify token

	valid, err := s.VerifyPasswordResetToken(username, req.Token)
	if err != nil || !valid {
		http.Error(w, "Invalid or expired reset token", http.StatusUnauthorized)
		return
	}

	// generate new salt and hash
	// update passwords table
	salt, hash, err := HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	// update password in database
	err = s.UpdatePasswordInTable(username, salt, hash)
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	// delete the reset token
	err = s.DeleteToken(username, "Password-reset-Token")
	if err != nil {
		log.Printf("Error deleting reset token: %v", err)
	}

	// send response of password reset
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successfully",
	})

}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	enablePostCors(w)
	w.Header().Set("Access-Control-Allow-Credentials", "true") // IMPORTANT for cookies

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

	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if (res.Email == "" && res.Username == "") || res.Password == "" {
		// !u and !e
		http.Error(w, "Username/email and password required", http.StatusBadRequest)
		return
	}

	var ch bool
	if res.Username != "" {
		// check username given
		ch = true
		// u and e
		// u and !e
	} else {

		// !check email given
		ch = false
	}

	// if email given, get username
	username := ""
	if !ch {
		Items, err := s.QueryWithEmail(res.Email)
		if err != nil {
			http.Error(w, "coundlt query with email", http.StatusInternalServerError)
			return

		}
		// checks if user exists, if not, say username or password incorrect
		if len(Items) == 0 {
			http.Error(w, "username or password incorrect", http.StatusUnauthorized)
			return
		}

		if u, ok := Items[0]["username"].(*types.AttributeValueMemberS); ok {
			username = u.Value
		}

	} else {
		username = res.Username
	}

	// check if user verified

	verified, err := s.CheckUserVerified(username)
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
	salt, hash, err := s.QueryPasswordTable(username)

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
	sID, err := s.CreateSession(username)

	if err != nil {
		http.Error(w, "error creating sesion", http.StatusInternalServerError)
		return

	}

	// set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sID,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HttpOnly: true,             // prevents js access
		Secure:   false,            // Set to true in production (HTTPS only)
		SameSite: http.SameSiteLaxMode,
	})

	// send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Login successful",
		"username": username,
	})

}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "No active session", http.StatusBadRequest)
		return
	}

	// Delete session from DB
	err = s.DeleteSession(cookie.Value)
	if err != nil {
		log.Printf("Error deleting session: %v", err)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// func (s *Server) ValidateSession(next http.HandlerFunc) http.HandlerFunc {
//     return func(w http.ResponseWriter, r *http.Request) {
//         w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
//         w.Header().Set("Access-Control-Allow-Credentials", "true")

//         // Get session cookie
//         cookie, err := r.Cookie("session_token")
//         if err != nil {
//             http.Error(w, "Unauthorized - no session", http.StatusUnauthorized)
//             return
//         }

//         // Validate session
//         username, err := s.GetSessionUser(cookie.Value)
//         if err != nil {
//             http.Error(w, "Unauthorized - invalid session", http.StatusUnauthorized)
//             return
//         }

//         // Add username to context
//         ctx := context.WithValue(r.Context(), "username", username)
//         next.ServeHTTP(w, r.WithContext(ctx))
//     }
// }

func (s *Server) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
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

	// queries the DB to see if email or username exists

	Item1, err := s.QueryWithEmail(res.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	// generate link

	link, err := s.GenerateVerificationLink(res.Username)
	if err != nil {
		http.Error(w, "error generating verification link", http.StatusInternalServerError)
		return
	}

	// stores user info in USER table and password in PASS table and send email

	var wg sync.WaitGroup

	ch := make(chan error, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()
		ch <- s.PutIntoPassTable(res.Username, salt, hash, ch)

	}()

	go func() {
		defer wg.Done()
		ch <- s.PutIntoUsersTable(res.Username, res.Email, ch)

	}()

	wg.Wait()
	close(ch)
	for msg := range ch {
		if msg != nil {
			http.Error(w, "error storing user info", http.StatusInternalServerError)
			return

		}
	}

	err = s.SendVerificationEmail(res.Email, res.Username, link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// signed up
	// send response of successful sign up and verification link
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account created successfully",
	})

}

func (s *Server) HandleEmailVerification(w http.ResponseWriter, r *http.Request) {
	// user clicks on link
	// frontend sends token to server
	// server verifies token
	// marks user as verified

	enablePostCors(w)

	if r.Method == http.MethodOptions {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get response (token)

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	// get username from token table

	username, err := s.GetUsernameWithToken(token)
	if err != nil {
		http.Error(w, "error querying token table", http.StatusInternalServerError)
		return
	}

	// check if token is expired
	Item, err := s.QueryEmailTokenTableWithUsername(username)
	if err != nil {
		http.Error(w, "error querying tokens table", http.StatusInternalServerError)
		return
	}

	valid, err := s.CheckTokenExpired(username, Item)
	if !valid {
		// token is no longer valid
		// resend link

		http.Error(w, "Token expired", http.StatusUnauthorized)
		return

	}

	// mark user as verified
	err = s.SetUserVerified(username)
	if err != nil {
		http.Error(w, "error verifying user", http.StatusInternalServerError)
		return
	}

	// delete token

	err = s.DeleteToken(username, "Email-token")
	if err != nil {
		http.Error(w, "error deleting token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verified successfully",
	})

}

func (s *Server) HandleResend(w http.ResponseWriter, r *http.Request) {
	// gets reponse as either verified or request to resend link
	enablePostCors(w)

	if r.Method == http.MethodOptions {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get response
	type result struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	var res result

	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// query token table to see if token for user exists and not expired
	link, err := s.GetTokenLinkWithUsername(res.Username)
	if err != nil {
		http.Error(w, "error getting link", http.StatusInternalServerError)
		return
	}

	// type Info struct {
	// 	Error error
	// 	Link  string
	// 	Item  map[string]types.AttributeValue
	// }

	// ch := make(chan *Info, 2)

	// go func() {

	// 	// using username, returns link if token valid, else generate new token and link and return new link
	// 	link, err := s.GetTokenLinkWithUsername(res.Username)
	// 	ch <- &Info{
	// 		Error: err,
	// 		Link:  link,
	// 	}

	// }()

	// go func() {

	// 	item, err := s.QueryWithUsername(res.Username)
	// 	ch <- &Info{
	// 		Error: err,
	// 		Item:  item,
	// 	}

	// }()

	// var Item map[string]types.AttributeValue
	// var link string

	// for msg := range ch {
	// 	if msg.Error != nil {
	// 		http.Error(w, "error querying token table", http.StatusInternalServerError)
	// 		return

	// 	}

	// 	if msg.Item != nil {
	// 		Item = msg.Item
	// 	}

	// 	if msg.Link != "" {
	// 		link = msg.Link
	// 	}
	// }
	// var (
	// 	wg   sync.WaitGroup
	// 	Item map[string]types.AttributeValue
	// 	link string
	// 	err1 error
	// 	err2 error
	// )

	// wg.Add(2)

	// go func() {
	// 	defer wg.Done()
	// 	link, err1 = s.GetTokenLinkWithUsername(res.Username)
	// }()

	// go func() {
	// 	defer wg.Done()
	// 	Item, err2 = s.QueryWithUsername(res.Username)
	// }()

	// wg.Wait()

	// if err1 != nil || err2 != nil {
	// 	http.Error(w, "error querying data", http.StatusInternalServerError)
	// 	return
	// }

	// link, err := s.GetTokenLinkWithUsername(res.Username) // using username, returns link if token valid, else generate new token and link and return new link
	// if err != nil {
	// 	http.Error(w, "error querying token table", http.StatusInternalServerError)
	// 	return
	// }

	// // get sns arn of user
	// Item, err := s.QueryWithUsername(res.Username)
	// if err != nil {
	// 	http.Error(w, "error querying user table", http.StatusInternalServerError)
	// 	return
	// }

	// get sns arn of user

	// arn, ok := Item["ARN"].(*types.AttributeValueMemberS)
	// if !ok {
	// 	http.Error(w, "error getting sns arn", http.StatusInternalServerError)
	// 	return
	// }

	// send link
	err = s.SendVerificationEmail(res.Email, res.Username, link)
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

}
