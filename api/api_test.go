package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/k20a/config"
	"github.com/soyersoyer/k20a/db/db"
	"github.com/soyersoyer/k20a/service"
)

type kv map[string]string

var (
	testDbName       = "k20a_test"
	userData         = createUserT{"admin@irl.hu", "adminlong"}
	user2Data        = createUserT{"admin2@irl.hu", "adminlong2"}
	tokenData        = createTokenT{"admin@irl.hu", "adminlong"}
	badTokenUserData = createTokenT{"adminn@irl.hu", "adminlong"}
	badTokenPwData   = createTokenT{"admin@irl.hu", "adminnlong"}
	collectionData   = collectionT{
		Name: "newcollection.org",
	}
	userAgent        = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:56.0) Gecko/20100101 Firefox/56.0"
	deviceOS         = "Ubuntu"
	browserName      = "Firefox"
	browserVersion   = "56.0"
	browserLanguage  = "hu"
	screenResolution = "1280x720"
	windowResolution = "1280x660"
	deviceType       = "desktop"
	timezone         = "Europe/Budapest"
	sessionData      = createSessionInputT{
		CollectionID:     "",
		BrowserLanguage:  browserLanguage,
		ScreenResolution: screenResolution,
		WindowResolution: windowResolution,
		DeviceType:       deviceType,
		Referrer:         "http://irl.hu",
	}
	sessionUpdateData = updateSessionInputT{
		SessionKey: "",
	}
	badSessionUpdateData = updateSessionInputT{
		SessionKey: "badsessionkey",
	}
	pageViewData = createPageviewInputT{
		SessionKey: "",
		Path:       "dl",
	}
	fromTime, _     = time.Parse("2006-01-02", time.Now().AddDate(0, 0, -2).Format("2006-01-02"))
	toTime          = fromTime.AddDate(0, 0, 7)
	collectionInput = db.CollectionDataInputT{
		From:     fromTime,
		To:       toTime,
		Bucket:   "day",
		Timezone: timezone,
	}
	pageviewInput = pageviewInputT{}
	sessionKey    = ""
	collectionID  = ""
)

func TestPublicConfig(t *testing.T) {
	w, r := postJSON(nil)
	getPublicConfig(w, r)
	var publicConfig publicConfigT
	testJSONBody(t, w, &publicConfig)
	if publicConfig.EnableRegistration != config.ActualConfig.EnableRegistration {
		t.Error(publicConfig)
	}
}

func TestCreateUserRegistrationDisabled(t *testing.T) {
	config.ActualConfig.EnableRegistration = false
	w, r := postJSON(userData)
	createUser(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Registration disabled\n")
	config.ActualConfig.EnableRegistration = true
}

func TestCreateUser(t *testing.T) {
	testCreateUserSuccess(t, userData)
}

func testCreateUserSuccess(t *testing.T, userData createUserT) string {
	w, r := postJSON(userData)
	createUser(w, r)
	testCode(t, w, 200)
	var email string
	testJSONBody(t, w, &email)
	if email != userData.Email {
		t.Error(email)
	}
	return email
}

func TestCreateUserSecondFail(t *testing.T) {
	w, r := postJSON(userData)
	createUser(w, r)
	testCode(t, w, 403)
	testBody(t, w, "User exist ("+userData.Email+")\n")
}

func TestCreateUser2(t *testing.T) {
	w, r := postJSON(user2Data)
	createUser(w, r)
	testCode(t, w, 200)
	var email string
	testJSONBody(t, w, &email)
	if email != user2Data.Email {
		t.Error(email)
	}
}

func TestCreateUserShortPw(t *testing.T) {
	userData := createUserT{"shortpw@irl.hu", "short"}
	w, r := postJSON(userData)
	createUser(w, r)
	testCode(t, w, 400)
	testBody(t, w, "Password too short\n")
}

func TestCreateUserBadEmail(t *testing.T) {
	userData := createUserT{"bademail", "short"}
	w, r := postJSON(userData)
	createUser(w, r)
	testCode(t, w, 400)
	testBody(t, w, "Invalid email ("+userData.Email+")\n")
}

func TestCreateTokenInvalidUser(t *testing.T) {
	w, r := postJSON(badTokenUserData)
	createToken(w, r)

	testCode(t, w, 404)
}

func TestCreateTokenInvalidPassword(t *testing.T) {
	w, r := postJSON(badTokenPwData)
	createToken(w, r)

	testCode(t, w, 403)
}

func testCreateTokenSuccess(t *testing.T, tokenData createTokenT) string {
	w, r := postJSON(tokenData)
	createToken(w, r)

	testCode(t, w, 200)

	var token db.AuthToken
	testJSONBody(t, w, &token)
	if token.ID == "" {
		t.Error("token id is empty")
	}

	return token.ID
}

func TestCreateTokenSuccess(t *testing.T) {
	testCreateTokenSuccess(t, tokenData)
}

func TestDeleteToken(t *testing.T) {
	token := testCreateTokenSuccess(t, tokenData)
	w, r := postJSON(nil)
	r = getReqWithRouteContext(r, kv{"token": token})
	deleteToken(w, r)
	testCode(t, w, 200)
	var tokenRet string
	testJSONBody(t, w, &tokenRet)
	if tokenRet != token {
		t.Error(tokenRet)
	}
}

func TestDeleteTokenFail(t *testing.T) {
	w, r := postJSON(nil)
	r = getReqWithRouteContext(r, kv{"token": "notoken"})
	deleteToken(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Authtoken not exist (notoken)\n")
}

func TestUserBaseHandlerBadEmail(t *testing.T) {
	w, r := postJSON(nil)
	r = setEmail(r, "adminn@irl.hu")
	userBaseHandler(getNoHandler(t)).ServeHTTP(w, r)
	testCode(t, w, 404)
	testBody(t, w, "User not exist (adminn@irl.hu)\n")
}

func TestUserBaseHandler(t *testing.T) {
	email := "admin@irl.hu"
	w, r := postJSON(nil)
	r = setEmail(r, email)
	userBaseHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserCtx(r.Context())
		if user.Email != email {
			t.Error(user)
		}
	})).ServeHTTP(w, r)
}

func TestUserAccessHandler(t *testing.T) {
	email := "admin@irl.hu"
	user := getDbUser(email)
	w, r := postJSON(nil)
	r = setUserEmailReq(r, user.Email)
	r = setEmail(r, email)
	userBaseHandler(userAccessHandler(getNullHandler())).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestUserAccessHandlerBad(t *testing.T) {
	email := "admin@irl.hu"
	user := getDbUser(email)
	w, r := postJSON(nil)
	r = setUserEmailReq(r, user.Email+"1")
	r = setEmail(r, email)
	userBaseHandler(userAccessHandler(getNoHandler(t))).ServeHTTP(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Access denied\n")
}

func TestUpdateUserPassword(t *testing.T) {
	email := "admin@irl.hu"
	updateUserPwShortData := updateUserPasswordT{"adminlong", "admin"}
	updateUserPwData := updateUserPasswordT{"adminlong", "adminnlong"}
	updateUserPwDataBack := updateUserPasswordT{"adminnlong", "adminlong"}
	testCreateTokenSuccess(t, tokenData)
	var emailOut string

	w, r := postJSON(updateUserPwShortData)
	r = setEmail(r, email)
	userBaseHandler(http.HandlerFunc(updateUserPassword)).ServeHTTP(w, r)
	testCode(t, w, 400)
	testBody(t, w, "Password too short\n")

	w, r = postJSON(updateUserPwData)
	r = setEmail(r, email)
	userBaseHandler(http.HandlerFunc(updateUserPassword)).ServeHTTP(w, r)
	testCode(t, w, 200)
	testJSONBody(t, w, &emailOut)
	if emailOut != email {
		t.Error(emailOut)
	}

	w, r = postJSON(updateUserPwData)
	r = setEmail(r, email)
	userBaseHandler(http.HandlerFunc(updateUserPassword)).ServeHTTP(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Password not match\n")

	w, r = postJSON(updateUserPwDataBack)
	r = setEmail(r, email)
	userBaseHandler(http.HandlerFunc(updateUserPassword)).ServeHTTP(w, r)
	testCode(t, w, 200)

	testJSONBody(t, w, &emailOut)
	if emailOut != email {
		t.Error(emailOut)
	}
}

func createCollectionSuccess(t *testing.T, email string, collection *collectionT) {
	name := collection.Name
	w, r := postJSON(collection)
	user := getDbUser(email)
	r = setUserEmailReq(r, user.Email)
	createCollection(w, r)
	testCode(t, w, 200)
	testJSONBody(t, w, &collection)
	if collection.Name != name {
		t.Error(collection)
	}
}

func TestDeleteUser(t *testing.T) {
	user := createUserT{
		Email:    "deleteuser@irl.hu",
		Password: "deleteuser",
	}
	collection := collectionT{
		Name: "azaz.org",
	}
	var emailOut string
	testCreateUserSuccess(t, user)
	token := testCreateTokenSuccess(t, createTokenT(user))

	createCollectionSuccess(t, user.Email, &collection)

	input := deleteUserInputT{
		Password: user.Password,
	}
	w, r := postJSON(input)
	r = setEmail(r, user.Email)
	userBaseHandler(http.HandlerFunc(deleteUser)).ServeHTTP(w, r)
	testCode(t, w, 200)
	testJSONBody(t, w, &emailOut)
	if emailOut != user.Email {
		t.Error(emailOut)
	}

	if _, err := db.GetAuthToken(token); err == nil {
		t.Error("token exists")
	}
	if _, err := db.GetCollection(collection.ID); err == nil {
		t.Error("collection exists")
	}
}

func TestLoggedInHandlerTokenNotSet(t *testing.T) {
	w, r := postJSON(nil)
	loggedOnlyHandler(getNoHandler(t)).ServeHTTP(w, r)

	testCode(t, w, 403)
	testBody(t, w, "Authtoken expired\n")
}

func TestLoggedInHandlerTokenInvalid(t *testing.T) {
	w, r := postJSON(nil)
	setAuthToken(r, "INVALIDTOKEN")
	loggedOnlyHandler(getNoHandler(t)).ServeHTTP(w, r)

	testCode(t, w, 403)
	testBody(t, w, "Authtoken expired\n")
}

func TestLoggedInHandlerTokenExpired(t *testing.T) {
	token := testCreateTokenSuccess(t, tokenData)

	dbToken, err := db.GetAuthToken(token)
	if err != nil {
		t.Error(err)
	}
	dbToken.TTL = 0
	dbToken.Created -= 3 * 60 * 60 * 24
	db.UpdateAuthToken(dbToken)
	w, r := postJSON(nil)
	setAuthToken(r, token)
	loggedOnlyHandler(getNoHandler(t)).ServeHTTP(w, r)

	testCode(t, w, 403)
	testBody(t, w, "Authtoken expired\n")
}

func TestLoggedInHandlerSuccess(t *testing.T) {
	token := testCreateTokenSuccess(t, tokenData)
	user := getDbUser(userData.Email)
	w, r := postJSON(nil)
	setAuthToken(r, token)
	loggedOnlyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userEmail := getUserEmailReq(r)
		if userEmail != user.Email {
			t.Error("bad useremail", userEmail, user.Email)
		}
	})).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestGetCollectionZero(t *testing.T) {
	w, r := postJSON(nil)
	user := getDbUser(userData.Email)
	r = setUserEmailReq(r, user.Email)
	getCollections(w, r)
	testCode(t, w, 200)
	collections := []service.CollectionSummaryT{}
	testJSONBody(t, w, &collections)
	if len(collections) != 0 {
		t.Error(w.Body, collections)
	}
}

func TestCreateCollectionSuccess(t *testing.T) {
	createCollectionSuccess(t, userData.Email, &collectionData)
}

func TestGetCollectionsOne(t *testing.T) {
	w, r := postJSON(nil)
	user := getDbUser(userData.Email)
	r = setUserEmailReq(r, user.Email)
	getCollections(w, r)
	testCode(t, w, 200)
	collections := []service.CollectionSummaryT{}
	testJSONBody(t, w, &collections)
	if len(collections) != 1 {
		t.Error(w.Body)
		t.Error(collections)
	}
	collection := collections[0]
	if collection.PageviewPercent != 0.0 {
		t.Error(collection)
	}
	collectionID = collection.ID
}

func TestGetCollection(t *testing.T) {
	w, r := postJSON(nil)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(getCollection)).ServeHTTP(w, r)
	var collection collectionT
	testJSONBody(t, w, &collection)
	if collection.Name != collectionData.Name || collection.ID != collectionID {
		t.Error(collection)
	}
}

func TestUpdateCollection(t *testing.T) {
	collection := collectionT{
		Name: "NewName",
	}
	createCollectionSuccess(t, userData.Email, &collection)
	collection.Name = "NewName2"
	w, r := postJSON(collection)
	r = setCollectionID(r, collection.ID)
	collectionBaseHandler(http.HandlerFunc(updateCollection)).ServeHTTP(w, r)
	var updated collectionT
	testCode(t, w, 200)
	testJSONBody(t, w, &updated)
	if collection.Name != updated.Name || collection.ID != updated.ID {
		t.Error(collection, updated)
	}
}

func TestDeleteCollection(t *testing.T) {
	collection := collectionT{
		Name: "NewName",
	}
	createCollectionSuccess(t, userData.Email, &collection)
	w, r := postJSON(nil)
	r = setCollectionID(r, collection.ID)
	collectionBaseHandler(http.HandlerFunc(deleteCollection)).ServeHTTP(w, r)
	var collectionID string
	testJSONBody(t, w, &collectionID)
	if collectionID != collection.ID {
		t.Error(collectionID, collection.ID)
	}
}

func TestCreateSession(t *testing.T) {
	sessionData.CollectionID = collectionID
	w, r := postJSON(sessionData)
	r.Header.Set("User-Agent", userAgent)
	createSession(w, r)
	testCode(t, w, 200)
}

func TestCollectionBaseHandler(t *testing.T) {
	w, r := postJSON(nil)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collection := getCollectionCtx(r.Context())
		if collection.ID != collectionID {
			t.Error("Bad collection id")
		}
	})).ServeHTTP(w, r)
}

func TestCollectionReadAccessHandler(t *testing.T) {
	w, r := postJSON(nil)
	user := getDbUser(userData.Email)
	r = setUserEmailReq(r, user.Email)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(collectionReadAccessHandler(getNullHandler())).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestCollectionWriteAccessHandler(t *testing.T) {
	w, r := postJSON(nil)
	user := getDbUser(userData.Email)
	r = setUserEmailReq(r, user.Email)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(collectionWriteAccessHandler(getNullHandler())).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestCollectionReadAccessHandlerNoRight(t *testing.T) {
	w, r := postJSON(nil)
	user2 := getDbUser(user2Data.Email)
	r = setUserEmailReq(r, user2.Email)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(collectionReadAccessHandler(getNoHandler(t))).ServeHTTP(w, r)
	testCode(t, w, 403)
}

func TestCollectionWriteAccessHandlerNoRight(t *testing.T) {
	w, r := postJSON(nil)
	user2 := getDbUser(user2Data.Email)
	r = setUserEmailReq(r, user2.Email)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(collectionWriteAccessHandler(getNoHandler(t))).ServeHTTP(w, r)
	testCode(t, w, 403)
}

func TestAddTeammate(t *testing.T) {
	notfoundTeammate := teammateT{Email: "notfound@irl.hu"}
	teammate := teammateT{Email: user2Data.Email}

	w, r := postJSON(notfoundTeammate)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(addTeammate)).ServeHTTP(w, r)
	testCode(t, w, 404)
	testBody(t, w, "User not exist ("+notfoundTeammate.Email+")\n")

	addTeammateSuccess(t, collectionID, &teammate)

	w, r = postJSON(teammate)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(addTeammate)).ServeHTTP(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Teammate exist ("+user2Data.Email+")\n")
}

func addTeammateSuccess(t *testing.T, collectionID string, teammate *teammateT) {
	w, r := postJSON(teammate)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(addTeammate)).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestGetCollaborators(t *testing.T) {
	w, r := postJSON(nil)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(getTeammates)).ServeHTTP(w, r)

	var teammates []*teammateT
	testJSONBody(t, w, &teammates)
	if len(teammates) != 1 {
		t.Error(len(teammates), "!=", 1)
	}
	if teammates[0].Email != user2Data.Email {
		t.Error(teammates[0].Email, "!=", user2Data.Email)
	}
}

func TestTeammateCollectionReadAccess(t *testing.T) {
	w, r := postJSON(nil)
	user2 := getDbUser(user2Data.Email)
	r = setUserEmailReq(r, user2.Email)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(collectionReadAccessHandler(getNullHandler())).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestTeammateCollectionWriteAccessNoRight(t *testing.T) {
	w, r := postJSON(nil)
	user2 := getDbUser(user2Data.Email)
	r = setUserEmailReq(r, user2.Email)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(collectionWriteAccessHandler(getNoHandler(t))).ServeHTTP(w, r)
	testCode(t, w, 403)
}

func TestRemoveTeammate(t *testing.T) {
	w, r := postJSON(nil)
	r = getReqWithRouteContext(r, kv{"email": user2Data.Email, "collectionID": collectionID})
	collectionBaseHandler(http.HandlerFunc(removeTeammate)).ServeHTTP(w, r)
	testCode(t, w, 200)

	w, r = postJSON(nil)
	r = getReqWithRouteContext(r, kv{"email": user2Data.Email, "collectionID": collectionID})
	collectionBaseHandler(http.HandlerFunc(removeTeammate)).ServeHTTP(w, r)
	testCode(t, w, 404)
	testBody(t, w, "User not exist ("+user2Data.Email+")\n")

	w, r = postJSON(nil)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(getTeammates)).ServeHTTP(w, r)

	var teammates []*teammateT
	testJSONBody(t, w, &teammates)
	if len(teammates) != 0 {
		t.Error(len(teammates), "!=", 0)
	}
}

func TestDeleteUserAndTeammate(t *testing.T) {
	user1 := createUserT{
		Email:    "deleteuser1@irl.hu",
		Password: "deleteuser1",
	}
	user2 := createUserT{
		Email:    "deleteuser2@irl.hu",
		Password: "deleteuser2",
	}
	collection := collectionT{
		Name: "azaz.org",
	}

	testCreateUserSuccess(t, user1)
	testCreateUserSuccess(t, user2)
	createCollectionSuccess(t, user1.Email, &collection)

	teammate := teammateT{Email: user2.Email}
	addTeammateSuccess(t, collection.ID, &teammate)

	input := deleteUserInputT{Password: user2.Password}
	w, r := postJSON(input)
	r = setEmail(r, user2.Email)
	userBaseHandler(http.HandlerFunc(deleteUser)).ServeHTTP(w, r)
	testCode(t, w, 200)

	coll, err := db.GetCollection(collection.ID)
	if err != nil {
		t.Error(err)
	}
	if db.GetTeammate(coll, user2.Email) != nil {
		t.Errorf("%s is collaborator already", user2.Email)
	}

	input = deleteUserInputT{Password: user1.Password}
	w, r = postJSON(input)
	r = setEmail(r, user1.Email)
	userBaseHandler(http.HandlerFunc(deleteUser)).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestGetSessions(t *testing.T) {
	w, r := postJSON(collectionInput)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(getSessions)).ServeHTTP(w, r)
	testCode(t, w, 200)
	sessions := []db.SessionDataT{}
	testJSONBody(t, w, &sessions)
	if len(sessions) != 1 {
		t.Error(w.Body, sessions, collectionInput)
	}
	session := sessions[0]
	if session.UserAgent != userAgent {
		t.Error(session)
	}
	if session.BrowserName != browserName {
		t.Error(session)
	}
	if session.BrowserVersion != browserVersion {
		t.Error(session)
	}
	if session.BrowserLanguage != browserLanguage {
		t.Error(session)
	}
	sessionKey = session.Key
}

func TestCreatePageView(t *testing.T) {
	pageViewData.CollectionID = collectionID
	pageViewData.SessionKey = sessionKey
	w, r := postJSON(pageViewData)
	r.Header.Set("User-Agent", userAgent)
	createPageview(w, r)
	testCode(t, w, 200)
}

func TestUpdateSession(t *testing.T) {
	sessionUpdateData.CollectionID = collectionID
	sessionUpdateData.SessionKey = sessionKey
	w, r := postJSON(sessionUpdateData)
	updateSession(w, r)
	testCode(t, w, 200)
}

func TestGetPageViews(t *testing.T) {
	pageviewInput.SessionKey = sessionKey
	w, r := postJSON(pageviewInput)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(getPageviews)).ServeHTTP(w, r)
	testCode(t, w, 200)
	pageViews := []db.PageviewDataT{}
	testJSONBody(t, w, &pageViews)
	if len(pageViews) != 1 {
		t.Error(w.Body)
		t.Error(pageViews)
	}
}

/*
func TestGetCollectionData(t *testing.T) {
	w, r := postJSON(collectionInput)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(getCollectionData)).ServeHTTP(w, r)
	testCode(t, w, 200)
	var output db.CollectionDataT
	testJSONBody(t, w, &output)

	if output.ID == "" {
		t.Error(output.ID)
	}
	if output.Name == "" {
		t.Error(output.Name)
	}
	if len(output.SessionSums) != 8 {
		t.Error(output.SessionSums)
	}
	ss := output.SessionSums[2]
	if ss.Bucket != time.Now().Format("2006-01-02") || ss.Count != 1 {
		t.Error(ss)
	}
	if len(output.PageviewSums) != 8 {
		t.Error(output.PageviewSums)
	}
	pvs := output.PageviewSums[2]
	if pvs.Bucket != time.Now().Format("2006-01-02") || pvs.Count != 1 {
		t.Error(pvs)
	}
}
*/
/*
func TestGetCollectionStatData(t *testing.T) {
	w, r := postJSON(collectionInput)
	r = setCollectionID(r, collectionID)
	collectionBaseHandler(http.HandlerFunc(getCollectionStatData)).ServeHTTP(w, r)
	testCode(t, w, 200)
	var output collectionStatDataT
	testJsonBody(t, w, &output)
	if output.SessionTotal.Count != 1 {
		t.Error(output.SessionTotal)
	}
	if output.PageviewTotal.Count != 1 {
		t.Error(output.PageviewTotal)
	}
	if output.AvgSessionLength.Count != 0 {
		t.Error(output.AvgSessionLength)
	}
	if output.BounceRate.Percent != 1.0 {
		t.Error(output.BounceRate)
	}

	if len(output.PageSums) != 1 {
		t.Error(output.PageSums)
	}
	ps := output.PageSums[0]
	if ps.Hostname != "irl.hu" || ps.Path != "dl" || ps.Count != 1 || ps.Percent != 1.0 {
		t.Error(ps)
	}
	testSums(t, output.ReferrerSums, pageViewData.Referrer)
	testSums(t, output.DeviceOSSums, deviceOS)
	testSums(t, output.BrowserSums, browserName)
	testSums(t, output.PageviewCountSums, "1")
	testSums(t, output.ScreenResolutionSums, screenResolution)
	testSums(t, output.WindowResolutionSums, windowResolution)
	testSums(t, output.DeviceTypeSums, deviceType)
	testSums(t, output.CountrySums, "")
	testSums(t, output.CitySums, "")
	testSums(t, output.AsNameSums, "")
	testSums(t, output.BrowserLanguageSums, browserLanguage)
}

func testSums(t *testing.T, sums []sumT, name string) {
	if len(sums) != 1 {
		t.Error(sums)
	}
	sum := sums[0]
	if sum.Name != name || sum.Count != 1 || sum.Percent != 1.0 {
		t.Error(sum, name)
	}
}
*/

func TestMain(m *testing.M) {
	runCmd("", "rm", "-rf", "data")
	runCmd("", "mkdir", "data")
	db.InitDatabase("data")
	ret := m.Run()
	if ret == 0 {
		runCmd("", "rm", "-rf", "data")
	}
	os.Exit(ret)
}

func runCmd(stdin, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()

	if len(stdin) != 0 {
		cmd.Stdin = strings.NewReader(stdin)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("failed running:", command, args)
		fmt.Println(stdout.String())
		fmt.Println(stderr.String())
		return err
	}

	return nil
}

func postJSON(data interface{}) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	b, _ := json.Marshal(data)
	r := httptest.NewRequest("POST", "/users", bytes.NewBuffer(b))
	r.Header.Set("content-type", "application/json")
	return w, r
}

func testPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	f()
}

func testCode(t *testing.T, w *httptest.ResponseRecorder, code int) {
	if w.Code != code {
		t.Error(w.Body)
		t.Error(code, "!=", w.Code)
	}
}

func testBody(t *testing.T, w *httptest.ResponseRecorder, body string) {
	if w.Body.String() != body {
		t.Error(w.Body, "!=", body)
	}
}

func testJSONBody(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Error(err)
	}
}

func getReqWithContext(r *http.Request, key interface{}, val interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, val))
}

func getReqWithRouteContext(r *http.Request, kvs kv) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range kvs {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func setAuthToken(r *http.Request, token string) {
	r.Header.Add("Authorization", token)
}

func getNullHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

func getNoHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("noHandler run")
	})
}

func setCollectionID(r *http.Request, collectionID string) *http.Request {
	return getReqWithRouteContext(r, kv{"collectionID": collectionID})
}

func setEmail(r *http.Request, email string) *http.Request {
	return getReqWithRouteContext(r, kv{"email": email})
}

func setUserEmailReq(r *http.Request, userEmail string) *http.Request {
	return r.WithContext(setUserEmailCtx(r.Context(), userEmail))
}

func getUserEmailReq(r *http.Request) string {
	return getUserEmailCtx(r.Context())
}

func getDbUser(email string) *db.User {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		panic(err)
	}
	return user
}
