package main

import (
	"fmt"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/soyersoyer/rightana/api"
	"github.com/soyersoyer/rightana/config"
	"github.com/soyersoyer/rightana/db/db"
	"github.com/soyersoyer/rightana/geoip"
	"github.com/soyersoyer/rightana/mail"
	"github.com/soyersoyer/rightana/service"
)

func inits() {
	config.ReadConfig()
	geoip.OpenDB(config.ActualConfig.GeoIPCityFile, config.ActualConfig.GeoIPASNFile)
	db.InitDatabase(config.ActualConfig.DataDir)
	mail.Configure(mail.SMTPConfig{
		Hostname: config.ActualConfig.SMTPHostname,
		User:     config.ActualConfig.SMTPUser,
		Password: config.ActualConfig.SMTPPassword,
		Port:     config.ActualConfig.SMTPPort,
		Sender:   config.ActualConfig.SMTPSender,
		AppName:  config.ActualConfig.AppName,
		AppURL:   config.ActualConfig.AppURL,
	})
}

// Serve starts a http server
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

// Seed seed a collection with count session
func Seed(trackingID string, count int) {
	inits()
	now := time.Now()
	start := now.AddDate(0, -int(now.Month())+1, -int(now.Day())+1)
	end := start.AddDate(1, 0, 0)
	if err := service.SeedCollection(start, end, trackingID, count); err != nil {
		log.Fatalln(err)
	}
}

// RegisterUser registers a new user
func RegisterUser(email string, name string) {
	inits()
	config.ActualConfig.EnableRegistration = true
	fmt.Print("Password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("")
	user, err := service.CreateUser(&service.CreateUserT{
		Email:    email,
		Name:     name,
		Password: string(password)})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("user created:", user.Name)
}

// ChangePassword changes a user's password
func ChangePassword(name string) {
	inits()
	user, err := service.GetUserByName(name)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print("Password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("")
	err = service.ChangePasswordForce(user, string(password))
	if err != nil {
		log.Fatalln(err)
	}
}

// CreateCollection creates a collection with name and the owner's username
func CreateCollection(collectionID string, name string, user string) {
	inits()
	collection, err := service.CreateCollectionByID(collectionID, name, user)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("collection created", collection)
}
