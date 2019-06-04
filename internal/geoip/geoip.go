package geoip

import (
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

var cityDB *geoip2.Reader
var asnDB *geoip2.Reader

// OpenDB opens the geoip2 databases
func OpenDB(cityDBFile string, asnDBFile string) {
	openDB(&cityDB, cityDBFile)
	openDB(&asnDB, asnDBFile)
}

func openDB(db **geoip2.Reader, dbFile string) {
	var err error
	*db, err = geoip2.Open(dbFile)
	if err != nil {
		log.Println(err, dbFile)
	}
}

// Location contains the Geoip2 Location data
type Location struct {
	CountryCode string
	City        string
}

// LocationByIP returns the corresponding Location data
func LocationByIP(ipAddr string) *Location {
	if cityDB != nil {
		ip := net.ParseIP(ipAddr)
		record, err := cityDB.City(ip)
		if err == nil {
			return &Location{
				record.Country.IsoCode,
				record.City.Names["en"],
			}
		}
	}
	return &Location{}
}

// AS contains the Geoip2 AS data
type AS struct {
	Number uint
	Name   string
}

// ASNByIP returns the corresponding AS data
func ASNByIP(ipAddr string) *AS {
	if asnDB != nil {
		ip := net.ParseIP(ipAddr)
		if record, err := asnDB.ASN(ip); err == nil {
			return &AS{
				record.AutonomousSystemNumber,
				record.AutonomousSystemOrganization,
			}
		}
	}
	return &AS{}
}
