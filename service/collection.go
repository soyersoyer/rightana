package service

import (
	"math/rand"
	"sort"
	"time"

	"github.com/soyersoyer/k20a/db/db"
	"github.com/soyersoyer/k20a/errors"
)

// Collection is the db's collection struct
type Collection = db.Collection

// CollectionDataInputT is the db's collectionDataInputT struct
type CollectionDataInputT = db.CollectionDataInputT

// CreateCollection creates a collection
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

// CreateCollectionByID creates a collection with a fixed ID
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

// CollectionReadAccessCheck checks the read access
func CollectionReadAccessCheck(collection *Collection, userEmail string) error {
	if collection.OwnerEmail != userEmail && !db.UserIsTeammate(collection, userEmail) {
		return errors.AccessDenied
	}
	return nil
}

// CollectionWriteAccessCheck checks the write access
func CollectionWriteAccessCheck(collection *Collection, userEmail string) error {
	if collection.OwnerEmail != userEmail {
		return errors.AccessDenied
	}
	return nil
}

// GetCollection fetch a collection by ID
func GetCollection(id string) (*Collection, error) {
	collection, err := db.GetCollection(id)
	if err != nil {
		return nil, errors.CollectionNotExist.T(id)
	}
	return collection, nil
}

// UpdateCollection updates the collection's name
func UpdateCollection(collection *Collection, name string) error {
	collection.Name = name
	if err := db.UpdateCollection(collection); err != nil {
		return errors.DBError.Wrap(err, collection)
	}
	return nil
}

// DeleteCollection deletes the collection
func DeleteCollection(collection *Collection) error {
	if err := db.DeleteCollection(collection); err != nil {
		return errors.DBError.Wrap(err, collection)
	}
	return nil
}

// CollectionSummaryT the struct for the Collection's summary
type CollectionSummaryT struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	PageviewCount   int     `json:"pageview_count"`
	PageviewPercent float32 `json:"pageview_percent"`
}

// GetCollectionSummariesByUserEmail returns the collection summaries for the user
func GetCollectionSummariesByUserEmail(email string) ([]CollectionSummaryT, error) {
	ret := []CollectionSummaryT{}
	collections, err := db.GetCollectionsByUserEmail(email)
	if err != nil {
		return nil, errors.DBError.Wrap(err, email)
	}
	for _, v := range collections {
		count, percent, err := db.GetPageviewPercent(v.ID, 7)
		if err != nil {
			return nil, errors.DBError.Wrap(err, v.ID)
		}
		ret = append(ret, CollectionSummaryT{
			ID:              v.ID,
			Name:            v.Name,
			PageviewCount:   count,
			PageviewPercent: percent,
		})
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})
	return ret, nil
}

// GetCollectionShards return the collection shards
func GetCollectionShards(collection *Collection) ([]db.ShardDataT, error) {
	shards, err := db.GetCollectionShardDatas(collection)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection)
	}
	return shards, nil
}

// DeleteCollectionShard deletes a shard
func DeleteCollectionShard(collection *Collection, shardID string) error {
	if err := db.DeleteCollectionShard(collection, shardID); err != nil {
		return errors.DBError.Wrap(err, shardID)
	}
	return nil
}

// AddTeammate adds a teammate to the collection
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

// RemoveTeammate removes the teammate from the collection
func RemoveTeammate(collection *Collection, email string) error {
	if coll := db.GetTeammate(collection, email); coll == nil {
		return errors.UserNotExist.T(email)
	}
	if err := db.RemoveTeammate(collection, email); err != nil {
		return errors.DBError.Wrap(err, collection, email)
	}
	return nil
}

// GetCollectionData returns the collection data
func GetCollectionData(collection *Collection, input *CollectionDataInputT) (*db.CollectionDataT, error) {
	data, err := db.GetBucketSums(collection, input)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection, input)
	}
	return data, nil
}

// GetCollectionStatData return the collection stats
func GetCollectionStatData(collection *Collection, input *CollectionDataInputT) (*db.CollectionStatDataT, error) {
	data, err := db.GetStatistics(collection, input)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection, input)
	}
	return data, nil
}

// GetSessions return the collection's sessions
func GetSessions(collection *Collection, input *CollectionDataInputT) ([]*db.SessionDataT, error) {
	data, err := db.GetSessions(collection, input)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection, input)
	}
	return data, nil
}

// GetPageviews return the pageviews for the collection
func GetPageviews(collection *Collection, sessionKey string) ([]*db.PageviewDataT, error) {
	key, err := db.DecodeSessionKey(sessionKey)
	if err != nil {
		return nil, errors.SessionNotExist.T(sessionKey).Wrap(err)
	}
	session, err := db.GetSession(collection.ID, key)
	if err != nil {
		return nil, errors.SessionNotExist.T(sessionKey).Wrap(err, collection.ID)
	}

	data, err := db.GetPageviews(collection, key)
	if err != nil {
		return nil, errors.DBError.Wrap(err, collection.ID, session)
	}
	return data, nil
}

// SeedCollection seed a collection with n sessions
func SeedCollection(from time.Time, to time.Time, collectionID string, n int) error {
	return db.Seed(from, to, collectionID, n)
}
