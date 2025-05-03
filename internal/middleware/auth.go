package middleware

// import (
// 	"net/http"

// 	"github.com/nu-kotov/gophermart/internal/auth"
// )

// func RequestSession(h http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		_, err := r.Cookie("token")
// 		if err != nil {
// 			value, err := auth.BuildJWTString()
// 			if err != nil {
// 				w.WriteHeader(http.StatusBadRequest)
// 				return
// 			}
// 			cookie := &http.Cookie{
// 				Name:     "token",
// 				Value:    value,
// 				HttpOnly: true,
// 			}
// 			r.AddCookie(cookie)
// 			http.SetCookie(w, cookie)
// 			h.ServeHTTP(w, r)
// 			return
// 		}
// 		h.ServeHTTP(w, r)
// 	})
// }
