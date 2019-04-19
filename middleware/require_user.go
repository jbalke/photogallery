package middleware

import (
	"net/http"

	"lenslocked.com/context"
	"lenslocked.com/models"
)

type User struct {
	models.UserService
}

type RequireUser struct {
	User
}

func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFN(next.ServeHTTP)
}

func (mw *User) ApplyFN(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}
		user, err := mw.ByRemember(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}
		ctx := context.WithUser(r.Context(), user)
		r = r.WithContext(ctx)
		next(w, r)
	})
}

func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFN(next.ServeHTTP)
}

func (mw *RequireUser) ApplyFN(next http.HandlerFunc) http.HandlerFunc {
	return mw.User.ApplyFN(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		next(w, r)
	})
}
