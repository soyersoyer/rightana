package db

import (
	"log"
	"os"
	"testing"
	"time"
)

var (
	dir          = "data"
	email        = "soyer@irl.hu"
	collectionID = "AAAA"
	from         = time.Now()
	to           = from.Add(time.Duration(10000) * time.Hour)
	collection   = Collection{
		ID:      collectionID,
		Name:    "test.org",
		OwnerID: 1,
	}
)

func TestMain(m *testing.M) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Mkdir(dir, 0700)
	if err != nil {
		log.Fatalln(err)
	}

	ret := m.Run()
	if ret == 0 {
		os.RemoveAll(dir)
	}

	os.Exit(ret)
}

func TestUserCreate(t *testing.T) {
	InitDatabase(dir)

	user := User{
		Email:    email,
		Password: "e!",
	}

	if err := InsertUser(&user); err != nil {
		t.Error(err)
	}

	user2, err := GetUserByEmail(email)
	if err != nil {
		t.Error(err)
	}
	if user.Email != user2.Email || user.Password != user2.Password {
		t.Error(user, user2)
	}
}

func TestCollectionCreate(t *testing.T) {
	if err := InsertCollection(&collection); err != nil {
		t.Error(err)
	}
}

func TestCollectionGetByID(t *testing.T) {
	if _, err := GetCollection(collection.ID); err != nil {
		t.Error(err)
	}
}

func TestCollectionGetByName(t *testing.T) {
	if _, err := GetCollectionByName(collection.OwnerID, collection.Name); err != nil {
		t.Error(err)
	}
}

func TestSeed(t *testing.T) {
	Seed(from, to, collection.ID, 100000)
}

func TestStat(t *testing.T) {
	input := CollectionDataInputT{
		From:   from,
		To:     to,
		Bucket: "hour",
	}

	start := time.Now()
	_, err := GetBucketSums(&collection, &input)
	if err != nil {
		t.Error(err)
	}
	elapsed := time.Since(start)
	log.Printf("bucketsums time: %s", elapsed)

	start = time.Now()
	_, err = GetStatistics(&collection, &input)
	if err != nil {
		t.Error(err)
	}
	elapsed = time.Since(start)
	log.Printf("stat time: %s", elapsed)
}
