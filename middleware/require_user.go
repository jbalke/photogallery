package middleware

import (
	"net/http"

	"lenslocked.com/context"
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
			// TODO: Remember original destination so we can redirect user to where they want to be
			// after after login
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		user, err := mw.ByRemember(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		next(w, r)
	})
}
