package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app                  = kingpin.New("rightana", "Rightana - open source web analytics.")
	serve                = app.Command("serve", "Serve")
	seed                 = app.Command("seed", "Seed")
	seedCollectionID     = seed.Arg("id", "Collection's ID").Required().String()
	seedCount            = seed.Arg("count", "Session Count").Required().Int()
	netseed              = app.Command("netseed", "Network Seed")
	netseedServer        = netseed.Arg("server", "Server address (eg http://localhost:3000)").Required().String()
	netseedCollectionID  = netseed.Arg("id", "Collection's ID").Required().String()
	netseedCount         = netseed.Arg("count", "Session Count").Required().Int()
	register             = app.Command("register", "Register a new user.")
	registerEmail        = register.Arg("email", "Email for user.").Required().String()
	registerName         = register.Arg("name", "Username for user.").Required().String()
	passwd               = app.Command("passwd", "Change user password")
	passwdName           = passwd.Arg("name", "username for user.").Required().String()
	createCollection     = app.Command("create-collection", "Create a collection")
	createCollectionID   = createCollection.Arg("id", "Collection's ID").Required().String()
	createCollectionName = createCollection.Arg("name", "Collection's name").Required().String()
	createCollectionUser = createCollection.Arg("user", "Owner's username").Required().String()
)

func main() {
	app.Version("0.4.1")
	app.UsageTemplate(kingpin.CompactUsageTemplate)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "serve":
		Serve()
	case "seed":
		Seed(*seedCollectionID, *seedCount)
	case "netseed":
		NetSeed(*netseedServer, *netseedCollectionID, *netseedCount)
	case "register":
		RegisterUser(*registerEmail, *registerName)
	case "passwd":
		ChangePassword(*passwdName)
	case "create-collection":
		CreateCollection(*createCollectionID, *createCollectionName, *createCollectionUser)
	}

}
