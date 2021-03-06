package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"lenslocked.com/controllers"
	"lenslocked.com/email"
	"lenslocked.com/middleware"
	"lenslocked.com/models"
	"lenslocked.com/rand"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// TODO: Add a 404 page
// TODO: Change gallery ID to unique id to deter discover without an invite link
func main() {
	prodPtr := flag.Bool("prod", false, "Include this flag in production. This ensures use of .config for application settings and will panic instead of using dev defaults.")
	flag.Parse()

	cfg := LoadConfig(*prodPtr)
	dbCfg := cfg.Database
	services, err := models.NewServices(
		models.WithGorm(dbCfg.Dialect(), dbCfg.ConnectionInfo()),
		models.WithUser(cfg.HMACKey, cfg.Pepper),
		models.WithGallery(),
		models.WithImage(),
		models.WithLogMode(!cfg.IsProd()),
	)
	must(err)
	defer services.Close()
	services.AutoMigrate()
	// must(services.DestructiveReset())

	mailCfg := cfg.Email
	emailer := email.NewClient(
		email.WithSender("Lenslocked.com Support", "support@lenslocked.com"),
		email.WithMailgun(mailCfg.Domain, mailCfg.APIKey, mailCfg.PublicKey),
	)

	r := mux.NewRouter()
	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User, emailer)
	galleriesController := controllers.NewGalleries(services.Gallery, services.Image, r)

	bytes, err := rand.Bytes(32)
	must(err)

	csrfMw := csrf.Protect(bytes, csrf.Secure(cfg.IsProd()))
	userMw := middleware.User{
		UserService: services.User,
	}
	requireUserMw := middleware.RequireUser{
		User: userMw,
	}

	// Static page routes
	r.Handle("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")

	// User routes
	r.Handle("/login", usersController.LoginView).Methods("GET")
	r.HandleFunc("/login", usersController.Login).Methods("POST")
	r.HandleFunc("/logout", requireUserMw.ApplyFN(usersController.Logout)).Methods("POST")
	r.HandleFunc("/signup", usersController.New).Methods("GET")
	r.HandleFunc("/signup", usersController.Create).Methods("POST")
	r.Handle("/forgot", usersController.ForgotPwView).Methods("GET")
	r.HandleFunc("/forgot", usersController.InitiateReset).Methods("POST")
	r.HandleFunc("/reset", usersController.ResetPw).Methods("GET")
	r.HandleFunc("/reset", usersController.CompleteReset).Methods("POST")
	//r.HandleFunc("/cookietest", usersController.CookieTest).Methods("GET")

	// Static assets
	assetHandler := http.FileServer(http.Dir("./assets"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", assetHandler))

	// Image routes
	imageHandler := http.FileServer(http.Dir("./images"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Galleries middleware & routes
	r.HandleFunc("/galleries", requireUserMw.ApplyFN(galleriesController.Index)).Methods("GET")
	r.Handle("/galleries/new", requireUserMw.Apply(galleriesController.New)).Methods("GET")
	r.HandleFunc("/galleries", requireUserMw.ApplyFN(galleriesController.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFN(galleriesController.Edit)).Methods("GET").Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMw.ApplyFN(galleriesController.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", requireUserMw.ApplyFN(galleriesController.ImageUpload)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", requireUserMw.ApplyFN(galleriesController.ImageDelete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", requireUserMw.ApplyFN(galleriesController.Delete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesController.Show).Methods("GET").Name(controllers.ShowGallery)
	log.Printf("Server listening on port: %d...\n", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), csrfMw(userMw.Apply(r))))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
