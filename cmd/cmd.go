package cmd

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"git.irl.hu/k20a/api"
	"git.irl.hu/k20a/config"
	"git.irl.hu/k20a/db/db"
	"git.irl.hu/k20a/geoip"
	"git.irl.hu/k20a/models"
)

var (
	configFile = "k20a"
)

func inits() {
	config.ReadConfig(configFile)
	geoip.OpenDB(config.ActualConfig.GeoIPCityFile, config.ActualConfig.GeoIPASNFile)
	db.InitDatabase(config.ActualConfig.DataDir)
}

func Serve() {
	inits()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.DefaultCompress)

	api.Wire(r)

	log.Println("HTTP server will now start listening on", config.ActualConfig.Listening)
	err := http.ListenAndServe(config.ActualConfig.Listening, r)
	log.Fatal(err)
}

func Seed(trackingID string, count int) {
	inits()
	now := time.Now()
	start := now.AddDate(0, -int(now.Month())+1, -int(now.Day())+1)
	end := start.AddDate(1, 0, 0)
	if err := models.SeedCollection(start, end, trackingID, count); err != nil {
		log.Fatalln(err)
	}
}

func RegisterUser(email string, password string) {
	inits()
	config.ActualConfig.EnableRegistration = true
	user, err := models.CreateUser(email, password)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("user created:", user.Email)
}

func CreateCollection(trackingID string, name string, email string) {
	inits()
	collection, err := models.CreateCollectionById(trackingID, name, email)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("collection created", collection)
}
