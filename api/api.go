package api

//go:generate statik -src=../frontend/dist/

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/rakyll/statik/fs"

	_ "github.com/soyersoyer/k20a/api/statik" //the embedded statik fs data
	"github.com/soyersoyer/k20a/config"
	"github.com/soyersoyer/k20a/errors"
)

type ctxKey int

const (
	keyUserID ctxKey = iota
	keyCollection
	keyUser
)

func webAppFileServer(dir string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(dir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(path.Clean("/"+r.URL.Path))))
		if err != nil && os.IsNotExist(err) {
			r.URL.Path = "/"
		}
		fs.ServeHTTP(w, r)
	})
}

func webAppFileServerBundled() http.HandlerFunc {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	fileServer := http.FileServer(statikFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := statikFS.Open(filepath.FromSlash(path.Clean("/" + r.URL.Path)))
		if err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}

// Wire function wires the http endpoints
func Wire(r *chi.Mux) {
	if config.ActualConfig.UseBundledWebApp {
		r.Get("/*", webAppFileServerBundled())
	} else {
		r.Get("/*", webAppFileServer("frontend/dist"))
	}
	r.Route("/api", func(r chi.Router) {
		cors := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"POST"},
			AllowedHeaders:   []string{},
			ExposedHeaders:   []string{},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		})
		r.Use(cors.Handler)
		r.Get("/config", getPublicConfig)
		r.Post("/sessions", createSession)
		r.Post("/sessions/update", updateSession)
		r.Post("/pageviews", createPageview)
		r.Post("/authtokens", createToken)
		r.Delete("/authtokens/{token}", deleteToken)
		r.Mount("/users", userRouter())
		r.Mount("/collections", loggedInRouter())
		r.Mount("/admin", adminRouter())
	})
}

func userRouter() http.Handler {
	r := chi.NewRouter()
	r.Post("/", createUser)
	r.Route("/{name}", func(r chi.Router) {
		r.Route("/settings", func(r chi.Router) {
			r.Use(loggedOnlyHandler)
			r.Use(userBaseHandler)
			r.Use(userAccessHandler)
			r.Patch("/password", updateUserPassword)
			r.Post("/delete", deleteUser)
		})
	})

	return r
}

func loggedInRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(loggedOnlyHandler)
	r.Get("/", getCollectionSummaries)
	r.Post("/", createCollection)
	r.Route("/{collectionID}", func(r chi.Router) {
		r.Use(collectionBaseHandler)
		r.Use(collectionReadAccessHandler)
		r.With(collectionWriteAccessHandler).Get("/", getCollection)
		r.With(collectionWriteAccessHandler).Put("/", updateCollection)
		r.With(collectionWriteAccessHandler).Delete("/", deleteCollection)
		r.With(collectionWriteAccessHandler).Get("/shards", getCollectionShards)
		r.With(collectionWriteAccessHandler).Delete("/shards/{shardID}", deleteCollectionShard)
		r.With(collectionWriteAccessHandler).Get("/teammates", getTeammates)
		r.With(collectionWriteAccessHandler).Post("/teammates", addTeammate)
		r.With(collectionWriteAccessHandler).Delete("/teammates/{email}", removeTeammate)
		r.Post("/data", getCollectionData)
		r.Post("/stat", getCollectionStatData)
		r.Post("/sessions", getSessions)
		r.Post("/pageviews", getPageviews)
	})
	return r
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(loggedOnlyHandler)
	r.Use(adminAccessHandler)
	r.Get("/users", getUsers)
	r.Get("/users/{email}", getUserInfo)
	r.Patch("/users/{email}", updateUser)
	r.Get("/collections", getCollections)
	return r
}

func respond(w http.ResponseWriter, d interface{}) error {
	w.Header().Set("content-type", "application/json")
	enc := json.NewEncoder(w)
	return enc.Encode(d)
}

type handlerFuncWithError func(w http.ResponseWriter, r *http.Request) error

func handleError(fn handlerFuncWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			switch e := err.(type) {
			default:
				log.Println(e)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			case *errors.Error:
				log.Println(e)
				http.Error(w, e.HTTPMessage(), e.Code)
			}
		}
	}
}
