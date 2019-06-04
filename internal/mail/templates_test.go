package mail

import (
	"strconv"
	"strings"
	"testing"
)

func TestResetPassword(t *testing.T) {
	user := "user"
	display := "display"
	resetKey := "resetKey"
	expireMinutes := 12
	tmpl, err := getResetPasswordBody(user, display, resetKey, 12)
	if err != nil {
		t.Error(err)
	}
	if strings.Index(tmpl, user) == -1 {
		t.Error(user)
	}
	if strings.Index(tmpl, display) == -1 {
		t.Error(display)
	}
	if strings.Index(tmpl, resetKey) == -1 {
		t.Error(resetKey)
	}
	if strings.Index(tmpl, strconv.Itoa(expireMinutes)) == -1 {
		t.Error(expireMinutes)
	}
}

func TestVerifyEmail(t *testing.T) {
	user := "user"
	display := "display"
	verificationKey := "verificationKey"
	tmpl, err := getVerifyEmailBody(user, display, verificationKey)
	if err != nil {
		t.Error(err)
	}
	if strings.Index(tmpl, user) == -1 {
		t.Error(user)
	}
	if strings.Index(tmpl, display) == -1 {
		t.Error(display)
	}
	if strings.Index(tmpl, verificationKey) == -1 {
		t.Error(verificationKey)
	}
}
