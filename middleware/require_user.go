package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"lenslocked.com/context"
	"lenslocked.com/models"
)

type User struct {
	models.UserService
}

func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFN(next.ServeHTTP)
}

func (mw *User) ApplyFN(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the user is requesting a static asset or image we can skip looking up the user.
		// Can also resolve this with sub-routers with different middlewares applied.
		path := r.URL.Path
		if strings.HasPrefix(path, "/assets/") || strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}

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

// RequireUser assumes that User middleware has already been run
// otherwise it will not work.
type RequireUser struct {
	User
}

func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFN(next.ServeHTTP)
}

func (mw *RequireUser) ApplyFN(next http.HandlerFunc) http.HandlerFunc {
	return mw.User.ApplyFN(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			url := fmt.Sprintf("/login?ref=%s", r.URL.Path)
			http.Redirect(w, r, url, http.StatusFound)
			return
		}
		next(w, r)
	})
}
