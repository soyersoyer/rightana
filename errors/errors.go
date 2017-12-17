package errors

import (
	"fmt"
)

var (
	RegistrationDisabled = &Error{"Registration disabled", 403, "", ""}
	InvalidEmail         = &Error{"Invalid email", 400, "", ""}
	PasswordNotMatch     = &Error{"Password not match", 403, "", ""}
	PasswordTooShort     = &Error{"Password too short", 400, "", ""}
	UserNotExist         = &Error{"User not exist", 404, "", ""}
	UserExist            = &Error{"User exist", 403, "", ""}
	AccessDenied         = &Error{"Access denied", 403, "", ""}
	InputDecodeFailed    = &Error{"Input decode failed", 400, "", ""}
	AuthtokenNotExist    = &Error{"Authtoken not exist", 403, "", ""}
	AuthtokenExpired     = &Error{"Authtoken expired", 403, "", ""}
	DBError              = &Error{"DB error", 500, "", ""}
	BotsDontMatter       = &Error{"Bots don't matter", 403, "", ""}
	CollectionNotExist   = &Error{"Collection not exist", 404, "", ""}
	SessionNotExist      = &Error{"Session not exist", 404, "", ""}
	TeammateExist        = &Error{"Teammate exist", 403, "", ""}
)

type Error struct {
	Message    string
	Code       int
	Thing      string
	Additional string
}

func (e *Error) Error() string {
	msg := e.Message
	if e.Thing != "" {
		msg += fmt.Sprintf(" (%s)", e.Thing)
	}
	if e.Additional != "" {
		msg += " " + e.Additional
	}
	return msg
}

func (e *Error) HttpMessage() string {
	if e.Thing != "" {
		return fmt.Sprintf("%s (%s)", e.Message, e.Thing)
	}
	return e.Message
}

func (base *Error) T(e string) *Error {
	return &Error{base.Message, base.Code, e, base.Additional}
}

func (base *Error) Wrap(v ...interface{}) *Error {
	return &Error{base.Message, base.Code, base.Thing, fmt.Sprint(v...)}
}
