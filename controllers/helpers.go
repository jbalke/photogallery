package controllers

import (
	"net/http"

	schema "github.com/gorilla/Schema"
)

// Declared globally due to metadata caching benefit.
var dec = schema.NewDecoder()

func ParseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := dec.Decode(dst, r.PostForm); err != nil {
		return err
	}

	return nil
}
