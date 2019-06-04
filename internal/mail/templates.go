package mail

import (
	"bytes"
	"strconv"
	"text/template"
	"time"
)

var (
	verifyEmail   *template.Template
	resetPassword *template.Template
)

func init() {
	header :=
		`<!DOCTYPE html>
		<html>
		<head>
			<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		</head>

		<body>
		<p>Hi <b>{{.DisplayName}}</b>,</p>`
	footer := `
		<p>Not working? Try copying and pasting it to your browser.</p>
		<p>© {{.YearStr}} <a target="_blank" href="{{.AppURL}}" rel="noopener">{{.AppName}}</a></p>
		</body>
		</html>`

	verifyEmail = template.Must(template.New("verifyEmail").Parse(`
		{{template "header" .}}
		<p>Please click the following link to verify your email address:</p>
		<p><a href="{{template "verifyEmailLink" .}}">{{template "verifyEmailLink" .}}</a></p>
		{{template "footer" .}}
	`))
	verifyEmailLink := `{{.AppURL}}/{{.UserName}}/verify-email?verification_key={{.VerificationKey}}`
	template.Must(verifyEmail.New("header").Parse(header))
	template.Must(verifyEmail.New("verifyEmailLink").Parse(verifyEmailLink))
	template.Must(verifyEmail.New("footer").Parse(footer))

	resetPassword = template.Must(template.New("resetPassword").Parse(`
		{{template "header" .}}
		<p>Please click the following link to change your password within <b>{{.ExpireMinutes}} minutes</b>:</p>
		<p><a href="{{template "resetPasswordLink" .}}">{{template "resetPasswordLink" .}}</a></p>
		{{template "footer" .}}
	`))
	resetPasswordLink := `{{.AppURL}}/{{.UserName}}/reset-password?reset_key={{.ResetKey}}`
	template.Must(resetPassword.New("header").Parse(header))
	template.Must(resetPassword.New("resetPasswordLink").Parse(resetPasswordLink))
	template.Must(resetPassword.New("footer").Parse(footer))
}

func getResetPasswordBody(userName, displayName, resetKey string, expireMinutes int) (string, error) {
	type params struct {
		UserName      string
		DisplayName   string
		ResetKey      string
		ExpireMinutes int
		AppURL        string
		AppName       string
		YearStr       string
	}
	body := &bytes.Buffer{}
	err := resetPassword.Execute(body, &params{
		userName, displayName, resetKey, expireMinutes,
		config.AppURL, config.AppName, getYearStr()})
	if err != nil {
		return "", err
	}
	return body.String(), nil
}

func getVerifyEmailBody(userName, displayName, verificationKey string) (string, error) {
	type params struct {
		UserName        string
		DisplayName     string
		VerificationKey string
		AppURL          string
		AppName         string
		YearStr         string
	}
	body := &bytes.Buffer{}
	err := verifyEmail.Execute(body, &params{
		userName, displayName, verificationKey,
		config.AppURL, config.AppName, getYearStr()})
	if err != nil {
		return "", err
	}
	return body.String(), nil
}

func getYearStr() string {
	return strconv.Itoa(time.Now().Year())
}
