package shardbolt

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/boltdb/bolt"
)

type shardArray []*shard

type DB struct {
	dir     string
	mapFn   func([]byte) string
	mode    os.FileMode
	options *Options
	shards  atomic.Value
}

type Options struct {
	FillPercent float64
	boltOptions *bolt.Options
}

func Open(dir string, mapFn func([]byte) string, mode os.FileMode, options *Options) (*DB, error) {
	if options == nil {
		options = &Options{
			FillPercent: 0.9,
		}
	}

	db := &DB{
		dir:     dir,
		mapFn:   mapFn,
		mode:    mode,
		options: options,
	}

	os.Mkdir(dir, os.ModePerm)

	shards := shardArray{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		shardID, err := getShardIDFromFilename(info.Name())
		if err != nil {
			log.Println(err)
			return nil
		}
		shard, err := db.openShard(shardID)
		if err != nil {
			log.Println("can't open shard:", path, "cause:", err)
			return nil
		}
		shards = append(shards, shard)
		return nil
	})
	if err != nil {
		log.Println("cant open dir:", dir, "cause:", err)
		return nil, err
	}
	sortShards(shards)
	db.setShardArray(shards)
	return db, nil
}

func (db *DB) Close() []error {
	shards := db.getShardArray()
	var errs []error
	for _, v := range shards {
		err := v.db.Close()
		if err != nil {
			errs = append(errs, err)
			log.Println(err)
		}
	}
	return errs
}

func (db *DB) DeleteShard(id string) error {
	shards := db.getShardArray()
	var newShards shardArray
	var ashard *shard
	for _, v := range shards {
		if v.id == id {
			ashard = v
		} else {
			newShards = append(newShards, v)
		}
	}
	if ashard == nil {
		return fmt.Errorf("shard not found '%s'", id)
	}
	db.setShardArray(newShards)
	if err := ashard.closeDB(); err != nil {
		return err
	}
	if err := os.Remove(db.getShardFileName(ashard)); err != nil {
		return err
	}
	return nil
}

func (db *DB) Iterate(bucket []byte, fromKey []byte, toKey []byte, fn func(k []byte, v []byte)) {
	shards := db.getShards(fromKey, toKey)
	for _, v := range shards {
		v.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucket)
			if b == nil {
				log.Println("bucket not found", string(bucket))
				return nil
			}
			c := b.Cursor()
			for k, v := c.Seek(fromKey); k != nil && bytes.Compare(k, toKey) < 0; k, v = c.Next() {
				fn(k, v)
			}
			return nil
		})
	}
}

func (db *DB) Get(bucket []byte, key []byte) ([]byte, error) {
	actualShard := db.getActualShard(key)
	if actualShard == nil {
		return nil, errors.New(fmt.Sprint("shard not found with key", key))
	}

	ret := []byte{}
	err := actualShard.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		ret = b.Get(key)
		if ret == nil {
			return errors.New(fmt.Sprint("key not found in shard", actualShard, key))
		}
		return nil
	})
	return ret, err
}

func (db *DB) Update(fn func(tx *MultiTx) error) error {
	tx := db.Begin(true)
	success := false

	defer func() {
		if !success {
			tx.Rollback()
		}
	}()

	err := fn(tx)

	if err != nil {
		tx.Rollback()
		return err
	}
	success = true
	return tx.Commit()
}

type ShardSize struct {
	Id   string
	Size int
}

func (db *DB) GetSizes() []ShardSize {
	shards := db.getShardArray()

	var sizes []ShardSize
	for _, v := range shards {
		size := -1
		fileinfo, err := os.Stat(v.db.Path())
		if err == nil {
			size = int(fileinfo.Size())
		}
		sizes = append(sizes, ShardSize{v.id, size})
	}
	return sizes
}

func (db *DB) getShardArray() shardArray {
	return db.shards.Load().(shardArray)
}

func (db *DB) setShardArray(shards shardArray) {
	db.shards.Store(shards)
}
