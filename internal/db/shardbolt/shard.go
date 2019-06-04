package shardbolt

import (
	"fmt"
	"sort"
	"strings"

	bolt "github.com/etcd-io/bbolt"
)

type shard struct {
	id string
	db *bolt.DB
}

func (db *DB) openShard(shardID string) (*shard, error) {
	actualDB, err := bolt.Open(db.dir+"/"+shardID+".bolt", db.mode, db.options.boltOptions)
	if err != nil {
		return nil, err
	}
	return &shard{shardID, actualDB}, nil
}

func (s *shard) closeDB() error {
	return s.db.Close()
}

func (db *DB) getShardFileName(s *shard) string {
	return db.dir + "/" + s.id + ".bolt"
}

func getShardIDFromFilename(fname string) (string, error) {
	idx := strings.Index(fname, ".bolt")
	if idx == -1 {
		return "", fmt.Errorf("invalid shard filename: %v", fname)
	}
	shardID := fname[:idx]
	return shardID, nil
}

func (db *DB) getActualShard(key []byte) *shard {
	shards := db.getShardArray()
	id := db.mapFn(key)
	for _, v := range shards {
		if v.id == id {
			return v
		}
	}
	return nil
}

func (db *DB) getShards(fromKey []byte, toKey []byte) []*shard {
	fromID := db.mapFn(fromKey)
	toID := db.mapFn(toKey)
	out := []*shard{}

	shards := db.getShardArray()

	for _, v := range shards {
		if fromID <= v.id && v.id <= toID {
			out = append(out, v)
		}
	}
	return out
}

func (db *DB) createActualShard(key []byte) (*shard, error) {
	shardID := db.mapFn(key)
	newShard, err := db.openShard(shardID)
	if err != nil {
		return nil, err
	}
	shards := db.getShardArray()

	newShards := make(shardArray, len(shards), len(shards)+1)
	copy(newShards, shards)
	newShards = append(newShards, newShard)
	sortShards(newShards)
	db.setShardArray(newShards)
	return newShard, nil
}

func (db *DB) ensureShard(key []byte) (*shard, error) {
	actualShard := db.getActualShard(key)
	if actualShard == nil {
		db.shardMutex.Lock()
		defer db.shardMutex.Unlock()
		actualShard = db.getActualShard(key)
		if actualShard != nil {
			return actualShard, nil
		}
		var err error
		actualShard, err = db.createActualShard(key)
		if err != nil {
			return nil, err
		}
	}
	return actualShard, nil
}

func sortShards(shards shardArray) {
	sort.Slice(shards, func(i, j int) bool {
		return shards[i].id < shards[j].id
	})
}
