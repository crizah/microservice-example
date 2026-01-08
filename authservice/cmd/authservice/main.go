package main

import (
	"authservice/internal/server"
	"log"
	"net/http"
)

// whats left (GREEP MENTIONED????????)
// primary
// make middleware for validation
// change to ses

// secondary
// modify so they dont NEED to verify before login
// merge all token tables??

func main() {
	s, err := server.InitialiseServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	http.HandleFunc("/signup", s.HandleSignUp)
	http.HandleFunc("/verify-email", s.HandleEmailVerification)
	http.HandleFunc("/resend-verification", s.HandleResend)
	http.HandleFunc("/reset-password-requst", s.HandlePasswordResetRequest)
	http.HandleFunc("/reset-password", s.HandlePasswordReset)
	http.HandleFunc("/profile", s.HandleLogin)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
