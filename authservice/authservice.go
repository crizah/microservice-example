package authservice

import (
	"net/http"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
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

	client := Dynamoclient()

	// checks if email exists, if not, say username or password incorrect

	if !ch {
		Items, err := QueryWithEmail(res.Email, client)
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
		Item, err := QueryWithUsername(res.Username, client)
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

	verified, err := UserVerified(res.Email, res.Username, client, ch)
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
	salt, hash, err := QueryPasswordTable(res.Username, res.Email, client, ch)

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

func handleSignUp(w http.ResponseWriter, r *http.Request) {
	// gets the data
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

	// get the response here

	client := Dynamoclient()

	// queries the DB to see if email or username exists

	Item1, err := QueryWithEmail(res.Email, client)
	if err != nil {
		http.Error(w, "error checking email", http.StatusInternalServerError)
		return
	}
	if len(Item1) != 0 {
		http.Error(w, "user already exists for username or email", http.StatusConflict)
		return
	}

	Item2, err := QueryWithUsername(res.Username, client)

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

	ARN, err := CreateSNSTopic(res.Username, cfg)
	if err != nil {
		http.Error(w, "error creating sns topic", http.StatusInternalServerError)
		return
	}

	// stores user info in USER table and password in PASS table

	err = PutIntoUsersTable(res.Username, res.Email, ARN, client)
	if err != nil {
		http.Error(w, "error storing user info", http.StatusInternalServerError)
		return
	}

	err = PutIntoPassTable(res.Username, salt, hash, client)
	if err != nil {
		http.Error(w, "error storing password info", http.StatusInternalServerError)
		return
	}

	// send verification email

	link, err := GenerateVerificationLink()
	if err != nil {
		http.Error(w, "error generating verification link", http.StatusInternalServerError)
		return
	}

	err = SendSNS(cfg, ARN, link)
	if err != nil {
		http.Error(w, "error sending verification link", http.StatusInternalServerError)
		return
	}

	// signed up
	// send response of successful sign up and verification link

}
