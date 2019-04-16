package middleware

import (
	"fmt"
	"net/http"

	"lenslocked.com/models"
)

type RequireUser struct {
	models.UserService
}

func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFN(next.ServeHTTP)
}

func (mw *RequireUser) ApplyFN(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		user, err := mw.ByRemember(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		fmt.Println("user: ", user)
		next(w, r)
	})
}
