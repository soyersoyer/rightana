package api

//go:generate go-bindata-assetfs -ignore .map -pkg api -prefix ../ ../frontend/dist/ ../frontend/dist/assets/

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"

	"github.com/soyersoyer/k20a/errors"
)

func WebAppFileServer(dir string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(dir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(path.Clean("/"+r.URL.Path))))
		if err != nil && os.IsNotExist(err) {
			r.URL.Path = "/"
		}
		fs.ServeHTTP(w, r)
	})
}

func WebAppFileServerBundled(dir string) http.HandlerFunc {
	fs := http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: dir})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := AssetInfo(filepath.Join(dir, filepath.FromSlash(path.Clean("/"+r.URL.Path))))
		if err != nil {
			r.URL.Path = "/"
		}
		fs.ServeHTTP(w, r)
	})
}

func Wire(r *chi.Mux) {
	r.Get("/*", WebAppFileServerBundled("frontend/dist"))
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
	})
}

func userRouter() http.Handler {
	r := chi.NewRouter()
	r.Post("/", createUser)
	r.Route("/{email}", func(r chi.Router) {
		r.Use(LoggedOnlyHandler)
		r.Use(userBaseHandler)
		r.Use(userAccessHandler)
		r.Patch("/password", updateUserPassword)
		r.Post("/delete", deleteUser)
	})
	return r
}

func loggedInRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(LoggedOnlyHandler)
	r.Get("/", getCollections)
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

func respond(w http.ResponseWriter, d interface{}) error {
	w.Header().Set("content-type", "application/json")
	enc := json.NewEncoder(w)
	return enc.Encode(d)
}

type HandlerFuncWithError func(w http.ResponseWriter, r *http.Request) error

func handleError(fn HandlerFuncWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			switch e := err.(type) {
			default:
				log.Println(e)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			case *errors.Error:
				log.Println(e)
				http.Error(w, e.HttpMessage(), e.Code)
			}
		}
	}
}
