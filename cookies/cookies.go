package cookies

import (
	"net/http"
	"time"
)

func SetRedirect(w http.ResponseWriter, originalURL string) {
	expiresAt := time.Now().Add(10 * time.Minute)
	redirect := http.Cookie{
		Name:     "redirect",
		Value:    originalURL,
		Expires:  expiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, &redirect)
}

func ClearRedirect(w http.ResponseWriter) {
	expiresAt := time.Now()
	redirect := http.Cookie{
		Name:     "redirect",
		Value:    "",
		Expires:  expiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, &redirect)
}

func GetRedirect(r *http.Request) string {
	redirect, err := r.Cookie("redirect")
	if err != nil {
		return ""
	}
	return redirect.Value
}
