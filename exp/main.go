package main

import (
	"html/template"
	"os"
)

type User struct {
	Name     string
	Dog      string
	Age      int
	Float    float64
	Slice    []string
	Script   string
	FavFoods map[string]string
}

func main() {
	t, err := template.ParseFiles("hello.gohtml")
	if err != nil {
		panic(err)
	}

	data := User{Name: "John Balke", Dog: "Jessie", Age: 5, Script: "<script>Alert('Ah Ha!')</script>", FavFoods: map[string]string{"Bananas": "Fresh Bananas", "Coffee": "Allpress Espresso"}, Float: 3.14, Slice: []string{"a", "b", "c"}}

	err = t.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}
