package cipobolt

import (
	"errors"

	"github.com/boltdb/bolt"
)

type EncFun func(interface{}) ([]byte, error)
type DecFun func([]byte, interface{}) error
type BucFun func(interface{}) []byte

var (
	ErrKeyNotExists = errors.New("key not exists")
	ErrKeyExists    = errors.New("key exists")
)

type DB struct {
	bolt   *bolt.DB
	encode EncFun
	decode DecFun
	bucket BucFun
}

func Open(b *bolt.DB, ef EncFun, df DecFun, bf BucFun) *DB {
	return &DB{
		bolt:   b,
		encode: ef,
		decode: df,
		bucket: bf,
	}
}

func (db *DB) Bolt() *bolt.DB {
	return db.bolt
}

func (db *DB) Get(key interface{}, value interface{}) error {
	return db.bolt.View(func(tx *bolt.Tx) error {
		return db.GetTx(tx, key, value)
	})
}

func (db *DB) Iterate(key interface{}, value interface{}, fn func() error) error {
	return db.bolt.View(func(tx *bolt.Tx) error {
		return db.IterateTx(tx, key, value, fn)
	})
}

func (db *DB) Update(key interface{}, value interface{}) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return db.UpdateTx(tx, key, value)
	})
}

func (db *DB) Insert(key interface{}, value interface{}) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return db.InsertTx(tx, key, value)
	})
}

func (db *DB) Upsert(key interface{}, value interface{}) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return db.UpsertTx(tx, key, value)
	})
}

func (db *DB) Delete(key interface{}, value interface{}) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return db.DeleteTx(tx, key, value)
	})
}

func (db *DB) GetTx(tx *bolt.Tx, key interface{}, value interface{}) error {
	bb := db.bucket(value)
	b := tx.Bucket(bb)
	if b == nil {
		return ErrKeyNotExists
	}

	kb, err := db.encode(key)
	if err != nil {
		return err
	}

	data := b.Get(kb)
	if data == nil {
		return ErrKeyNotExists
	}

	return db.decode(data, value)
}

func (db *DB) IterateTx(tx *bolt.Tx, key interface{}, value interface{}, fn func() error) error {
	bb := db.bucket(value)
	b := tx.Bucket(bb)
	if b == nil {
		return nil
	}
	c := b.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if err := db.decode(k, key); err != nil {
			return err
		}
		if err := db.decode(v, value); err != nil {
			return err
		}
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) UpdateTx(tx *bolt.Tx, key interface{}, value interface{}) error {
	bb, kb, vb, err := db.getBytes(key, value)
	if err != nil {
		return err
	}

	b, _ := tx.CreateBucketIfNotExists(bb)

	v := b.Get(kb)
	if v == nil {
		return ErrKeyNotExists
	}

	return b.Put(kb, vb)
}

func (db *DB) InsertTx(tx *bolt.Tx, key interface{}, value interface{}) error {
	bb, kb, vb, err := db.getBytes(key, value)
	if err != nil {
		return err
	}

	b, _ := tx.CreateBucketIfNotExists(bb)

	v := b.Get(kb)
	if v != nil {
		return ErrKeyExists
	}

	return b.Put(kb, vb)
}

func (db *DB) UpsertTx(tx *bolt.Tx, key interface{}, value interface{}) error {
	bb, kb, vb, err := db.getBytes(key, value)
	if err != nil {
		return err
	}

	b, _ := tx.CreateBucketIfNotExists(bb)

	return b.Put(kb, vb)
}

func (db *DB) DeleteTx(tx *bolt.Tx, key interface{}, value interface{}) error {
	kb, err := db.encode(key)
	if err != nil {
		return err
	}

	b := tx.Bucket(db.bucket(value))
	if b == nil {
		return ErrKeyNotExists
	}

	v := b.Get(kb)
	if v == nil {
		return ErrKeyNotExists
	}

	return b.Delete(kb)
}

func (db *DB) getBytes(key interface{}, value interface{}) ([]byte, []byte, []byte, error) {
	bb := db.bucket(value)
	kb, err := db.encode(key)
	if err != nil {
		return nil, nil, nil, err
	}
	vb, err := db.encode(value)
	if err != nil {
		return nil, nil, nil, err
	}
	return bb, kb, vb, nil
}
