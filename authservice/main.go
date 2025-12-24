package authservice

import "net/http"

func main() {
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/signup", handleSignUp)
	http.ListenAndServe(":443", nil)

}
