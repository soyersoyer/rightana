package db

import (
	"fmt"
)

func GetTeammate(collection *Collection, email string) *Teammate {
	idx := findTeammate(collection, email)
	if idx == -1 {
		return nil
	}
	return collection.Teammates[idx]
}

func AddTeammate(collection *Collection, user *User) error {
	idx := findTeammate(collection, user.Email)
	if idx != -1 {
		return fmt.Errorf("teammate already added")
	}
	collection.Teammates = append(collection.Teammates, &Teammate{Email: user.Email})
	return UpdateCollection(collection)
}

func RemoveTeammate(collection *Collection, email string) error {
	idx := findTeammate(collection, email)
	if idx == -1 {
		return fmt.Errorf("teammate not found")
	}
	removeTeammateByIdx(collection, idx)
	return UpdateCollection(collection)
}

func findTeammate(collection *Collection, email string) int {
	for k, v := range collection.Teammates {
		if v.Email == email {
			return k
		}
	}
	return -1
}

func removeTeammateByIdx(collection *Collection, idx int) {
	cs := collection.Teammates
	collection.Teammates = append(cs[:idx], cs[idx+1:]...)
}
