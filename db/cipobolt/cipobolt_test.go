package cipobolt

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

type Data struct {
	value string
}

var (
	dir  = "data"
	cipo *DB

	encode = func(v interface{}) ([]byte, error) {
		switch v := v.(type) {
		default:
			return nil, fmt.Errorf("encode: bad type %+v", v)
		case string:
			return []byte(v), nil
		case *Data:
			return []byte(v.value), nil
		}
	}
	decode = func(data []byte, v interface{}) error {
		switch v := v.(type) {
		default:
			return fmt.Errorf("decode: bad type %+v", v)
		case *string:
			*v = string(data)
			return nil
		case *Data:
			v.value = string(data)
			return nil
		}
	}
	bucket = func(v interface{}) []byte {
		switch v := v.(type) {
		default:
			log.Panicf("bucket: bad type %+v", v)
			return nil
		case *Data:
			return []byte("data")
		}
	}
	key1 = "hello1"
	key2 = "hello2"
	key3 = "prehello"
	val1 = &Data{"world1"}
	val2 = &Data{"world2"}
	val3 = &Data{"prehello"}
)

func TestMain(m *testing.M) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Mkdir(dir, 0700)
	if err != nil {
		log.Fatalln(err)
	}

	bdb, err := bolt.Open(dir+"/bolt.db", 0700, nil)
	if err != nil {
		log.Fatalln(err)
	}

	cipo = Open(bdb, encode, decode, bucket)

	ret := m.Run()

	if ret == 0 {
		os.RemoveAll(dir)
	}
	os.Exit(ret)
}

func TestGet(t *testing.T) {
	err := cipo.Get(key1, val1)
	if err == nil {
		t.Error("non existent get should fail")
	}
}

func TestInsert(t *testing.T) {
	err := cipo.Insert(key1, val1)
	if err != nil {
		t.Error(err)
	}

	checkValue(t, key1, val1)

	err = cipo.Insert(key1, val1)
	if err == nil {
		t.Error("second insert should fail")
	}
}

func TestUpdate(t *testing.T) {
	err := cipo.Update(key1, val2)
	if err != nil {
		t.Error(err)
	}
	checkValue(t, key1, val2)

	err = cipo.Update(key2, val2)
	if err == nil {
		t.Error("non existent update should fail")
	}
}

func TestUpsert(t *testing.T) {
	err := cipo.Upsert(key1, val1)
	if err != nil {
		t.Error(err)
	}
	checkValue(t, key1, val1)

	err = cipo.Upsert(key2, val2)
	if err != nil {
		t.Error(err)
	}
	checkValue(t, key2, val2)
}

func TestIterate(t *testing.T) {
	k := ""
	v := &Data{}
	count := 0
	err := cipo.Iterate(&k, v, func() error {
		count++
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if count != 2 {
		t.Error("there should be two value")
	}
}

func TestDelete(t *testing.T) {
	err := cipo.Delete(key1, val1)
	if err != nil {
		t.Error(err)
	}
	err = cipo.Get(key1, val1)
	if err == nil {
		t.Error("non existent get should fail")
	}
}

func TestCountPrefix(t *testing.T) {
	if err := cipo.Insert(key3, val3); err != nil {
		t.Error(err)
	}

	checkValue(t, key3, val3)

	count := 0
	if err := cipo.CountPrefix("pre", &Data{}, &count); err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("there should be one value")
	}
}

func TestIteratePrefix(t *testing.T) {
	k := ""
	v := &Data{}
	count := 0
	if err := cipo.IteratePrefix(&k, v, "pre", func() error {
		count++
		return nil
	}); err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("there should be one value")
	}
}

func DeletePrefix(t *testing.T) {
	if err := cipo.DeletePrefix("pre", &Data{}); err != nil {
		t.Error(err)
	}
	if err := cipo.Get(key3, val3); err != nil {
		t.Error("non existent get should fail")
	}
}

func checkValue(t *testing.T, key string, value *Data) {
	v := &Data{}
	err := cipo.Get(key, v)
	if err != nil {
		t.Error(err)
	}
	if v.value != value.value {
		t.Error(v, "!=", value)
	}
}
