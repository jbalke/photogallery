package main

import (
	"fmt"
	"net/http"

	"lenslocked.com/controllers"
	"lenslocked.com/models"

	"github.com/gorilla/mux"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "sParhwk72"
	dbname   = "lenslocked_dev"
	httpPort = ":3000"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	services, err := models.NewServices(psqlInfo, true)
	must(err)
	defer services.Close()

	// must(services.DestructiveReset())
	services.AutoMigrate()

	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User)
	galleriesController := controllers.NewGalleries(services.Gallery)

	r := mux.NewRouter()
	r.Handle("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")
	r.Handle("/login", usersController.LoginView).Methods("GET")
	r.HandleFunc("/login", usersController.Login).Methods("POST")
	r.HandleFunc("/signup", usersController.New).Methods("GET")
	r.HandleFunc("/signup", usersController.Create).Methods("POST")
	r.HandleFunc("/cookietest", usersController.CookieTest).Methods("GET")

	// Galleries routes
	r.Handle("/galleries/new", galleriesController.New).Methods("GET")
	r.HandleFunc("/galleries", galleriesController.Create).Methods("POST")
	fmt.Println("Server listening on port", httpPort)
	http.ListenAndServe(httpPort, r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
