package db

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/mssola/user_agent"

	"github.com/soyersoyer/rightana/geoip"
)

var userAgents = []string{
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2049.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36",
}

var deviceTypes = []string{
	"mobile",
	"desktop",
	"tablet",
}

var browserLanguages = []string{
	"en-US",
	"en-US",
	"nl-NL",
	"fr-FR",
	"de-DE",
	"es-ES",
	"hu-HU",
}

var resolutions = []string{
	"2560x1440",
	"1920x1080",
	"1920x1080",
	"360x640",
}

var urls = []string{
	"dl",
	"ld?q=22222",
	"ndl;matrixnotation=true",
	"hdl",
}

var referrers = []string{
	"https://google.com",
	"https://google.com/?q=longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglsearch",
	"https://yahoo.com",
	"https://bing.com",
	"https://wikipedia.org",
}

var userHostnames = []string{
	"localhost",
	"catv-176-63-166-75.catv.broadband.hu.",
	"telekom.hu",
	"digi.hu",
}

// Seed seeds a collection with n session
func Seed(from time.Time, to time.Time, collectionID string, n int) error {
	rand.Seed(time.Now().UTC().UnixNano())
	start := time.Now()
	collection, err := GetCollection(collectionID)
	if err != nil {
		return err
	}
	fmt.Println("seeding from:", from, "to:", to, "n:", n)
	sdb, err := getShardDB(collectionID)
	if err != nil {
		return err
	}
	tx := sdb.Begin(true)
	for i := 0; i < n; i++ {
		sessionID := rand.Uint32()
		userAgent := randElem(userAgents)
		ua := user_agent.New(userAgent)
		browserName, browserVersion := ua.Browser()
		ip := fmt.Sprintf("95.85.%d.%d", randInt(1, 254), randInt(1, 254))
		location := geoip.LocationByIP(ip)
		asn := geoip.ASNByIP(ip)
		host := collection.Name
		duration := time.Duration(to.Sub(from).Seconds()/float64(n)*float64(i)) * time.Second
		tfrom := from.Add(duration)
		randomSessionDuration := rand.Intn(7200)
		if i%100 == 0 {
			fmt.Printf("\r%d", i)
		}

		session := &Session{
			Duration:         int32(randomSessionDuration),
			Hostname:         host,
			DeviceOS:         ua.OS(),
			UserIP:           ip,
			UserHostname:     randElem(userHostnames),
			BrowserName:      browserName,
			BrowserVersion:   browserVersion,
			BrowserLanguage:  randElem(browserLanguages),
			ScreenResolution: randElem(resolutions),
			WindowResolution: randElem(resolutions),
			DeviceType:       randElem(deviceTypes),
			CountryCode:      location.CountryCode,
			City:             location.City,
			ASNumber:         int32(asn.Number),
			ASName:           asn.Name,
			UserAgent:        userAgent,
			Referrer:         randElem(referrers),
		}
		sessionKey := GetKey(tfrom, sessionID)
		if err := ShardUpsertTx(tx, sessionKey, session); err != nil {
			return fmt.Errorf("session %v insert error err: %v session: %v t %v id %v", i, err, session, tfrom, sessionID)
		}
		for j := 0; j < i%10+1; j++ {
			pvfrom := tfrom.Add(time.Duration(j) * time.Minute)
			path, queryString := splitURL(randElem(urls))
			pageview := &Pageview{
				Path:        path,
				QueryString: queryString,
			}
			if err := ShardUpsertTx(tx, GetPVKey(sessionKey, pvfrom), pageview); err != nil {
				return fmt.Errorf("pageview %v %v insert error err: %v pv %v", i, j, err, pageview)
			}
		}

	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	fmt.Println("")
	elapsed := time.Since(start)
	log.Printf("seeding time: %s", elapsed)
	return nil
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randElem(list []string) string {
	return list[rand.Intn(len(list))]
}

/* TODO - remove from here */
func splitURL(url string) (path, queryString string) {
	idx := strings.IndexAny(url, "?;")
	if idx < 0 {
		return url, ""
	}
	return url[:idx], url[idx+1:]
}
