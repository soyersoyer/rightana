package db

import (
	"fmt"
)

// GetTeammate returns the teammate by ID
func GetTeammate(collection *Collection, ID uint64) *Teammate {
	idx := findTeammate(collection, ID)
	if idx == -1 {
		return nil
	}
	return collection.Teammates[idx]
}

// AddTeammate adds a Teammate to a user
func AddTeammate(collection *Collection, user *User) error {
	idx := findTeammate(collection, user.ID)
	if idx != -1 {
		return fmt.Errorf("teammate already added")
	}
	collection.Teammates = append(collection.Teammates, &Teammate{ID: user.ID})
	return UpdateCollection(collection)
}

// RemoveTeammate removes a teammate by email
func RemoveTeammate(collection *Collection, ID uint64) error {
	idx := findTeammate(collection, ID)
	if idx == -1 {
		return fmt.Errorf("teammate not found")
	}
	removeTeammateByIdx(collection, idx)
	return UpdateCollection(collection)
}

func findTeammate(collection *Collection, ID uint64) int {
	for k, v := range collection.Teammates {
		if v.ID == ID {
			return k
		}
	}
	return -1
}

func removeTeammateByIdx(collection *Collection, idx int) {
	cs := collection.Teammates
	collection.Teammates = append(cs[:idx], cs[idx+1:]...)
}
