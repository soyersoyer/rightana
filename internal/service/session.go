package service

import (
	"encoding/base64"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/mssola/user_agent"

	"github.com/soyersoyer/rightana/internal/db"
	"github.com/soyersoyer/rightana/internal/geoip"
)

// CreateSessionInputT is the struct for creating a session
type CreateSessionInputT struct {
	CollectionID     string
	Hostname         string
	BrowserLanguage  string
	ScreenResolution string
	WindowResolution string
	DeviceType       string
	Referrer         string
}

// CreateSession creates a session
func CreateSession(userAgent string, remoteAddr string, input CreateSessionInputT) (string, error) {
	now := time.Now()
	ua := user_agent.New(userAgent)

	if ua.Bot() {
		return "", ErrBotsDontMatter
	}

	ip, err := getIP(remoteAddr)
	if err != nil {
		return "", err
	}

	browserName, browserVersion := ua.Browser()

	location := geoip.LocationByIP(ip)
	asn := geoip.ASNByIP(ip)

	userHostname := ""
	userHostnames, _ := net.LookupAddr(ip)
	if len(userHostnames) > 0 {
		userHostname = userHostnames[0]
	}

	collection, err := db.GetCollection(input.CollectionID)
	if err != nil {
		if err == db.ErrKeyNotExists {
			return "", ErrCollectionNotExist.T(input.CollectionID).Wrap(err)
		}
		return "", ErrDB.Wrap(err, input.CollectionID)
	}

	session := &db.Session{
		Hostname:         input.Hostname,
		UserIP:           ip,
		UserHostname:     userHostname,
		DeviceOS:         ua.OS(),
		BrowserName:      browserName,
		BrowserVersion:   browserVersion,
		BrowserLanguage:  input.BrowserLanguage,
		ScreenResolution: input.ScreenResolution,
		WindowResolution: input.WindowResolution,
		DeviceType:       input.DeviceType,
		CountryCode:      location.CountryCode,
		City:             location.City,
		ASNumber:         int32(asn.Number),
		ASName:           asn.Name,
		UserAgent:        userAgent,
		Duration:         0,
		Referrer:         input.Referrer,
	}
	key := db.GetKey(now, rand.Uint32())
	if err := db.ShardUpsertBatch(collection.ID, key, session); err != nil {
		return "", ErrDB.Wrap(err, session)
	}
	sessionKey := db.EncodeSessionKey(key)
	return sessionKey, nil
}

// UpdateSession updates the session.End field
func UpdateSession(userAgent string, CollectionID string, sessionKey string) error {
	ua := user_agent.New(userAgent)

	if ua.Bot() {
		return ErrBotsDontMatter
	}

	key, err := db.DecodeSessionKey(sessionKey)
	if err != nil {
		return ErrSessionNotExist.T(sessionKey).Wrap(err)
	}
	session, err := db.GetSession(CollectionID, key)
	if err != nil {
		return ErrSessionNotExist.T(sessionKey).Wrap(err, CollectionID)
	}
	sessionBegin := db.GetTimeFromKey(key)
	session.Duration = int32(time.Now().Sub(sessionBegin).Seconds())

	if err := db.ShardUpsertBatch(CollectionID, key, session); err != nil {
		return ErrDB.Wrap(err, CollectionID, key, session)
	}
	return nil
}

// CreatePageviewInputT is the input for the CreatePageView
type CreatePageviewInputT struct {
	CollectionID string
	SessionKey   string
	Path         string
}

// CreatePageview creates a pageview
func CreatePageview(userAgent string, input CreatePageviewInputT) error {
	now := time.Now()
	ua := user_agent.New(userAgent)

	if ua.Bot() {
		return ErrBotsDontMatter
	}

	sessKey, err := base64.StdEncoding.DecodeString(input.SessionKey)
	if err != nil {
		return ErrSessionNotExist.T(input.SessionKey).Wrap(err, input.SessionKey)
	}
	_, err = db.GetSession(input.CollectionID, sessKey)
	if err != nil {
		return ErrSessionNotExist.T(input.SessionKey).Wrap(err, input.CollectionID)
	}

	pvKey := db.GetPVKey(sessKey, now)

	path, queryString := splitURL(input.Path)

	pageview := &db.Pageview{
		Path:        path,
		QueryString: queryString,
	}

	if err := db.ShardUpsertBatch(input.CollectionID, pvKey, pageview); err != nil {
		return ErrDB.Wrap(err, input)
	}
	return nil
}

func getIP(remoteAddr string) (string, error) {
	if i := strings.IndexRune(remoteAddr, ':'); i < 0 {
		return remoteAddr, nil
	}
	ip, _, err := net.SplitHostPort(remoteAddr)
	return ip, err
}

func splitURL(url string) (path, queryString string) {
	idx := strings.IndexAny(url, "?;")
	if idx < 0 {
		return url, ""
	}
	return url[:idx], url[idx+1:]
}
