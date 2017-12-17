package geoip

import (
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

var cityDB *geoip2.Reader
var asnDB *geoip2.Reader

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

type Location struct {
	CountryCode string
	City        string
}

func LocationByIp(ipAddr string) *Location {
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

type AS struct {
	Number uint
	Name   string
}

func ASNByIp(ipAddr string) *AS {
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
