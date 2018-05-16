package db

import (
	"errors"
	"fmt"
	"time"

	proto "github.com/golang/protobuf/proto"
)

// These are the bucket's names
var (
	BUser       = []byte("User")
	BCollection = []byte("Collection")
	BSession    = []byte("Session")
	BPageview   = []byte("Pageview")
	BAuthToken  = []byte("AuthToken")
)

func bucketName(value interface{}) []byte {
	switch value := value.(type) {
	default:
		panic(fmt.Errorf("bucketName: invalid type: %T", value))
	case *User:
		return BUser
	case *Collection:
		return BCollection
	case *Session:
		return BSession
	case *Pageview:
		return BPageview
	case *AuthToken:
		return BAuthToken
	}
}

func protoEncode(value interface{}) ([]byte, error) {
	switch value := value.(type) {
	default:
		return nil, fmt.Errorf("protoEncode: invalid type: %T %v", value, value)
	case string:
		return []byte(value), nil
	case uint32:
		return marshal(value), nil
	case uint64:
		return marshaluint64(value), nil
	case proto.Message:
		return proto.Marshal(value)
	}
}

func protoDecode(data []byte, value interface{}) error {
	switch value := value.(type) {
	default:
		return fmt.Errorf("protoDecode: invalid type: %T", value)
	case *string:
		*value = string(data)
		return nil
	case *uint32:
		var err error
		*value, err = unmarshal(data)
		return err
	case *uint64:
		var err error
		*value, err = unmarshaluint64(data)
		return err
	case proto.Message:
		return proto.Unmarshal(data, value)
	}
}

func marshal(id uint32) []byte {
	return []byte{
		byte(id >> 24),
		byte(id >> 16),
		byte(id >> 8),
		byte(id),
	}
}

func marshaluint64(id uint64) []byte {
	return []byte{
		byte(id >> 56),
		byte(id >> 48),
		byte(id >> 40),
		byte(id >> 32),
		byte(id >> 24),
		byte(id >> 16),
		byte(id >> 8),
		byte(id),
	}
}

func unmarshal(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, errors.New("unmarshal uint32 invalid length")
	}
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24, nil
}

func unmarshaluint64(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, errors.New("unmarshal uint64 invalid length")
	}
	return uint64(b[7]) |
		uint64(b[6])<<8 |
		uint64(b[5])<<16 |
		uint64(b[4])<<24 |
		uint64(b[3])<<32 |
		uint64(b[2])<<40 |
		uint64(b[1])<<48 |
		uint64(b[0])<<56, nil
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
		return time.Time{}, errors.New(fmt.Sprint("unmarshalTime: invalid length ", len(data), " < 8"))
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
