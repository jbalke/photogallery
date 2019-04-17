package main

import (
	"fmt"
	"log"
	"net/http"

	"lenslocked.com/controllers"
	"lenslocked.com/middleware"
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

	r := mux.NewRouter()
	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User)
	galleriesController := controllers.NewGalleries(services.Gallery, r)

	r.Handle("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")
	r.Handle("/login", usersController.LoginView).Methods("GET")
	r.HandleFunc("/login", usersController.Login).Methods("POST")
	r.HandleFunc("/signup", usersController.New).Methods("GET")
	r.HandleFunc("/signup", usersController.Create).Methods("POST")
	r.HandleFunc("/cookietest", usersController.CookieTest).Methods("GET")

	// Galleries middleware & routes
	requireUserMw := middleware.RequireUser{
		UserService: services.User,
	}

	r.Handle("/galleries/new", requireUserMw.Apply(galleriesController.New)).Methods("GET")
	r.HandleFunc("/galleries", requireUserMw.ApplyFN(galleriesController.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesController.Show).Methods("GET").Name(controllers.ShowGallery)
	fmt.Println("Server listening on port", httpPort)
	log.Fatal(http.ListenAndServe(httpPort, r))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
