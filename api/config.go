package api

import (
	"net/http"

	"github.com/soyersoyer/rightana/config"
)

type publicConfigT struct {
	EnableRegistration bool `json:"enable_registration"`
}

func getPublicConfigE(w http.ResponseWriter, r *http.Request) error {
	return respond(w, publicConfigT{
		config.ActualConfig.EnableRegistration,
	})
}

var getPublicConfig = handleError(getPublicConfigE)
