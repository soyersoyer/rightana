package shardbolt

import (
	"log"

	"github.com/boltdb/bolt"
)

type MultiTx struct {
	db        *DB
	writeable bool
	txs       []*shardTx
}

type shardTx struct {
	id string
	tx *bolt.Tx
}

func (db *DB) Begin(writeable bool) *MultiTx {
	return &MultiTx{db, writeable, nil}
}

func (tx *MultiTx) Rollback() error {
	var err error
	for _, v := range tx.txs {
		err_ := v.tx.Rollback()
		if err_ != nil {
			log.Println(err_)
			err = err_
		}
	}
	return err
}

func (tx *MultiTx) Commit() error {
	var err error
	for _, v := range tx.txs {
		err_ := v.tx.Commit()
		if err_ != nil {
			log.Println(err_)
			err = err_
		}
	}
	return err
}

func (tx *MultiTx) Put(bucket []byte, key []byte, value []byte) error {
	stx, err := tx.ensureTx(key)
	if err != nil {
		return err
	}

	b, err := stx.tx.CreateBucketIfNotExists(bucket)
	if err != nil {
		return err
	}
	b.FillPercent = tx.db.options.FillPercent
	return b.Put(key, value)
}

func (tx *MultiTx) ensureTx(key []byte) (*shardTx, error) {
	id := tx.db.mapFn(key)
	for _, v := range tx.txs {
		if v.id == id {
			return v, nil
		}
	}

	actualShard, err := tx.db.ensureShard(key)
	if err != nil {
		return nil, err
	}

	btx, err := actualShard.db.Begin(tx.writeable)
	if err != nil {
		return nil, err
	}
	stx := &shardTx{id, btx}
	tx.txs = append(tx.txs, stx)
	return stx, nil
}
