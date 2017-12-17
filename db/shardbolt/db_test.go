package shardbolt

import (
	"bytes"
	"errors"
	"log"
	"os"
	"testing"
	"time"
)

var (
	dir    = "test"
	bucket = []byte("test")
	key    = []byte("test")
	value  = []byte("test")
	now    = time.Now()
	mapFn  = func(key []byte) string {
		t, err := unmarshalTime(key)
		if err != nil {
			panic(err)
		}
		return t.Format("2006-01")
	}
)

func TestMain(m *testing.M) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalln(err)
	}

	ret := m.Run()

	os.Exit(ret)
}

func TestOpenNotExists(t *testing.T) {
	db, err := Open(dir, mapFn, 0666, nil)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
}

func TestPut(t *testing.T) {
	db, err := Open(dir, mapFn, 0666, nil)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	err = db.Update(func(tx *MultiTx) error {
		return tx.Put(bucket, createKey(now, key), value)
	})
	if err != nil {
		t.Error(err)
	}
	err = db.Update(func(tx *MultiTx) error {
		return tx.Put(bucket, createKey(now.Add(time.Duration(1)*time.Second), key), value)
	})
	if err != nil {
		t.Error(err)
	}
}

func TestIterate(t *testing.T) {
	db, err := Open(dir, mapFn, 0666, nil)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	count := 0
	db.Iterate(bucket, marshalTime(now), marshalTime(now.AddDate(0, 0, 1)), func(k []byte, v []byte) {
		count++
		if !bytes.HasSuffix(k, key) {
			t.Error("bad key", k, key)
		}
		if bytes.Compare(v, value) != 0 {
			t.Error("bad value", v, value)
		}
	})
	if count != 2 {
		t.Error(count)
	}

}

func TestGet(t *testing.T) {
	db, err := Open(dir, mapFn, 0666, nil)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	v, err := db.Get(bucket, createKey(now, key))
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(v, value) != 0 {
		t.Error("bad value", v, value)
	}

}

func marshalTime(t time.Time) []byte {
	nsec := t.UnixNano()
	enc := []byte{
		byte(nsec >> 56),
		byte(nsec >> 48),
		byte(nsec >> 40),
		byte(nsec >> 32),
		byte(nsec >> 24),
		byte(nsec >> 16),
		byte(nsec >> 8),
		byte(nsec),
	}
	return enc
}

func unmarshalTime(data []byte) (time.Time, error) {
	if len(data) < 8 {
		return time.Time{}, errors.New("unmarshalTime: invalid length")
	}
	data = data[:8]
	nsec := int64(data[0])<<56 |
		int64(data[1])<<48 |
		int64(data[2])<<40 |
		int64(data[3])<<32 |
		int64(data[4])<<24 |
		int64(data[5])<<16 |
		int64(data[6])<<8 |
		int64(data[7])
	return time.Unix(0, nsec), nil
}

func createKey(t time.Time, key []byte) []byte {
	return append(marshalTime(t), key...)
}
