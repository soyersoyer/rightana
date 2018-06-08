package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/soyersoyer/rightana/service"
)

func getBackupsE(w http.ResponseWriter, r *http.Request) error {
	return respond(w, service.GetBackups())
}

var getBackups = handleError(getBackupsE)

func runBackupE(w http.ResponseWriter, r *http.Request) error {
	backupID := chi.URLParam(r, "backupID")

	return service.RunBackup(backupID)
}

var runBackup = handleError(runBackupE)
