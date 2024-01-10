package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/config"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/graphql"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Handler struct {
	service *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{service: services}
}

func (h *Handler) InitRoutes() *chi.Mux {
	graphqlHandler := graphql.Handler(h.service)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	if config.Get().App.Mode == "development" {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(h.AuthJWT)
	r.Handle("/graphql", graphqlHandler)
	r.Post("/upload", h.UploadFile)

	r.Group(func(r chi.Router) {

	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 — Page not found"))
	})

	// Create a route that will serve contents from the ../assets/ folder.
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "../../storage"))
	fileServer(r, "/", filesDir)

	return r
}

func fileServer(mx *chi.Mux, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		mx.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	mx.Get(path, func(w http.ResponseWriter, r *http.Request) {
		c := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(c.RoutePattern(), "/*")
		fh := http.StripPrefix(pathPrefix, http.FileServer(root))

		if fs, err := os.Stat(fmt.Sprintf("%s", root) + r.URL.Path); os.IsNotExist(err) || fs.IsDir() {
			mx.NotFoundHandler().ServeHTTP(w, r)
		} else {
			fh.ServeHTTP(w, r)
		}
	})
}

//TODO: Add error rendering handlers.
