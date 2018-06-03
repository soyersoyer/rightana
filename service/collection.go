package service

import (
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/soyersoyer/rightana/db/db"
	"github.com/soyersoyer/rightana/errors"
)

// Collection is the db's collection struct
type Collection = db.Collection

// CollectionDataInputT is the db's collectionDataInputT struct
type CollectionDataInputT = db.CollectionDataInputT

// CreateCollection creates a collection
func CreateCollection(ownerID uint64, name string) (*Collection, error) {
	user, err := db.GetUserByID(ownerID)
	if err != nil {
		return nil, errors.UserNotExist.T(string(ownerID)).Wrap(err)
	}
	if user.LimitCollections {
		collections, err := db.GetCollectionsByUserID(user.ID)
		if err != nil {
			return nil, errors.DBError.Wrap(err, user.ID)
		}
		if len(collections) >= int(user.CollectionLimit) {
			return nil, errors.CollectionLimitExceeded.T(strconv.Itoa(int(user.CollectionLimit)))
		}
	}
	collection := &Collection{
		ID:      randStringBytes(8),
		OwnerID: user.ID,
		Name:    name,
		Created: time.Now().UnixNano(),
	}
	err = db.InsertCollection(collection)
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
		ID:      id,
		OwnerID: user.ID,
		Name:    name,
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
func CollectionReadAccessCheck(collection *Collection, userID uint64) error {
	if collection.OwnerID != userID && !db.UserIsTeammate(collection, userID) {
		return errors.AccessDenied
	}
	return nil
}

// CollectionWriteAccessCheck checks the write access
func CollectionWriteAccessCheck(collection *Collection, userID uint64) error {
	if collection.OwnerID != userID {
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

// GetCollectionByName fetch a collection by name
func GetCollectionByName(user *db.User, name string) (*Collection, error) {
	collection, err := db.GetCollectionByName(user.ID, name)
	if err != nil {
		return nil, errors.CollectionNotExist.T(user.Name + "/" + name)
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
	User            string  `json:"user"`
	Name            string  `json:"name"`
	PageviewCount   int     `json:"pageview_count"`
	PageviewPercent float32 `json:"pageview_percent"`
}

// GetCollectionSummariesByUserID returns the collection summaries for the user
func GetCollectionSummariesByUserID(ID uint64) ([]CollectionSummaryT, error) {
	ret := []CollectionSummaryT{}
	collections, err := db.GetCollectionsByUserID(ID)
	if err != nil {
		return nil, errors.DBError.Wrap(err, ID)
	}
	for _, v := range collections {
		count, percent, err := db.GetPageviewPercent(v.ID, 7)
		if err != nil {
			return nil, errors.DBError.Wrap(err, v.ID)
		}
		user, err := db.GetUserByID(v.OwnerID)
		if err != nil {
			return nil, errors.DBError.Wrap(err, "can't get user", v.OwnerID)
		}
		ret = append(ret, CollectionSummaryT{
			ID:              v.ID,
			User:            user.Name,
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

// TeammateT contains the teammates information for the client
type TeammateT struct {
	Email string `json:"email"`
}

// AddTeammate adds a teammate to the collection
func AddTeammate(collection *Collection, input TeammateT) error {
	user, err := db.GetUserByEmail(input.Email)
	if err != nil {
		return errors.UserNotExist.T(input.Email).Wrap(err)
	}
	if coll := db.GetTeammate(collection, user.ID); coll != nil {
		return errors.TeammateExist.T(input.Email)
	}
	if err := db.AddTeammate(collection, user); err != nil {
		return errors.DBError.Wrap(err, collection.ID, input.Email)
	}
	return nil
}

// RemoveTeammate removes the teammate from the collection
func RemoveTeammate(collection *Collection, email string) error {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return errors.UserNotExist.T(email).Wrap(err)
	}
	if coll := db.GetTeammate(collection, user.ID); coll == nil {
		return errors.UserNotExist.T(email)
	}
	if err := db.RemoveTeammate(collection, user.ID); err != nil {
		return errors.DBError.Wrap(err, collection, email)
	}
	return nil
}

// GetCollectionTeammates returns the teammates emails
func GetCollectionTeammates(collection *Collection) ([]TeammateT, error) {
	tms := []TeammateT{}
	for _, v := range collection.Teammates {
		user, err := GetUserByID(v.ID)
		if err != nil {
			return nil, errors.DBError.T(string(v.ID)).Wrap(err)
		}
		tms = append(tms, TeammateT{user.Email})
	}
	return tms, nil
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
