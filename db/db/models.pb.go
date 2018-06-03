// Code generated by protoc-gen-go. DO NOT EDIT.
// source: models.proto

/*
Package db is a generated protocol buffer package.

It is generated from these files:
	models.proto

It has these top-level messages:
	User
	Teammate
	Collection
	AuthToken
	Session
	Pageview
*/
package db

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type User struct {
	ID                  uint64 `protobuf:"varint,1,opt,name=ID,json=iD" json:"ID,omitempty"`
	Email               string `protobuf:"bytes,2,opt,name=Email,json=email" json:"Email,omitempty"`
	Password            string `protobuf:"bytes,3,opt,name=Password,json=password" json:"Password,omitempty"`
	Created             int64  `protobuf:"varint,4,opt,name=Created,json=created" json:"Created,omitempty"`
	Name                string `protobuf:"bytes,5,opt,name=Name,json=name" json:"Name,omitempty"`
	IsAdmin             bool   `protobuf:"varint,10,opt,name=IsAdmin,json=isAdmin" json:"IsAdmin,omitempty"`
	DisablePwChange     bool   `protobuf:"varint,11,opt,name=DisablePwChange,json=disablePwChange" json:"DisablePwChange,omitempty"`
	LimitCollections    bool   `protobuf:"varint,12,opt,name=LimitCollections,json=limitCollections" json:"LimitCollections,omitempty"`
	CollectionLimit     uint32 `protobuf:"varint,13,opt,name=CollectionLimit,json=collectionLimit" json:"CollectionLimit,omitempty"`
	DisableUserDeletion bool   `protobuf:"varint,14,opt,name=DisableUserDeletion,json=disableUserDeletion" json:"DisableUserDeletion,omitempty"`
}

func (m *User) Reset()                    { *m = User{} }
func (m *User) String() string            { return proto.CompactTextString(m) }
func (*User) ProtoMessage()               {}
func (*User) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *User) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *User) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *User) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *User) GetCreated() int64 {
	if m != nil {
		return m.Created
	}
	return 0
}

func (m *User) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *User) GetIsAdmin() bool {
	if m != nil {
		return m.IsAdmin
	}
	return false
}

func (m *User) GetDisablePwChange() bool {
	if m != nil {
		return m.DisablePwChange
	}
	return false
}

func (m *User) GetLimitCollections() bool {
	if m != nil {
		return m.LimitCollections
	}
	return false
}

func (m *User) GetCollectionLimit() uint32 {
	if m != nil {
		return m.CollectionLimit
	}
	return 0
}

func (m *User) GetDisableUserDeletion() bool {
	if m != nil {
		return m.DisableUserDeletion
	}
	return false
}

type Teammate struct {
	ID uint64 `protobuf:"varint,1,opt,name=ID,json=iD" json:"ID,omitempty"`
}

func (m *Teammate) Reset()                    { *m = Teammate{} }
func (m *Teammate) String() string            { return proto.CompactTextString(m) }
func (*Teammate) ProtoMessage()               {}
func (*Teammate) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Teammate) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

type Collection struct {
	ID        string      `protobuf:"bytes,1,opt,name=ID,json=iD" json:"ID,omitempty"`
	OwnerID   uint64      `protobuf:"varint,2,opt,name=OwnerID,json=ownerID" json:"OwnerID,omitempty"`
	Name      string      `protobuf:"bytes,3,opt,name=Name,json=name" json:"Name,omitempty"`
	Teammates []*Teammate `protobuf:"bytes,4,rep,name=Teammates,json=teammates" json:"Teammates,omitempty"`
	Created   int64       `protobuf:"varint,5,opt,name=Created,json=created" json:"Created,omitempty"`
}

func (m *Collection) Reset()                    { *m = Collection{} }
func (m *Collection) String() string            { return proto.CompactTextString(m) }
func (*Collection) ProtoMessage()               {}
func (*Collection) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Collection) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Collection) GetOwnerID() uint64 {
	if m != nil {
		return m.OwnerID
	}
	return 0
}

func (m *Collection) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Collection) GetTeammates() []*Teammate {
	if m != nil {
		return m.Teammates
	}
	return nil
}

func (m *Collection) GetCreated() int64 {
	if m != nil {
		return m.Created
	}
	return 0
}

type AuthToken struct {
	ID      string `protobuf:"bytes,1,opt,name=ID,json=iD" json:"ID,omitempty"`
	OwnerID uint64 `protobuf:"varint,2,opt,name=OwnerID,json=ownerID" json:"OwnerID,omitempty"`
	TTL     int32  `protobuf:"varint,3,opt,name=TTL,json=tTL" json:"TTL,omitempty"`
	Created int64  `protobuf:"varint,4,opt,name=Created,json=created" json:"Created,omitempty"`
}

func (m *AuthToken) Reset()                    { *m = AuthToken{} }
func (m *AuthToken) String() string            { return proto.CompactTextString(m) }
func (*AuthToken) ProtoMessage()               {}
func (*AuthToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *AuthToken) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *AuthToken) GetOwnerID() uint64 {
	if m != nil {
		return m.OwnerID
	}
	return 0
}

func (m *AuthToken) GetTTL() int32 {
	if m != nil {
		return m.TTL
	}
	return 0
}

func (m *AuthToken) GetCreated() int64 {
	if m != nil {
		return m.Created
	}
	return 0
}

type Session struct {
	Duration         int32  `protobuf:"varint,1,opt,name=Duration,json=duration" json:"Duration,omitempty"`
	Hostname         string `protobuf:"bytes,2,opt,name=Hostname,json=hostname" json:"Hostname,omitempty"`
	DeviceOS         string `protobuf:"bytes,3,opt,name=DeviceOS,json=deviceOS" json:"DeviceOS,omitempty"`
	BrowserName      string `protobuf:"bytes,4,opt,name=BrowserName,json=browserName" json:"BrowserName,omitempty"`
	BrowserVersion   string `protobuf:"bytes,5,opt,name=BrowserVersion,json=browserVersion" json:"BrowserVersion,omitempty"`
	BrowserLanguage  string `protobuf:"bytes,6,opt,name=BrowserLanguage,json=browserLanguage" json:"BrowserLanguage,omitempty"`
	ScreenResolution string `protobuf:"bytes,7,opt,name=ScreenResolution,json=screenResolution" json:"ScreenResolution,omitempty"`
	WindowResolution string `protobuf:"bytes,8,opt,name=WindowResolution,json=windowResolution" json:"WindowResolution,omitempty"`
	DeviceType       string `protobuf:"bytes,9,opt,name=DeviceType,json=deviceType" json:"DeviceType,omitempty"`
	CountryCode      string `protobuf:"bytes,10,opt,name=CountryCode,json=countryCode" json:"CountryCode,omitempty"`
	City             string `protobuf:"bytes,11,opt,name=City,json=city" json:"City,omitempty"`
	UserAgent        string `protobuf:"bytes,12,opt,name=UserAgent,json=userAgent" json:"UserAgent,omitempty"`
	UserIP           string `protobuf:"bytes,13,opt,name=UserIP,json=userIP" json:"UserIP,omitempty"`
	UserHostname     string `protobuf:"bytes,14,opt,name=UserHostname,json=userHostname" json:"UserHostname,omitempty"`
	Referrer         string `protobuf:"bytes,15,opt,name=Referrer,json=referrer" json:"Referrer,omitempty"`
	ASNumber         int32  `protobuf:"varint,16,opt,name=ASNumber,json=aSNumber" json:"ASNumber,omitempty"`
	ASName           string `protobuf:"bytes,17,opt,name=ASName,json=aSName" json:"ASName,omitempty"`
}

func (m *Session) Reset()                    { *m = Session{} }
func (m *Session) String() string            { return proto.CompactTextString(m) }
func (*Session) ProtoMessage()               {}
func (*Session) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Session) GetDuration() int32 {
	if m != nil {
		return m.Duration
	}
	return 0
}

func (m *Session) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

func (m *Session) GetDeviceOS() string {
	if m != nil {
		return m.DeviceOS
	}
	return ""
}

func (m *Session) GetBrowserName() string {
	if m != nil {
		return m.BrowserName
	}
	return ""
}

func (m *Session) GetBrowserVersion() string {
	if m != nil {
		return m.BrowserVersion
	}
	return ""
}

func (m *Session) GetBrowserLanguage() string {
	if m != nil {
		return m.BrowserLanguage
	}
	return ""
}

func (m *Session) GetScreenResolution() string {
	if m != nil {
		return m.ScreenResolution
	}
	return ""
}

func (m *Session) GetWindowResolution() string {
	if m != nil {
		return m.WindowResolution
	}
	return ""
}

func (m *Session) GetDeviceType() string {
	if m != nil {
		return m.DeviceType
	}
	return ""
}

func (m *Session) GetCountryCode() string {
	if m != nil {
		return m.CountryCode
	}
	return ""
}

func (m *Session) GetCity() string {
	if m != nil {
		return m.City
	}
	return ""
}

func (m *Session) GetUserAgent() string {
	if m != nil {
		return m.UserAgent
	}
	return ""
}

func (m *Session) GetUserIP() string {
	if m != nil {
		return m.UserIP
	}
	return ""
}

func (m *Session) GetUserHostname() string {
	if m != nil {
		return m.UserHostname
	}
	return ""
}

func (m *Session) GetReferrer() string {
	if m != nil {
		return m.Referrer
	}
	return ""
}

func (m *Session) GetASNumber() int32 {
	if m != nil {
		return m.ASNumber
	}
	return 0
}

func (m *Session) GetASName() string {
	if m != nil {
		return m.ASName
	}
	return ""
}

type Pageview struct {
	Path        string `protobuf:"bytes,1,opt,name=Path,json=path" json:"Path,omitempty"`
	QueryString string `protobuf:"bytes,2,opt,name=QueryString,json=queryString" json:"QueryString,omitempty"`
}

func (m *Pageview) Reset()                    { *m = Pageview{} }
func (m *Pageview) String() string            { return proto.CompactTextString(m) }
func (*Pageview) ProtoMessage()               {}
func (*Pageview) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Pageview) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Pageview) GetQueryString() string {
	if m != nil {
		return m.QueryString
	}
	return ""
}

func init() {
	proto.RegisterType((*User)(nil), "db.User")
	proto.RegisterType((*Teammate)(nil), "db.Teammate")
	proto.RegisterType((*Collection)(nil), "db.Collection")
	proto.RegisterType((*AuthToken)(nil), "db.AuthToken")
	proto.RegisterType((*Session)(nil), "db.Session")
	proto.RegisterType((*Pageview)(nil), "db.Pageview")
}

func init() { proto.RegisterFile("models.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 641 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xdb, 0x8e, 0xda, 0x3c,
	0x10, 0x56, 0x20, 0x40, 0x32, 0xb0, 0xc0, 0xef, 0xfd, 0x55, 0x59, 0xab, 0xaa, 0x42, 0x5c, 0x54,
	0x68, 0x2f, 0x56, 0x55, 0xfb, 0x02, 0xa5, 0x50, 0xa9, 0x48, 0x68, 0x97, 0x1a, 0xda, 0x5e, 0x3b,
	0x64, 0x0a, 0x56, 0x93, 0x98, 0xda, 0xce, 0x46, 0xbc, 0x45, 0x5f, 0xa3, 0xaf, 0xd1, 0x27, 0xab,
	0xec, 0x84, 0xe5, 0x50, 0xf5, 0xa2, 0x97, 0xdf, 0xc1, 0x19, 0xcf, 0xcc, 0xe7, 0x40, 0x27, 0x95,
	0x31, 0x26, 0xfa, 0x6e, 0xa7, 0xa4, 0x91, 0xa4, 0x16, 0x47, 0xc3, 0x5f, 0x35, 0xf0, 0x3f, 0x69,
	0x54, 0xa4, 0x0b, 0xb5, 0xd9, 0x94, 0x7a, 0x03, 0x6f, 0xe4, 0xb3, 0x9a, 0x98, 0x92, 0xff, 0xa1,
	0xf1, 0x3e, 0xe5, 0x22, 0xa1, 0xb5, 0x81, 0x37, 0x0a, 0x59, 0x03, 0x2d, 0x20, 0x37, 0x10, 0x2c,
	0xb8, 0xd6, 0x85, 0x54, 0x31, 0xad, 0x3b, 0x21, 0xd8, 0x55, 0x98, 0x50, 0x68, 0x4d, 0x14, 0x72,
	0x83, 0x31, 0xf5, 0x07, 0xde, 0xa8, 0xce, 0x5a, 0xeb, 0x12, 0x12, 0x02, 0xfe, 0x3d, 0x4f, 0x91,
	0x36, 0xdc, 0x09, 0x3f, 0xe3, 0x29, 0x5a, 0xf7, 0x4c, 0x8f, 0xe3, 0x54, 0x64, 0x14, 0x06, 0xde,
	0x28, 0x60, 0x2d, 0x51, 0x42, 0x32, 0x82, 0xde, 0x54, 0x68, 0x1e, 0x25, 0xb8, 0x28, 0x26, 0x5b,
	0x9e, 0x6d, 0x90, 0xb6, 0x9d, 0xa3, 0x17, 0x9f, 0xd3, 0xe4, 0x16, 0xfa, 0x73, 0x91, 0x0a, 0x33,
	0x91, 0x49, 0x82, 0x6b, 0x23, 0x64, 0xa6, 0x69, 0xc7, 0x59, 0xfb, 0xc9, 0x05, 0x6f, 0xbf, 0x7a,
	0x84, 0xee, 0x14, 0xbd, 0x1a, 0x78, 0xa3, 0x2b, 0xd6, 0x5b, 0x9f, 0xd3, 0xe4, 0x15, 0x5c, 0x57,
	0xf5, 0xed, 0x60, 0xa6, 0x98, 0xa0, 0xd5, 0x68, 0xd7, 0x7d, 0xf8, 0x3a, 0xfe, 0x53, 0x1a, 0xde,
	0x40, 0xb0, 0x42, 0x9e, 0xa6, 0xdc, 0xe0, 0xe5, 0x1c, 0x87, 0x3f, 0x3c, 0x80, 0x63, 0xe1, 0x13,
	0x39, 0x74, 0x63, 0xa6, 0xd0, 0x7a, 0x28, 0x32, 0x54, 0xb3, 0xa9, 0x1b, 0xb4, 0xcf, 0x5a, 0xb2,
	0x84, 0x4f, 0x43, 0xab, 0x9f, 0x0c, 0xed, 0x16, 0xc2, 0x43, 0x21, 0x4d, 0xfd, 0x41, 0x7d, 0xd4,
	0x7e, 0xdd, 0xb9, 0x8b, 0xa3, 0xbb, 0x03, 0xc9, 0x42, 0x73, 0x90, 0x4f, 0xd7, 0xd1, 0x38, 0x5b,
	0xc7, 0x90, 0x43, 0x38, 0xce, 0xcd, 0x76, 0x25, 0xbf, 0xe1, 0xbf, 0x5c, 0xa8, 0x0f, 0xf5, 0xd5,
	0x6a, 0xee, 0xee, 0xd3, 0x60, 0x75, 0xb3, 0x9a, 0xff, 0x7d, 0xe3, 0xc3, 0x9f, 0x3e, 0xb4, 0x96,
	0xa8, 0xb5, 0x6d, 0xf9, 0x06, 0x82, 0x69, 0xae, 0xb8, 0x1b, 0xa2, 0xe7, 0x0e, 0x07, 0x71, 0x85,
	0xad, 0xf6, 0x41, 0x6a, 0x63, 0x9b, 0xab, 0x82, 0x16, 0x6c, 0x2b, 0xec, 0xce, 0xe1, 0xa3, 0x58,
	0xe3, 0xc3, 0xf2, 0x90, 0xb5, 0xb8, 0xc2, 0x64, 0x00, 0xed, 0x77, 0x4a, 0x16, 0x1a, 0x95, 0x9b,
	0x91, 0xef, 0xe4, 0x76, 0x74, 0xa4, 0xc8, 0x4b, 0xe8, 0x56, 0x8e, 0xcf, 0xa8, 0xec, 0x3d, 0xaa,
	0xf4, 0x75, 0xa3, 0x33, 0xd6, 0xe6, 0xa2, 0xf2, 0xcd, 0x79, 0xb6, 0xc9, 0xf9, 0x06, 0x69, 0xd3,
	0x19, 0x7b, 0xd1, 0x39, 0x6d, 0xd3, 0xb6, 0x5c, 0x2b, 0xc4, 0x8c, 0xa1, 0x96, 0x49, 0xee, 0xfa,
	0x69, 0x39, 0x6b, 0x5f, 0x5f, 0xf0, 0xd6, 0xfb, 0x45, 0x64, 0xb1, 0x2c, 0x4e, 0xbc, 0x41, 0xe9,
	0x2d, 0x2e, 0x78, 0xf2, 0x02, 0xa0, 0xec, 0x73, 0xb5, 0xdf, 0x21, 0x0d, 0x9d, 0x0b, 0xe2, 0x27,
	0xc6, 0xf6, 0x3a, 0x91, 0x79, 0x66, 0xd4, 0x7e, 0x22, 0x63, 0x74, 0xaf, 0x25, 0x64, 0xed, 0xf5,
	0x91, 0xb2, 0x51, 0x99, 0x08, 0xb3, 0x77, 0xcf, 0x24, 0x64, 0xfe, 0x5a, 0x98, 0x3d, 0x79, 0x0e,
	0xa1, 0xcd, 0xe8, 0x78, 0x83, 0x99, 0x71, 0x8f, 0x22, 0x64, 0x61, 0x7e, 0x20, 0xc8, 0x33, 0x68,
	0x5a, 0x75, 0xb6, 0x70, 0x8f, 0x20, 0x64, 0xcd, 0xdc, 0x21, 0x32, 0x84, 0x8e, 0xe5, 0x9f, 0x76,
	0xd2, 0x75, 0x6a, 0x27, 0x3f, 0xe1, 0xec, 0x5e, 0x18, 0x7e, 0x45, 0xa5, 0x50, 0xd1, 0x5e, 0xb9,
	0x17, 0x55, 0x61, 0xab, 0x8d, 0x97, 0xf7, 0x79, 0x1a, 0xa1, 0xa2, 0xfd, 0x72, 0xd7, 0xbc, 0xc2,
	0xb6, 0xe6, 0x78, 0xe9, 0xd6, 0xf5, 0x5f, 0x59, 0x93, 0x3b, 0x34, 0x7c, 0x6b, 0xff, 0x29, 0x1b,
	0x7c, 0x14, 0x58, 0xd8, 0x4e, 0x16, 0xdc, 0x6c, 0xab, 0x3c, 0xfa, 0x3b, 0x6e, 0xb6, 0xb6, 0xff,
	0x8f, 0x39, 0xaa, 0xfd, 0xd2, 0x28, 0x91, 0x6d, 0xaa, 0x98, 0xb4, 0xbf, 0x1f, 0xa9, 0xa8, 0xe9,
	0xfe, 0x67, 0x6f, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0xb7, 0xa7, 0xb6, 0x1a, 0xdf, 0x04, 0x00,
	0x00,
}
