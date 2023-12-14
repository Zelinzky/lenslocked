package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"

	"lenslocked/controllers"
	"lenslocked/models"
	"lenslocked/static"
	"lenslocked/templates"
	"lenslocked/views"
)

func galleriesHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Fprint(w, "the id of the gallery is ", id)

}

func main() {
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "tailwind.gohtml", "home.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "tailwind.gohtml", "contact.gohtml"))))
	r.Get("/galleries/{id}", galleriesHandler)
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "tailwind.gohtml", "faq.gohtml"))))

	usersService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}
	usersC := controllers.Users{
		UserService:    &usersService,
		SessionService: &sessionService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(templates.FS, "tailwind.gohtml", "signup.gohtml"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "tailwind.gohtml", "signin.gohtml"))
	usersC.Templates.CurrentUser = views.Must(views.ParseFS(templates.FS, "tailwind.gohtml", "currentUser.gohtml"))
	r.Get("/signup", usersC.New)
	r.Post("/signup", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Get("/users/me", ipLog(usersC.CurrentUser))
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	csrfKey := "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		// TODO: Fix this before deploying
		csrf.Secure(false),
	)
	log.Println("Starting server on :3000")
	http.ListenAndServe(":3000", csrfMw(r))
}

// excercise middleware:

func ipLog(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		log.Printf("the ip of the request is: %s\n", ip)
		log.Printf("the path requested is: %s\n", r.RequestURI)
		next(w, r)
	}
	return fn
}
