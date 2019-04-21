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

// TODO: Use Gorilla Session to make use of Flash messages and sessions
// OR: https://www.alexedwards.net/blog/simple-flash-messages-in-golang
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

	userMw := middleware.User{
		UserService: services.User,
	}

	r.Handle("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")
	r.Handle("/login", usersController.LoginView).Methods("GET")
	r.HandleFunc("/login", usersController.Login).Methods("POST")
	r.HandleFunc("/signup", usersController.New).Methods("GET")
	r.HandleFunc("/signup", usersController.Create).Methods("POST")
	r.HandleFunc("/cookietest", usersController.CookieTest).Methods("GET")

	// Galleries middleware & routes
	requireUserMw := middleware.RequireUser{
		User: userMw,
	}

	r.HandleFunc("/galleries", requireUserMw.ApplyFN(galleriesController.Index)).Methods("GET")
	r.Handle("/galleries/new", requireUserMw.Apply(galleriesController.New)).Methods("GET")
	r.HandleFunc("/galleries", requireUserMw.ApplyFN(galleriesController.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFN(galleriesController.Edit)).Methods("GET").Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMw.ApplyFN(galleriesController.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", requireUserMw.ApplyFN(galleriesController.ImageUpload)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", requireUserMw.ApplyFN(galleriesController.Delete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesController.Show).Methods("GET").Name(controllers.ShowGallery)
	fmt.Println("Server listening on port", httpPort)
	log.Fatal(http.ListenAndServe(httpPort, userMw.Apply(r)))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
