package api

import (
	"net/http"

	"github.com/soyersoyer/rightana/config"
)

type publicConfigT struct {
	EnableRegistration bool   `json:"enable_registration"`
	TrackingID         string `json:"tracking_id"`
}

func getPublicConfigE(w http.ResponseWriter, r *http.Request) error {
	return respond(w, publicConfigT{
		config.ActualConfig.EnableRegistration,
		config.ActualConfig.TrackingID,
	})
}

var getPublicConfig = handleError(getPublicConfigE)
