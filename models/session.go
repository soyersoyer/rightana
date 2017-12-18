package models

import (
	"encoding/base64"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/mssola/user_agent"

	"github.com/soyersoyer/k20a/db/db"
	"github.com/soyersoyer/k20a/db/shardbolt"
	"github.com/soyersoyer/k20a/errors"
	"github.com/soyersoyer/k20a/geoip"
)

type CreateSessionInputT struct {
	CollectionID     string
	Hostname         string
	BrowserLanguage  string
	ScreenResolution string
	WindowResolution string
	DeviceType       string
}

func CreateSession(userAgent string, remoteAddr string, input CreateSessionInputT) (string, error) {
	now := time.Now()
	ua := user_agent.New(userAgent)

	if ua.Bot() {
		return "", errors.BotsDontMatter
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
			return "", errors.CollectionNotExist.T(input.CollectionID).Wrap(err)
		}
		return "", errors.DBError.Wrap(err, input.CollectionID)
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
		End:              now.UnixNano(),
		PageviewCount:    0,
	}
	key := db.GetKey(now, rand.Uint32())
	if err := db.ShardUpsert(collection.ID, key, session); err != nil {
		return "", errors.DBError.Wrap(err, session)
	}
	sessionKey := db.EncodeSessionKey(key)
	return sessionKey, nil
}

func UpdateSession(userAgent string, CollectionID string, sessionKey string) error {
	ua := user_agent.New(userAgent)

	if ua.Bot() {
		return errors.BotsDontMatter
	}

	key, err := db.DecodeSessionKey(sessionKey)
	if err != nil {
		return errors.SessionNotExist.T(sessionKey).Wrap(err)
	}
	session, err := db.GetSession(CollectionID, key)
	if err != nil {
		return errors.SessionNotExist.T(sessionKey).Wrap(err, CollectionID)
	}
	session.End = time.Now().UnixNano()

	if err := db.ShardUpsert(CollectionID, key, session); err != nil {
		return errors.DBError.Wrap(err, CollectionID, key, session)
	}
	return nil
}

type CreatePageviewInputT struct {
	CollectionID string
	SessionKey   string
	Path         string
	Referrer     string
}

func CreatePageview(userAgent string, input CreatePageviewInputT) error {
	now := time.Now()
	ua := user_agent.New(userAgent)

	if ua.Bot() {
		return errors.BotsDontMatter
	}

	key, err := base64.StdEncoding.DecodeString(input.SessionKey)
	if err != nil {
		return errors.SessionNotExist.T(input.SessionKey).Wrap(err, input.SessionKey)
	}
	session, err := db.GetSession(input.CollectionID, key)
	if err != nil {
		return errors.SessionNotExist.T(input.SessionKey).Wrap(err, input.CollectionID)
	}
	session.End = now.UnixNano()
	session.PageviewCount = session.PageviewCount + 1

	sID := db.GetIDFromKey(key)
	pKey := db.GetKey(now, sID)

	pageview := &db.Pageview{
		Path:        input.Path,
		ReferrerURL: input.Referrer,
	}

	err = db.ShardUpdate(input.CollectionID, func(tx *shardbolt.MultiTx) error {
		if err := db.ShardUpsertTx(tx, key, session); err != nil {
			return err
		}

		return db.ShardUpsertTx(tx, pKey, pageview)
	})
	if err != nil {
		return errors.DBError.Wrap(err, input, key, session, pKey, pageview)
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
