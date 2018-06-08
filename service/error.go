package service

import (
	"fmt"
)

// The possible errors
var (
	ErrRegistrationDisabled    = &Error{"Registration disabled", 403, "", ""}
	ErrInvalidEmail            = &Error{"Invalid email", 400, "", ""}
	ErrInvalidUsername         = &Error{"Invalid username", 400, "", ""}
	ErrInvalidCollectionName   = &Error{"Invalid collection name", 400, "", ""}
	ErrPasswordNotMatch        = &Error{"Password not match", 403, "", ""}
	ErrPasswordTooShort        = &Error{"Password too short", 400, "", ""}
	ErrPasswordChangeDisabled  = &Error{"Password change disabled for this account", 403, "", ""}
	ErrUserNotExist            = &Error{"User not exist", 404, "", ""}
	ErrUserNameExist           = &Error{"User Name exist", 403, "", ""}
	ErrUserEmailExist          = &Error{"User Email exist", 403, "", ""}
	ErrUserDeletionDisabled    = &Error{"User deletion disabled for this account", 403, "", ""}
	ErrUserIsTheLastAdmin      = &Error{"User is the last admin", 403, "", ""}
	ErrAccessDenied            = &Error{"Access denied", 403, "", ""}
	ErrInputDecodeFailed       = &Error{"Input decode failed", 400, "", ""}
	ErrAuthtokenNotExist       = &Error{"Authtoken not exist", 403, "", ""}
	ErrAuthtokenExpired        = &Error{"Authtoken expired", 403, "", ""}
	ErrDB                      = &Error{"DB error", 500, "", ""}
	ErrBotsDontMatter          = &Error{"Bots don't matter", 403, "", ""}
	ErrCollectionNotExist      = &Error{"Collection not exist", 404, "", ""}
	ErrCollectionLimitExceeded = &Error{"Collection limit exceeded", 403, "", ""}
	ErrCollectionNameExist     = &Error{"Collection name exists", 403, "", ""}
	ErrSessionNotExist         = &Error{"Session not exist", 404, "", ""}
	ErrTeammateExist           = &Error{"Teammate exist", 403, "", ""}
)

// Error is the Extended error struct
type Error struct {
	Message    string
	Code       int
	Thing      string
	Additional string
}

func (e *Error) Error() string {
	msg := e.Message
	if e.Thing != "" {
		msg += fmt.Sprintf(" (%v)", e.Thing)
	}
	if e.Additional != "" {
		msg += " " + e.Additional
	}
	return msg
}

// HTTPMessage returns the default HTTP error message
func (e *Error) HTTPMessage() string {
	if e.Thing != "" {
		return fmt.Sprintf("%v (%v)", e.Message, e.Thing)
	}
	return e.Message
}

// T sets the thing which causes the error
func (e *Error) T(thing string) *Error {
	return &Error{e.Message, e.Code, thing, e.Additional}
}

// Wrap add more information to the error
func (e *Error) Wrap(v ...interface{}) *Error {
	return &Error{e.Message, e.Code, e.Thing, fmt.Sprint(v...)}
}
