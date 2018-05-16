package cipobolt

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/boltdb/bolt"
)

// EncFun is an encoder function type for encoding the types
type EncFun func(interface{}) ([]byte, error)

// DecFun is a decoder function type for decoding the types
type DecFun func([]byte, interface{}) error

// BucFun is a function type for getting the bucket names
type BucFun func(interface{}) []byte

// The error types
var (
	ErrKeyNotExists = errors.New("key not exists")
	ErrKeyExists    = errors.New("key exists")
)

// DB is the struct for the DB
type DB struct {
	bolt   *bolt.DB
	encode EncFun
	decode DecFun
	bucket BucFun
}

// Open opens the database, sets the encoder/decoder functions
func Open(b *bolt.DB, ef EncFun, df DecFun, bf BucFun) *DB {
	return &DB{
		bolt:   b,
		encode: ef,
		decode: df,
		bucket: bf,
	}
}

// Bolt returns the underlying bolt instance
func (db *DB) Bolt() *bolt.DB {
	return db.bolt
}

// Get gets the value based on key
func (db *DB) Get(key interface{}, value interface{}) error {
	return db.bolt.View(func(tx *bolt.Tx) error {
		return db.GetTx(tx, key, value)
	})
}

// CountPrefix counts the values which belongs to a prefix key
func (db *DB) CountPrefix(prefix interface{}, value interface{}, count *int) error {
	return db.bolt.View(func(tx *bolt.Tx) error {
		return db.CountPrefixTx(tx, prefix, value, count)
	})
}

// Iterate iterates over the bucket and run fn on every element
// the key and the value stores the actual keys/values
func (db *DB) Iterate(key interface{}, value interface{}, fn func() error) error {
	return db.bolt.View(func(tx *bolt.Tx) error {
		return db.IterateTx(tx, key, value, fn)
	})
}

// IteratePrefix iterates the elements with prefix key over the bucket and run fn on every element
// the key and the value stores the actual keys/values
func (db *DB) IteratePrefix(key interface{}, value interface{}, prefix interface{}, fn func() error) error {
	return db.bolt.View(func(tx *bolt.Tx) error {
		return db.IteratePrefixTx(tx, key, value, prefix, fn)
	})
}

// Update updates an element
func (db *DB) Update(key interface{}, value interface{}) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return db.UpdateTx(tx, key, value)
	})
}

// Insert inserts an element
func (db *DB) Insert(key interface{}, value interface{}) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return db.InsertTx(tx, key, value)
	})
}

// Upsert upserts an element
func (db *DB) Upsert(key interface{}, value interface{}) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return db.UpsertTx(tx, key, value)
	})
}

// Delete deletes an element
func (db *DB) Delete(key interface{}, value interface{}) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return db.DeleteTx(tx, key, value)
	})
}

// DeletePrefix deletes elements with prefix
func (db *DB) DeletePrefix(prefix interface{}, value interface{}) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return db.DeleteTx(tx, prefix, value)
	})
}

// GetTx gets an element in a transaction
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

// CountPrefixTx counts elements with prefix in a transaction
func (db *DB) CountPrefixTx(tx *bolt.Tx, prefix interface{}, value interface{}, count *int) error {
	*count = 0
	bb := db.bucket(value)
	b := tx.Bucket(bb)
	if b == nil {
		return nil
	}

	pb, err := db.encode(prefix)
	if err != nil {
		return err
	}

	c := b.Cursor()
	for k, _ := c.Seek(pb); k != nil && bytes.HasPrefix(k, pb); k, _ = c.Next() {
		*count++
	}
	return nil
}

// IterateTx iterates over the bucket in a transaction
// calls fn in every element, stores the actual key/values in key/value
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

// IteratePrefixTx iterates elements with prefix over the bucket in a transaction
// calls fn in every element, stores the actual key/values in key/value
func (db *DB) IteratePrefixTx(tx *bolt.Tx, key interface{}, value interface{}, prefix interface{}, fn func() error) error {
	bb := db.bucket(value)
	b := tx.Bucket(bb)
	if b == nil {
		return nil
	}
	c := b.Cursor()

	bPrefix, err := db.encode(prefix)
	if err != nil {
		return err
	}
	for k, v := c.Seek(bPrefix); k != nil && bytes.HasPrefix(k, bPrefix); k, v = c.Next() {
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

// UpdateTx updates an element in a transaction
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

// InsertTx inserts an element in a transaction
func (db *DB) InsertTx(tx *bolt.Tx, key interface{}, value interface{}) error {
	var err error
	kb := []byte{}
	bb := db.bucket(value)

	b, _ := tx.CreateBucketIfNotExists(bb)

	if key == nil {
		ukey, _ := b.NextSequence()
		kb, err = db.encode(ukey)
		if err != nil {
			return err
		}

		dv := reflect.Indirect(reflect.ValueOf(value))
		dt := dv.Type()
		if dt.NumField() > 0 && dt.Field(0).Name == "ID" && dt.Field(0).Type.Name() == "uint64" {
			dv.Field(0).SetUint(ukey)
		}
	} else {
		kb, err = db.encode(key)
		if err != nil {
			return err
		}
		v := b.Get(kb)
		if v != nil {
			return ErrKeyExists
		}
	}
	vb, err := db.encode(value)
	if err != nil {
		return err
	}
	return b.Put(kb, vb)
}

// UpsertTx upserts an element in a transaction
func (db *DB) UpsertTx(tx *bolt.Tx, key interface{}, value interface{}) error {
	bb, kb, vb, err := db.getBytes(key, value)
	if err != nil {
		return err
	}

	b, _ := tx.CreateBucketIfNotExists(bb)

	return b.Put(kb, vb)
}

// DeleteTx deletes an element in a transaction
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

// DeletePrefixTx deletes elements with prefix in a transaction
func (db *DB) DeletePrefixTx(tx *bolt.Tx, prefix interface{}, value interface{}) error {
	pb, err := db.encode(prefix)
	if err != nil {
		return err
	}

	b := tx.Bucket(db.bucket(value))
	if b == nil {
		return ErrKeyNotExists
	}

	c := b.Cursor()
	for k, _ := c.Seek(pb); k != nil && bytes.HasPrefix(k, pb); k, _ = c.Next() {
		if err := b.Delete(k); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) getBytes(key interface{}, value interface{}) (bb []byte, kb []byte, vb []byte, err error) {
	bb = db.bucket(value)
	kb, err = db.encode(key)
	if err != nil {
		return
	}
	vb, err = db.encode(value)
	if err != nil {
		return
	}
	return
}
