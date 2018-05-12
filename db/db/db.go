package db

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"

	"github.com/soyersoyer/k20a/db/cipobolt"
	"github.com/soyersoyer/k20a/db/shardbolt"
)

type shardMap map[string]*shardbolt.DB

var (
	basedir  = "data/"
	filename = "k20a.bolt"

	cipo     *cipobolt.DB
	shardDBs = atomic.Value{}
	dbMutex  sync.Mutex

	// ErrKeyExists is an error what you can get if the key exists but it shouldn't
	ErrKeyExists = cipobolt.ErrKeyExists
	// ErrKeyNotExists is an error what you can get if the key not exists but it should
	ErrKeyNotExists = cipobolt.ErrKeyNotExists
)

func map2Month(key []byte) string {
	t, err := unmarshalTime(key)
	if err != nil {
		panic(err)
	}
	return t.Format("2006-01")
}

// InitDatabase initializes the databases, creates the directories if necessary
func InitDatabase(basedirParam string) {
	basedir = path.Clean(basedirParam) + "/"
	os.MkdirAll(basedir, os.ModePerm)
	bdb, err := bolt.Open(basedir+filename, 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	cipo = cipobolt.Open(bdb, protoEncode, protoDecode, bucketName)
	shardDBs.Store(shardMap{})
}

// UpdateUser updates an user
func UpdateUser(user *User) error {
	return cipo.Update(user.Email, user)
}

// InsertUser inserts an user
func InsertUser(user *User) error {
	return cipo.Insert(user.Email, user)
}

// UpsertUser upserts an user
func UpsertUser(user *User) error {
	return cipo.Upsert(user.Email, user)
}

// GetUserByEmail returns an user with the email parameter
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := cipo.Get(email, user)
	return user, err
}

// DeleteUser deletes an user
func DeleteUser(user *User) error {
	return cipo.Bolt().Update(func(tx *bolt.Tx) error {
		if err := cipo.DeleteTx(tx, user.Email, user); err != nil {
			return err
		}

		if err := deleteAuthTokensByUserEmailTx(tx, user.Email); err != nil {
			return err
		}
		if err := deleteCollectionsByUserEmailTx(tx, user.Email); err != nil {
			return err
		}
		return deleteTeammateByUserEmailTx(tx, user.Email)
	})
}

func deleteAuthTokensByUserEmailTx(tx *bolt.Tx, email string) error {
	key := ""
	v := AuthToken{}
	return cipo.IterateTx(tx, &key, &v, func() error {
		if v.OwnerEmail == email {
			return cipo.DeleteTx(tx, key, &v)
		}
		return nil
	})
}

func deleteCollectionsByUserEmailTx(tx *bolt.Tx, email string) error {
	key := ""
	v := Collection{}
	return cipo.IterateTx(tx, &key, &v, func() error {
		if v.OwnerEmail == email {
			return deleteCollectionTx(tx, &v)
		}
		return nil
	})
}

func deleteTeammateByUserEmailTx(tx *bolt.Tx, email string) error {
	key := ""
	v := Collection{}
	return cipo.IterateTx(tx, &key, &v, func() error {
		idx := findTeammate(&v, email)
		if idx == -1 {
			return nil
		}
		removeTeammateByIdx(&v, idx)
		return updateCollectionTx(tx, &v)
	})
}

// InsertAuthToken inserts an authtoken
func InsertAuthToken(token *AuthToken) error {
	token.TTL = 1209600
	token.Created = time.Now().UnixNano()
	return cipo.Insert(token.ID, token)
}

// GetAuthToken returns an authtoken with the id parameter
func GetAuthToken(id string) (*AuthToken, error) {
	token := &AuthToken{}
	err := cipo.Get(id, token)
	return token, err
}

// UpdateAuthToken updates an authtoken
func UpdateAuthToken(token *AuthToken) error {
	return cipo.Update(token.ID, token)
}

// DeleteAuthToken deletes an authtoken
func DeleteAuthToken(id string) error {
	return cipo.Delete(id, &AuthToken{})
}

// InsertCollection inserts a new collection
func InsertCollection(collection *Collection) error {
	return cipo.Insert(collection.ID, collection)
}

// UpdateCollection updates a new collection
func UpdateCollection(collection *Collection) error {
	return cipo.Update(collection.ID, collection)
}

func updateCollectionTx(tx *bolt.Tx, collection *Collection) error {
	return cipo.UpdateTx(tx, collection.ID, collection)
}

// GetCollection returns a collection with the id parameter
func GetCollection(id string) (*Collection, error) {
	collection := Collection{}
	err := cipo.Get(id, &collection)
	return &collection, err
}

// DeleteCollection deletes a collection
func DeleteCollection(collection *Collection) error {
	return cipo.Bolt().Update(func(tx *bolt.Tx) error {
		return deleteCollectionTx(tx, collection)
	})
}

func deleteCollectionTx(tx *bolt.Tx, collection *Collection) error {
	if err := cipo.DeleteTx(tx, collection.ID, collection); err != nil {
		return err
	}
	return deleteShardDB(collection.ID)
}

// GetCollectionsByUserEmail returns collections for the usser with the email address
func GetCollectionsByUserEmail(email string) ([]Collection, error) {
	key := ""
	collection := Collection{}
	collections := []Collection{}
	err := cipo.Iterate(&key, &collection, func() error {
		if collection.OwnerEmail == email || UserIsTeammate(&collection, email) {
			collections = append(collections, collection)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return collections, nil
}

// UserIsTeammate check whether the User inside the collection's team
func UserIsTeammate(collection *Collection, email string) bool {
	for _, v := range collection.Teammates {
		if v.Email == email {
			return true
		}
	}
	return false
}

// ShardDataT is the shard data struct for the clients
type ShardDataT struct {
	ID   string `json:"id"`
	Size int    `json:"size"`
}

// GetCollectionShardDatas returns the collection's shards information
func GetCollectionShardDatas(collection *Collection) ([]ShardDataT, error) {
	db, err := getShardDB(collection.ID)
	if err != nil {
		return nil, err
	}
	ret := []ShardDataT{}
	for _, v := range db.GetSizes() {
		ret = append(ret, ShardDataT(v))
	}
	return ret, nil
}

// DeleteCollectionShard deletes a shard from the collection
func DeleteCollectionShard(collection *Collection, shardID string) error {
	db, err := getShardDB(collection.ID)
	if err != nil {
		return err
	}
	return db.DeleteShard(shardID)
}

// GetSession returns a session
func GetSession(collectionID string, key []byte) (*Session, error) {
	db, err := getShardDB(collectionID)
	if err != nil {
		return nil, err
	}
	session := &Session{}
	buf, err := db.Get(BSession, key)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(buf, session)
	return session, err
}

// GetPageviewPercent returns the last week versus the before last week difference in percent
func GetPageviewPercent(collectionID string, dayBefore int) (float32, error) {
	now := time.Now()
	n7dAgo := now.AddDate(0, 0, -dayBefore)
	n14dAgo := n7dAgo.AddDate(0, 0, -dayBefore)

	nowK := GetKeyFromTime(now)
	n7dAgoK := GetKeyFromTime(n7dAgo)
	n14dAgoK := GetKeyFromTime(n14dAgo)

	sdb, err := getShardDB(collectionID)
	if err != nil {
		return 0, err
	}
	sumFirst := 0
	sdb.Iterate(BPageview, n14dAgoK, n7dAgoK, func(k []byte, v []byte) {
		sumFirst++
	})

	sumSecond := 0
	sdb.Iterate(BPageview, n7dAgoK, nowK, func(k []byte, v []byte) {
		sumSecond++
	})
	percent := float32(0.0)
	if sumFirst != 0 {
		percent = float32(sumSecond)/float32(sumFirst) - 1.0
	}
	return percent, nil
}

func getShardDB(collectionID string) (*shardbolt.DB, error) {
	dbs := shardDBs.Load().(shardMap)
	db, ok := dbs[collectionID]
	if !ok {
		dbMutex.Lock()
		defer dbMutex.Unlock()
		dbs := shardDBs.Load().(shardMap)
		db2, ok := dbs[collectionID]
		if ok {
			log.Println(db2)
			return db2, nil
		}

		if collectionID == "" {
			return nil, fmt.Errorf("collectionId is empty")
		}
		var err error
		db, err = shardbolt.Open(basedir+collectionID, map2Month, 0666, nil)
		if err != nil {
			return nil, err
		}
		dbsCopy := shardMap{}
		for k, v := range dbs {
			dbsCopy[k] = v
		}
		dbsCopy[collectionID] = db
		shardDBs.Store(dbsCopy)
	}
	return db, nil
}

func deleteShardDB(collectionID string) error {
	sdb, err := getShardDB(collectionID)
	if err != nil {
		return err
	}
	errs := sdb.Close()
	if errs != nil {
		log.Println(errs)
		return fmt.Errorf("can't close the shard db %v", errs)
	}
	if err := os.RemoveAll(basedir + collectionID); err != nil {
		return err
	}

	dbs := shardDBs.Load().(shardMap)
	dbsCopy := shardMap{}
	for k, v := range dbs {
		dbsCopy[k] = v
	}
	delete(dbsCopy, collectionID)
	shardDBs.Store(dbsCopy)
	return nil
}

// GetKey returns the databse key based on time and id
func GetKey(t time.Time, id uint32) []byte {
	return append(marshalTime(t), marshal(id)...)
}

// GetKeyFromTime returns a database key based on time
func GetKeyFromTime(t time.Time) []byte {
	return marshalTime(t)
}

// GetTimeFromKey returns the time based on key
func GetTimeFromKey(key []byte) time.Time {
	t, err := unmarshalTime(key)
	if err != nil {
		panic(err)
	}
	return t
}

// GetIDFromKey returns the id based on key
func GetIDFromKey(key []byte) uint32 {
	id, err := unmarshal(key[len(key)-4:])
	if err != nil {
		panic(err)
	}
	return id
}

// GetPVKey returns the pageview key based on sessionkey and time
func GetPVKey(sessionkey []byte, t time.Time) []byte {
	return append(sessionkey, marshalTime(t)...)
}

// GetTimeFromPVKey return the time from the pageview key
func GetTimeFromPVKey(key []byte) time.Time {
	t, err := unmarshalTime(key[len(key)-8:])
	if err != nil {
		panic(err)
	}
	return t
}

// ShardUpdate runs the fn in a shard
func ShardUpdate(collectionID string, fn func(tx *shardbolt.MultiTx) error) error {
	sdb, err := getShardDB(collectionID)
	if err != nil {
		return err
	}
	return sdb.Update(fn)
}

// ShardUpsert upsert a value into shards
func ShardUpsert(collectionID string, key []byte, v proto.Message) error {
	sdb, err := getShardDB(collectionID)
	if err != nil {
		return err
	}
	bb := bucketName(v)
	vb, err := proto.Marshal(v)
	if err != nil {
		return err
	}
	return sdb.Update(func(tx *shardbolt.MultiTx) error {
		return tx.Put(bb, key, vb)
	})
}

// ShardUpsertBatch upsert a value into shards, but not in a separated transaction
func ShardUpsertBatch(collectionID string, key []byte, v proto.Message) error {
	sdb, err := getShardDB(collectionID)
	if err != nil {
		return err
	}
	bb := bucketName(v)
	vb, err := proto.Marshal(v)
	if err != nil {
		return err
	}
	return sdb.BatchUpsert(bb, key, vb)
}

// ShardUpsertTx upsert a value into shards in a transaction
func ShardUpsertTx(tx *shardbolt.MultiTx, key []byte, v proto.Message) error {
	bb := bucketName(v)
	vb, err := proto.Marshal(v)
	if err != nil {
		return err
	}
	return tx.Put(bb, key, vb)
}
