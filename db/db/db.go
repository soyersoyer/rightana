package db

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync/atomic"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"

	"git.irl.hu/k20a/db/cipobolt"
	"git.irl.hu/k20a/db/shardbolt"
)

type shardMap map[string]*shardbolt.DB

var (
	basedir  = "data/"
	filename = "k20a.bolt"

	cipo     *cipobolt.DB
	shardDBs = atomic.Value{}

	KeyExists    = cipobolt.KeyExists
	KeyNotExists = cipobolt.KeyNotExists
)

func map2Month(key []byte) string {
	t, err := unmarshalTime(key)
	if err != nil {
		panic(err)
	}
	return t.Format("2006-01")
}

func InitDatabase(basedir_ string) {
	basedir = path.Clean(basedir_) + "/"
	os.MkdirAll(basedir, os.ModePerm)
	bdb, err := bolt.Open(basedir+filename, 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	cipo = cipobolt.Open(bdb, protoEncode, protoDecode, bucketName)
	shardDBs.Store(shardMap{})
}

func UpdateUser(user *User) error {
	return cipo.Update(user.Email, user)
}

func InsertUser(user *User) error {
	return cipo.Insert(user.Email, user)
}

func UpsertUser(user *User) error {
	return cipo.Upsert(user.Email, user)
}

func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := cipo.Get(email, user)
	return user, err
}

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
		if err := deleteCollababorationsByUserEmailTx(tx, user.Email); err != nil {
			return err
		}
		return nil
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

func deleteCollababorationsByUserEmailTx(tx *bolt.Tx, email string) error {
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

func InsertAuthToken(token *AuthToken) error {
	token.TTL = 1209600
	token.Created = time.Now().UnixNano()
	return cipo.Insert(token.ID, token)
}

func GetAuthToken(id string) (*AuthToken, error) {
	token := &AuthToken{}
	err := cipo.Get(id, token)
	return token, err
}

func UpdateAuthToken(token *AuthToken) error {
	return cipo.Update(token.ID, token)
}

func DeleteAuthToken(id string) error {
	return cipo.Delete(id, &AuthToken{})
}

func InsertCollection(collection *Collection) error {
	return cipo.Insert(collection.ID, collection)
}

func UpdateCollection(collection *Collection) error {
	return cipo.Update(collection.ID, collection)
}

func updateCollectionTx(tx *bolt.Tx, collection *Collection) error {
	return cipo.UpdateTx(tx, collection.ID, collection)
}

func GetCollection(id string) (*Collection, error) {
	collection := Collection{}
	err := cipo.Get(id, &collection)
	return &collection, err
}

func DeleteCollection(collection *Collection) error {
	return cipo.Bolt().Update(func(tx *bolt.Tx) error {
		return deleteCollectionTx(tx, collection)
	})
}

func deleteCollectionTx(tx *bolt.Tx, collection *Collection) error {
	if err := cipo.DeleteTx(tx, collection.ID, collection); err != nil {
		return err
	}
	if err := deleteShardDB(collection.ID); err != nil {
		return err
	}
	return nil
}

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

func UserIsTeammate(collection *Collection, email string) bool {
	for _, v := range collection.Teammates {
		if v.Email == email {
			return true
		}
	}
	return false
}

type Shard struct {
	Id   string `json:"id"`
	Size int    `json:"size"`
}

func GetCollectionShards(collection *Collection) ([]Shard, error) {
	db, err := getShardDB(collection.ID)
	if err != nil {
		return nil, err
	}
	ret := []Shard{}
	for _, v := range db.GetSizes() {
		ret = append(ret, Shard(v))
	}
	return ret, nil
}

func DeleteCollectionShard(collection *Collection, shardID string) error {
	db, err := getShardDB(collection.ID)
	if err != nil {
		return err
	}
	return db.DeleteShard(shardID)
}

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

func GetKey(t time.Time, id uint32) []byte {
	return append(marshalTime(t), marshal(id)...)
}

func GetKeyFromTime(t time.Time) []byte {
	return marshalTime(t)
}

func GetTimeFromKey(key []byte) time.Time {
	t, err := unmarshalTime(key)
	if err != nil {
		panic(err)
	}
	return t
}

func GetIdFromKey(key []byte) uint32 {
	id, err := unmarshal(key[len(key)-4:])
	if err != nil {
		panic(err)
	}
	return id
}

func ShardUpdate(collectionID string, fn func(tx *shardbolt.MultiTx) error) error {
	sdb, err := getShardDB(collectionID)
	if err != nil {
		return err
	}
	return sdb.Update(fn)
}

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

func ShardUpsertTx(tx *shardbolt.MultiTx, key []byte, v proto.Message) error {
	bb := bucketName(v)
	vb, err := proto.Marshal(v)
	if err != nil {
		return err
	}
	return tx.Put(bb, key, vb)
}
