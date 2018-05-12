package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type empty struct{}

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

// NetSeed seeds a collection over the network
func NetSeed(serverURL, collectionID string, n int) {
	defer trace("seed")()
	log.Println("seeding:", serverURL, "id:", collectionID, "with n:", n)
	var wg sync.WaitGroup
	var tokens = make(chan empty, 300)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tokens <- empty{}
			sessionID := createSession(serverURL, collectionID)
			for j := 0; j < (i%10)+1; j++ {
				createPageview(serverURL, collectionID, sessionID)
			}
			<-tokens
		}(i)
	}

	wg.Wait()
}

type createSessionInput struct {
	CollectionID     string `json:"c"`
	Hostname         string `json:"h"`
	BrowserLanguage  string `json:"bl"`
	ScreenResolution string `json:"sr"`
	WindowResolution string `json:"wr"`
	DeviceType       string `json:"dt"`
	Referrer         string `json:"r"`
}

func createSession(serverURL, collectionID string) string {
	createSession := createSessionInput{
		collectionID,
		"localhost",
		randElem(browserLanguages),
		randElem(resolutions),
		randElem(resolutions),
		randElem(deviceTypes),
		randElem(referrers),
	}

	b, _ := json.Marshal(createSession)

	req, err := http.NewRequest("POST", serverURL+"/api/sessions", bytes.NewBuffer(b))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("x-real-ip", fmt.Sprintf("95.85.%d.%d", randInt(1, 254), randInt(1, 254)))
	req.Header.Set("user-agent", randElem(userAgents))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var sessionID string
	if err := json.NewDecoder(resp.Body).Decode(&sessionID); err != nil {
		log.Fatalln(err)
	}
	return sessionID
}

type createPageviewInputT struct {
	CollectionID string `json:"c"`
	SessionKey   string `json:"s"`
	Path         string `json:"p"`
}

func createPageview(serverURL, collectionID, sessionID string) {
	input := &createPageviewInputT{
		collectionID,
		sessionID,
		randElem(urls),
	}
	b, _ := json.Marshal(input)

	req, err := http.NewRequest("POST", serverURL+"/api/pageviews", bytes.NewBuffer(b))
	if err != nil {
		log.Fatalln(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randElem(list []string) string {
	return list[rand.Intn(len(list))]
}

func trace(msg string) func() {
	start := time.Now()
	log.Printf("enter %s", msg)
	return func() { log.Printf("exit %s (%s)", msg, time.Since(start)) }
}
