package server

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
