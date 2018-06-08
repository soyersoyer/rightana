package service

import (
	"github.com/soyersoyer/rightana/config"
	"github.com/soyersoyer/rightana/db/db"
)

// Backup stores the backup's properties
type Backup struct {
	ID  string `json:"id"`
	Dir string `json:"dir"`
}

// RunBackup runs the backup
func RunBackup(backupID string) error {
	dir, ok := config.ActualConfig.Backup[backupID]
	if !ok {
		return ErrBackupNotExist.T(backupID)
	}
	if err := db.RunBackup(dir); err != nil {
		return ErrDB.Wrap(err).T(dir)
	}
	return nil
}

// GetBackups returns the backup configuration
func GetBackups() []Backup {
	backups := []Backup{}
	for k, v := range config.ActualConfig.Backup {
		backups = append(backups, Backup{k, v})
	}
	return backups
}
