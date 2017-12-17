package db

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/mssola/user_agent"

	"github.com/soyersoyer/k20a/geoip"
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
	"ld",
	"ndl",
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
		location := geoip.LocationByIp(ip)
		asn := geoip.ASNByIp(ip)
		host := collection.Name
		duration := time.Duration(to.Sub(from).Seconds()/float64(n)*float64(i)) * time.Second
		tfrom := from.Add(duration)
		randomSessionDuration := time.Duration(rand.Intn(7200)) * time.Second
		tend := tfrom.Add(randomSessionDuration)
		if i%100 == 0 {
			fmt.Printf("\r%d", i)
		}

		session := &Session{
			End:              tend.UnixNano(),
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
			PageviewCount:    int32(i%10 + 1),
		}
		if err := ShardUpsertTx(tx, GetKey(tfrom, sessionID), session); err != nil {
			return fmt.Errorf("session %d insert error err: %s session: %s t %s id %s", i, err, session, tfrom, sessionID)
		}
		for j := 0; j < i%10+1; j++ {
			referrer := ""
			if j == 0 {
				referrer = randElem(referrers)
			}
			tfrom = tfrom.Add(time.Duration(j) * time.Minute)
			pageview := &Pageview{
				Path:        randElem(urls),
				ReferrerURL: referrer,
			}
			if err := ShardUpsertTx(tx, GetKey(tfrom, sessionID), pageview); err != nil {
				return fmt.Errorf("pageview %d %d insert error err: %s pv %s t %s id %s", i, j, err, pageview, tfrom, sessionID)
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
