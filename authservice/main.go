package authservice

import (
	"log"
	"net/http"
)

func main() {
	server, err := initialiseServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	http.HandleFunc("/signup", server.handleSignUp)
	http.HandleFunc("/login", server.handleLogin)
	http.HandleFunc("/resend-verification", server.handleResend)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
