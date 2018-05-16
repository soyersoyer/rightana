package api

import (
	"encoding/json"
	"net/http"

	"github.com/soyersoyer/rightana/errors"
	"github.com/soyersoyer/rightana/service"
)

type createSessionInputT struct {
	CollectionID     string `json:"c"`
	Hostname         string `json:"h"`
	BrowserLanguage  string `json:"bl"`
	ScreenResolution string `json:"sr"`
	WindowResolution string `json:"wr"`
	DeviceType       string `json:"dt"`
	Referrer         string `json:"r"`
}

func createSessionE(w http.ResponseWriter, r *http.Request) error {
	var input createSessionInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	sessionKey, err := service.CreateSession(r.UserAgent(), r.RemoteAddr, service.CreateSessionInputT(input))
	if err != nil {
		return err
	}

	return respond(w, sessionKey)
}

var createSession = handleError(createSessionE)

type updateSessionInputT struct {
	CollectionID string `json:"c"`
	SessionKey   string `json:"s"`
}

func updateSessionE(w http.ResponseWriter, r *http.Request) error {
	var input updateSessionInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	return service.UpdateSession(r.UserAgent(), input.CollectionID, input.SessionKey)
}

var updateSession = handleError(updateSessionE)

type createPageviewInputT struct {
	CollectionID string `json:"c"`
	SessionKey   string `json:"s"`
	Path         string `json:"p"`
}

func createPageviewE(w http.ResponseWriter, r *http.Request) error {
	var input createPageviewInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	return service.CreatePageview(r.UserAgent(), service.CreatePageviewInputT(input))
}

var createPageview = handleError(createPageviewE)
