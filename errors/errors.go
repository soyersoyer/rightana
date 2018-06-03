package errors

import (
	"fmt"
)

// The possible errors
var (
	RegistrationDisabled    = &Error{"Registration disabled", 403, "", ""}
	InvalidEmail            = &Error{"Invalid email", 400, "", ""}
	InvalidUsername         = &Error{"Invalid username", 400, "", ""}
	PasswordNotMatch        = &Error{"Password not match", 403, "", ""}
	PasswordTooShort        = &Error{"Password too short", 400, "", ""}
	PasswordChangeDisabled  = &Error{"Password change disabled for this account", 403, "", ""}
	UserNotExist            = &Error{"User not exist", 404, "", ""}
	UserNameExist           = &Error{"User Name exist", 403, "", ""}
	UserEmailExist          = &Error{"User Email exist", 403, "", ""}
	UserDeletionDisabled    = &Error{"User deletion disabled for this account", 403, "", ""}
	UserIsTheLastAdmin      = &Error{"User is the last admin", 403, "", ""}
	AccessDenied            = &Error{"Access denied", 403, "", ""}
	InputDecodeFailed       = &Error{"Input decode failed", 400, "", ""}
	AuthtokenNotExist       = &Error{"Authtoken not exist", 403, "", ""}
	AuthtokenExpired        = &Error{"Authtoken expired", 403, "", ""}
	DBError                 = &Error{"DB error", 500, "", ""}
	BotsDontMatter          = &Error{"Bots don't matter", 403, "", ""}
	CollectionNotExist      = &Error{"Collection not exist", 404, "", ""}
	CollectionLimitExceeded = &Error{"Collection limit exceeded", 403, "", ""}
	SessionNotExist         = &Error{"Session not exist", 404, "", ""}
	TeammateExist           = &Error{"Teammate exist", 403, "", ""}
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
