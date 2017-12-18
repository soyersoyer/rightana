package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/soyersoyer/k20a/cmd"
)

var (
	app                   = kingpin.New("k20a", "Knight20 Studio's open source web analytics.")
	serve                 = app.Command("serve", "Serve")
	seed                  = app.Command("seed", "Seed")
	seedCollectionID      = seed.Arg("id", "Collection's ID").Required().String()
	seedCount             = seed.Arg("count", "Session Count").Required().Int()
	register              = app.Command("register", "Register a new user.")
	registerEmail         = register.Arg("email", "Email for user.").Required().String()
	registerPassword      = register.Arg("password", "Password for user.").Required().String()
	createCollection      = app.Command("create-collection", "Create a collection")
	createCollectionID    = createCollection.Arg("id", "Collection's ID").Required().String()
	createCollectionName  = createCollection.Arg("name", "Collection's name").Required().String()
	createCollectionEmail = createCollection.Arg("email", "Owner's email").Required().String()
)

func main() {
	app.Version("0.2")
	app.UsageTemplate(kingpin.CompactUsageTemplate)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "serve":
		cmd.Serve()
	case "seed":
		cmd.Seed(*seedCollectionID, *seedCount)
	case "register":
		cmd.RegisterUser(*registerEmail, *registerPassword)
	case "create-collection":
		cmd.CreateCollection(*createCollectionID, *createCollectionName, *createCollectionEmail)
	}

}
