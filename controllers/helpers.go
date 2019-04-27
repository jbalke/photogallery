package controllers

import (
	"net/http"
	"net/url"

	schema "github.com/gorilla/Schema"
)

// Declared globally due to metadata caching benefit.
var dec = schema.NewDecoder()

func ParseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseValues(r.PostForm, dst)
}

func ParseURLParams(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseValues(r.Form, dst)
}

func parseValues(values url.Values, dst interface{}) error {
	// ignore unknown keys so we can use gorilla csrf tokens
	dec.IgnoreUnknownKeys(true)

	if err := dec.Decode(dst, values); err != nil {
		return err
	}

	return nil
}
