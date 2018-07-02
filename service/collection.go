package service

import (
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/soyersoyer/rightana/db/db"
)

// Collection is the db's collection struct
type Collection = db.Collection

// CollectionDataInputT is the db's collectionDataInputT struct
type CollectionDataInputT = db.CollectionDataInputT

// CreateCollection creates a collection
func CreateCollection(ownerID uint64, name string) (*Collection, error) {
	id := randStringBytes(8)
	user, err := db.GetUserByID(ownerID)
	if err != nil {
		return nil, ErrUserNotExist.T(string(ownerID)).Wrap(err)
	}
	return createCollection(id, name, user)
}

// CreateCollectionByID creates a collection with a fixed ID
func CreateCollectionByID(id string, name string, username string) (*Collection, error) {
	user, err := GetUserByName(username)
	if err != nil {
		return nil, err
	}
	return createCollection(id, name, user)
}

func createCollection(id string, name string, user *User) (*Collection, error) {
	if user.LimitCollections {
		collections, err := db.GetCollectionsByUserID(user.ID)
		if err != nil {
			return nil, ErrDB.Wrap(err, user.ID)
		}
		if len(collections) >= int(user.CollectionLimit) {
			return nil, ErrCollectionLimitExceeded.T(strconv.Itoa(int(user.CollectionLimit)))
		}
	}
	collection := &Collection{
		ID:      randStringBytes(8),
		OwnerID: user.ID,
		Name:    name,
		Created: time.Now().UnixNano(),
	}
	if err := validateCollection(collection); err != nil {
		return nil, err
	}
	err := db.InsertCollection(collection)
	if err != nil {
		return nil, ErrDB.Wrap(err, collection)
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
func CollectionReadAccessCheck(collection *Collection, user *User) error {
	if collection.OwnerID == user.ID || db.UserIsTeammate(collection, user.ID) || user.IsAdmin {
		return nil
	}
	return ErrAccessDenied
}

// CollectionWriteAccessCheck checks the write access
func CollectionWriteAccessCheck(collection *Collection, user *User) error {
	if collection.OwnerID == user.ID || user.IsAdmin {
		return nil
	}
	return ErrAccessDenied
}

// CollectionCreateAccessCheck checks the create access
func CollectionCreateAccessCheck(user *User, loggedInUser *User) error {
	if loggedInUser.IsAdmin || user.ID == loggedInUser.ID {
		return nil
	}
	return ErrAccessDenied
}

// GetCollection fetch a collection by ID
func GetCollection(id string) (*Collection, error) {
	collection, err := db.GetCollection(id)
	if err != nil {
		return nil, ErrCollectionNotExist.T(id)
	}
	return collection, nil
}

// GetCollectionByName fetch a collection by name
func GetCollectionByName(user *db.User, name string) (*Collection, error) {
	collection, err := db.GetCollectionByName(user.ID, name)
	if err != nil {
		return nil, ErrCollectionNotExist.T(user.Name + "/" + name)
	}
	return collection, nil
}

// UpdateCollection updates the collection's name
func UpdateCollection(collection *Collection, name string) error {
	collection.Name = name
	if err := validateCollection(collection); err != nil {
		return err
	}
	if err := db.UpdateCollection(collection); err != nil {
		return ErrDB.Wrap(err, collection)
	}
	return nil
}

// DeleteCollection deletes the collection
func DeleteCollection(collection *Collection) error {
	if err := db.DeleteCollection(collection); err != nil {
		return ErrDB.Wrap(err, collection)
	}
	return nil
}

// CollectionSummaryT the struct for the Collection's summary
type CollectionSummaryT struct {
	ID   string `json:"id"`
	User string `json:"user"`
	Name string `json:"name"`
	db.CollectionSummary
}

//CollectionSummaryOptions contains options for the db.GetCollectionSummary function
type CollectionSummaryOptions = db.CollectionSummaryOptions

// GetCollectionSummariesByUserID returns the collection summaries for the user
func GetCollectionSummariesByUserID(ID uint64, readerID uint64, options CollectionSummaryOptions) ([]CollectionSummaryT, error) {
	ret := []CollectionSummaryT{}
	collections, err := db.GetCollectionsByUserID(ID)
	if err != nil {
		return nil, ErrDB.Wrap(err, ID)
	}
	for _, v := range collections {
		if v.OwnerID != readerID && !db.UserIsTeammate(&v, readerID) {
			continue
		}
		cs, err := db.GetCollectionSummary(v.ID, 7, options)
		if err != nil {
			return nil, ErrDB.Wrap(err, v.ID)
		}
		user, err := db.GetUserByID(v.OwnerID)
		if err != nil {
			return nil, ErrDB.Wrap(err, "can't get user", v.OwnerID)
		}
		ret = append(ret, CollectionSummaryT{
			ID:                v.ID,
			User:              user.Name,
			Name:              v.Name,
			CollectionSummary: cs,
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
		return nil, ErrDB.Wrap(err, collection)
	}
	return shards, nil
}

// DeleteCollectionShard deletes a shard
func DeleteCollectionShard(collection *Collection, shardID string) error {
	if err := db.DeleteCollectionShard(collection, shardID); err != nil {
		return ErrDB.Wrap(err, shardID)
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
		return ErrUserNotExist.T(input.Email).Wrap(err)
	}
	if coll := db.GetTeammate(collection, user.ID); coll != nil {
		return ErrTeammateExist.T(input.Email)
	}
	if err := db.AddTeammate(collection, user); err != nil {
		return ErrDB.Wrap(err, collection.ID, input.Email)
	}
	return nil
}

// RemoveTeammate removes the teammate from the collection
func RemoveTeammate(collection *Collection, email string) error {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return ErrUserNotExist.T(email).Wrap(err)
	}
	if coll := db.GetTeammate(collection, user.ID); coll == nil {
		return ErrUserNotExist.T(email)
	}
	if err := db.RemoveTeammate(collection, user.ID); err != nil {
		return ErrDB.Wrap(err, collection, email)
	}
	return nil
}

// GetCollectionTeammates returns the teammates emails
func GetCollectionTeammates(collection *Collection) ([]TeammateT, error) {
	tms := []TeammateT{}
	for _, v := range collection.Teammates {
		user, err := GetUserByID(v.ID)
		if err != nil {
			return nil, ErrDB.T(string(v.ID)).Wrap(err)
		}
		tms = append(tms, TeammateT{user.Email})
	}
	return tms, nil
}

// GetCollectionData returns the collection data
func GetCollectionData(collection *Collection, input *CollectionDataInputT) (*db.CollectionDataT, error) {
	data, err := db.GetBucketSums(collection, input)
	if err != nil {
		return nil, ErrDB.Wrap(err, collection, input)
	}
	return data, nil
}

// GetCollectionStatData return the collection stats
func GetCollectionStatData(collection *Collection, input *CollectionDataInputT) (*db.CollectionStatDataT, error) {
	data, err := db.GetStatistics(collection, input)
	if err != nil {
		return nil, ErrDB.Wrap(err, collection, input)
	}
	return data, nil
}

// GetSessions return the collection's sessions
func GetSessions(collection *Collection, input *CollectionDataInputT) ([]*db.SessionDataT, error) {
	data, err := db.GetSessions(collection, input)
	if err != nil {
		return nil, ErrDB.Wrap(err, collection, input)
	}
	return data, nil
}

// GetPageviews return the pageviews for the collection
func GetPageviews(collection *Collection, sessionKey string) ([]*db.PageviewDataT, error) {
	key, err := db.DecodeSessionKey(sessionKey)
	if err != nil {
		return nil, ErrSessionNotExist.T(sessionKey).Wrap(err)
	}
	session, err := db.GetSession(collection.ID, key)
	if err != nil {
		return nil, ErrSessionNotExist.T(sessionKey).Wrap(err, collection.ID)
	}

	data, err := db.GetPageviews(collection, key)
	if err != nil {
		return nil, ErrDB.Wrap(err, collection.ID, session)
	}
	return data, nil
}

// SeedCollection seed a collection with n sessions
func SeedCollection(from time.Time, to time.Time, collectionID string, n int) error {
	return db.Seed(from, to, collectionID, n)
}

var collectionRegexp = regexp.MustCompile("^[a-z0-9.]+$")

func validateCollection(c *Collection) error {
	if !collectionRegexp.MatchString(c.Name) {
		return ErrInvalidCollectionName.T(c.Name)
	}

	aColl, err := db.GetCollectionByName(c.OwnerID, c.Name)
	if err != nil && err != db.ErrKeyNotExists {
		return ErrDB.T(c.Name).Wrap(err)
	}
	if err == nil && aColl.ID != c.ID {
		return ErrCollectionNameExist.T(c.Name)
	}

	return nil
}
