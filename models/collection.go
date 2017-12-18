package models

import (
	"math/rand"
	"sort"
	"time"

	"github.com/soyersoyer/k20a/db/db"
	"github.com/soyersoyer/k20a/errors"
)

type Collection = db.Collection
type CollectionDataInputT = db.CollectionDataInputT

func CreateCollection(ownerEmail string, name string) (*Collection, error) {
	collection := &Collection{
		ID:         "K20A-" + randStringBytes(8),
		OwnerEmail: ownerEmail,
		Name:       name,
	}
	err := db.InsertCollection(collection)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection)
	}
	return collection, nil
}

func CreateCollectionByID(id string, name string, ownerEmail string) (*Collection, error) {
	user, err := GetUserByEmail(ownerEmail)
	if err != nil {
		return nil, err
	}
	collection := &Collection{
		ID:         id,
		OwnerEmail: user.Email,
		Name:       name,
	}
	if err := db.InsertCollection(collection); err != nil {
		if err != nil {
			return nil, errors.DBError.Wrap(err, collection)
		}
	}
	return collection, nil
}

var src = rand.NewSource(time.Now().UnixNano())
var myRand = rand.New(src)

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[myRand.Intn(len(letterBytes))]
	}
	return string(b)
}

func CollectionReadAccessCheck(collection *Collection, userEmail string) error {
	if collection.OwnerEmail != userEmail && !db.UserIsTeammate(collection, userEmail) {
		return errors.AccessDenied
	}
	return nil
}

func CollectionWriteAccessCheck(collection *Collection, userEmail string) error {
	if collection.OwnerEmail != userEmail {
		return errors.AccessDenied
	}
	return nil
}

func GetCollection(id string) (*Collection, error) {
	collection, err := db.GetCollection(id)
	if err != nil {
		return nil, errors.CollectionNotExist.T(id)
	}
	return collection, nil
}

func UpdateCollection(collection *Collection, name string) error {
	collection.Name = name
	if err := db.UpdateCollection(collection); err != nil {
		return errors.DBError.Wrap(err, collection)
	}
	return nil
}

func DeleteCollection(collection *Collection) error {
	if err := db.DeleteCollection(collection); err != nil {
		return errors.DBError.Wrap(err, collection)
	}
	return nil
}

type CollectionSummaryT struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	PageviewPercent float32 `json:"pageview_percent"`
}

func GetCollectionSummaryByUserEmail(email string) ([]CollectionSummaryT, error) {
	ret := []CollectionSummaryT{}
	collections, err := db.GetCollectionsByUserEmail(email)
	if err != nil {
		return nil, errors.DBError.Wrap(err, email)
	}
	for _, v := range collections {
		percent, err := db.GetPageviewPercent(v.ID, 7)
		if err != nil {
			return nil, errors.DBError.Wrap(err, v.ID)
		}
		ret = append(ret, CollectionSummaryT{
			ID:              v.ID,
			Name:            v.Name,
			PageviewPercent: percent,
		})
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})
	return ret, nil
}

func GetCollectionShards(collection *Collection) ([]db.Shard, error) {
	shards, err := db.GetCollectionShards(collection)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection)
	}
	return shards, nil
}

func DeleteCollectionShard(collection *Collection, shardID string) error {
	if err := db.DeleteCollectionShard(collection, shardID); err != nil {
		return errors.DBError.Wrap(err, shardID)
	}
	return nil
}

func AddTeammate(collection *Collection, email string) error {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return errors.UserNotExist.T(email).Wrap(err)
	}
	if coll := db.GetTeammate(collection, email); coll != nil {
		return errors.TeammateExist.T(email)
	}
	if err := db.AddTeammate(collection, user); err != nil {
		return errors.DBError.Wrap(err, collection.ID, email)
	}
	return nil
}

func RemoveTeammate(collection *Collection, email string) error {
	if coll := db.GetTeammate(collection, email); coll == nil {
		return errors.UserNotExist.T(email)
	}
	if err := db.RemoveTeammate(collection, email); err != nil {
		return errors.DBError.Wrap(err, collection, email)
	}
	return nil
}

func GetCollectionData(collection *Collection, input *CollectionDataInputT) (*db.CollectionDataT, error) {
	data, err := db.GetBucketSums(collection, input)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection, input)
	}
	return data, nil
}

func GetCollectionStatData(collection *Collection, input *CollectionDataInputT) (*db.CollectionStatDataT, error) {
	data, err := db.GetStatistics(collection, input)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection, input)
	}
	return data, nil
}

func GetSessions(collection *Collection, input *CollectionDataInputT) ([]*db.SessionDataT, error) {
	data, err := db.GetSessions(collection, input)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection, input)
	}
	return data, nil
}

func GetPageviews(collection *Collection, sessionKey string) ([]*db.PageviewDataT, error) {
	key, err := db.DecodeSessionKey(sessionKey)
	if err != nil {
		return nil, errors.SessionNotExist.T(sessionKey).Wrap(err)
	}
	session, err := db.GetSession(collection.ID, key)
	if err != nil {
		return nil, errors.SessionNotExist.T(sessionKey).Wrap(err, collection.ID)
	}

	data, err := db.GetPageviews(collection, key, session)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection.ID, session)
	}
	return data, nil
}

func SeedCollection(from time.Time, to time.Time, collectionID string, n int) error {
	return db.Seed(from, to, collectionID, n)
}
