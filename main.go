package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.URL.Path == "/" {
		fmt.Fprint(w, "<h1>Welcome to my site!</h1>")
	} else if r.URL.Path == "/contact" {
		fmt.Fprint(w, `<h2>To get in touch, please send an email to <a href="mailto:support@lenslocked.com">support@lenslocked.com</a></h2>`)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<h1>We could not find the page you were looking for :(</h1><p>Please email us if you keep being sent to an invalid page.`)
	}
}

func main() {
	http.HandleFunc("/", handlerFunc)
	http.ListenAndServe(":3000", nil)
}
