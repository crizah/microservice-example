package authservice

import (
	"net/http"
)

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

	// queries the DB to see if email exists

	check, err := CheckEmailExists(res.Email, client)
	if err != nil {
		http.Error(w, "error checking email", http.StatusInternalServerError)
		return
	}

	if check {
		http.Error(w, "email already exists", http.StatusConflict)
		return
	}

	// generates hashed password

	salt, hash, err := HashPassword(res.Password)

	// stores user info in USER table and password in PASS table

	err = PutIntoUsersTable(res.Username, res.Email, client)
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

	// signed up

}
