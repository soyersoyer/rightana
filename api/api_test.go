package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/rightana/config"
	"github.com/soyersoyer/rightana/db/db"
	"github.com/soyersoyer/rightana/service"
)

type kv map[string]string

var (
	testDbName        = "rightana_test"
	userData          = service.CreateUserT{Email: "admin@irl.hu", Name: "admin", Password: "adminlong"}
	userSameNameData  = service.CreateUserT{Email: "adminsame@irl.hu", Name: "admin", Password: "adminlong"}
	user2Data         = service.CreateUserT{Email: "admin2@irl.hu", Name: "admin2", Password: "adminlong2"}
	tokenData         = createTokenT{"admin@irl.hu", "adminlong"}
	tokenDataUserName = createTokenT{"admin", "adminlong"}
	badTokenUserData  = createTokenT{"adminn@irl.hu", "adminlong"}
	badTokenPwData    = createTokenT{"admin@irl.hu", "adminnlong"}
	collectionData    = collectionT{
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

func TestRegisterUserRegistrationDisabled(t *testing.T) {
	config.ActualConfig.EnableRegistration = false
	w, r := postJSON(userData)
	registerUser(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Registration disabled\n")
	config.ActualConfig.EnableRegistration = true
}

func TestRegisterUser(t *testing.T) {
	testRegisterUserSuccess(t, userData)
}

func testRegisterUserSuccess(t *testing.T, userData service.CreateUserT) string {
	w, r := postJSON(userData)
	registerUser(w, r)
	testCode(t, w, 200)
	var email string
	testJSONBody(t, w, &email)
	if email != userData.Email {
		t.Error(email)
	}
	return email
}

func TestRegisterUserSecondEmailFail(t *testing.T) {
	w, r := postJSON(userData)
	registerUser(w, r)
	testCode(t, w, 403)
	testBody(t, w, "User Email exist ("+userData.Email+")\n")
}

func TestRegisterUserSecondNameFail(t *testing.T) {
	w, r := postJSON(userSameNameData)
	registerUser(w, r)
	testCode(t, w, 403)
	testBody(t, w, "User Name exist ("+userSameNameData.Name+")\n")
}
func TestRegisterUser2(t *testing.T) {
	w, r := postJSON(user2Data)
	registerUser(w, r)
	testCode(t, w, 200)
	var email string
	testJSONBody(t, w, &email)
	if email != user2Data.Email {
		t.Error(email)
	}
}

func TestRegisterUserShortPw(t *testing.T) {
	userData := service.CreateUserT{
		Email:    "shortpw@irl.hu",
		Name:     "short",
		Password: "short",
	}
	w, r := postJSON(userData)
	registerUser(w, r)
	testCode(t, w, 400)
	testBody(t, w, "Password too short\n")
}

func TestRegisterUserBadEmail(t *testing.T) {
	userData := service.CreateUserT{
		Email:    "bademail",
		Name:     "short",
		Password: "short",
	}
	w, r := postJSON(userData)
	registerUser(w, r)
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

func TestCreateTokenSuccessUserName(t *testing.T) {
	testCreateTokenSuccess(t, tokenDataUserName)
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
	r = setUserName(r, "adminn")
	userBaseHandler(getNoHandler(t)).ServeHTTP(w, r)
	testCode(t, w, 404)
	testBody(t, w, "User not exist (adminn)\n")
}

func TestUserBaseHandler(t *testing.T) {
	name := "admin"
	w, r := postJSON(nil)
	r = setUserName(r, name)
	userBaseHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserCtx(r.Context())
		if user.Name != name {
			t.Error(user)
		}
	})).ServeHTTP(w, r)
}

func TestUserAccessHandler(t *testing.T) {
	name := "admin"
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, name)
	r = setUserName(r, name)
	userBaseHandler(userAccessHandler(getNullHandler())).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestUserAccessHandlerBad(t *testing.T) {
	name := "admin"
	user := getDbUserByName(name)
	w, r := postJSON(nil)
	user.ID++
	r = setLoggedInUserReq(r, user)
	r = setUserName(r, name)
	userBaseHandler(userAccessHandler(getNoHandler(t))).ServeHTTP(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Access denied\n")
}

func TestUpdateUserPassword(t *testing.T) {
	name := "admin"
	updateUserPwShortData := updateUserPasswordT{"adminlong", "admin"}
	updateUserPwData := updateUserPasswordT{"adminlong", "adminnlong"}
	updateUserPwDataBack := updateUserPasswordT{"adminnlong", "adminlong"}
	testCreateTokenSuccess(t, tokenData)
	var out string

	w, r := postJSON(updateUserPwShortData)
	r = setUserName(r, name)
	userBaseHandler(http.HandlerFunc(updateUserPassword)).ServeHTTP(w, r)
	testCode(t, w, 400)
	testBody(t, w, "Password too short\n")

	w, r = postJSON(updateUserPwData)
	r = setUserName(r, name)
	userBaseHandler(http.HandlerFunc(updateUserPassword)).ServeHTTP(w, r)
	testCode(t, w, 200)
	testJSONBody(t, w, &out)
	if out != "" {
		t.Error(out)
	}

	w, r = postJSON(updateUserPwData)
	r = setUserName(r, name)
	userBaseHandler(http.HandlerFunc(updateUserPassword)).ServeHTTP(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Password not match\n")

	w, r = postJSON(updateUserPwDataBack)
	r = setUserName(r, name)
	userBaseHandler(http.HandlerFunc(updateUserPassword)).ServeHTTP(w, r)
	testCode(t, w, 200)

	testJSONBody(t, w, &out)
	if out != "" {
		t.Error(out)
	}
}

func TestPWChangeDisabled(t *testing.T) {
	newUser := service.CreateUserT{
		Name:     "pwchangedisabled",
		Email:    "a@a.com",
		Password: "pwchangedisabled",
	}

	testRegisterUserSuccess(t, newUser)

	updateUserData := service.UserUpdateT{
		Name:            newUser.Name,
		Email:           newUser.Email,
		DisablePwChange: true,
	}

	w, r := postJSON(updateUserData)
	r = setUserName(r, newUser.Name)
	userBaseHandler(http.HandlerFunc(updateUser)).ServeHTTP(w, r)
	testCode(t, w, 200)

	updateUserPw := updateUserPasswordT{newUser.Password, newUser.Password}

	w, r = postJSON(updateUserPw)
	r = setUserName(r, newUser.Name)
	userBaseHandler(http.HandlerFunc(updateUserPassword)).ServeHTTP(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Password change disabled for this account\n")
}

func TestUserDeletionDisabled(t *testing.T) {
	newUser := service.CreateUserT{
		Name:     "userdeletiondisabled",
		Email:    "ud@ud.com",
		Password: "userdeletiondisabled",
	}

	testRegisterUserSuccess(t, newUser)

	updateUserData := service.UserUpdateT{
		Name:                newUser.Name,
		Email:               newUser.Email,
		DisableUserDeletion: true,
	}

	w, r := postJSON(updateUserData)
	r = setUserName(r, newUser.Name)
	userBaseHandler(http.HandlerFunc(updateUser)).ServeHTTP(w, r)
	testCode(t, w, 200)

	input := deleteUserInputT{
		Password: newUser.Password,
	}

	w, r = postJSON(input)
	r = setUserName(r, newUser.Name)
	userBaseHandler(http.HandlerFunc(deleteUser)).ServeHTTP(w, r)
	testCode(t, w, 403)
	testBody(t, w, "User deletion disabled for this account\n")
}

// TODO check last admin too

func TestCollectionLimit(t *testing.T) {
	newUser := service.CreateUserT{
		Name:     "collectionlimit",
		Email:    "c@c.com",
		Password: "collectionlimit",
	}

	testRegisterUserSuccess(t, newUser)

	updateUserData := service.UserUpdateT{
		Name:             newUser.Name,
		Email:            newUser.Email,
		LimitCollections: true,
		CollectionLimit:  0,
	}

	w, r := postJSON(updateUserData)
	r = setUserName(r, newUser.Name)
	userBaseHandler(http.HandlerFunc(updateUser)).ServeHTTP(w, r)
	testCode(t, w, 200)

	collection := collectionT{
		Name: "azaz.org",
	}
	w, r = postJSON(collection)
	r = setUserName(r, newUser.Name)
	userBaseHandler(http.HandlerFunc(createCollection)).ServeHTTP(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Collection limit exceeded (0)\n")
}

func createCollectionSuccess(t *testing.T, username string, collection *collectionT) {
	collName := collection.Name
	w, r := postJSON(collection)
	r = setUserName(r, username)
	userBaseHandler(http.HandlerFunc(createCollection)).ServeHTTP(w, r)
	testCode(t, w, 200)
	testJSONBody(t, w, &collection)
	if collection.Name != collName {
		t.Error(collection)
	}
}

func TestDeleteUser(t *testing.T) {
	user := service.CreateUserT{
		Email:    "deleteuser@irl.hu",
		Name:     "deleteuser",
		Password: "deleteuser",
	}
	collection := collectionT{
		Name: "azaz.org",
	}
	var emailOut string
	testRegisterUserSuccess(t, user)
	token := testCreateTokenSuccess(t, createTokenT{user.Email, user.Password})

	createCollectionSuccess(t, user.Name, &collection)

	input := deleteUserInputT{
		Password: user.Password,
	}
	w, r := postJSON(input)
	r = setUserName(r, user.Name)
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
	user := getDbUserByName(userData.Name)
	w, r := postJSON(nil)
	setAuthToken(r, token)
	loggedOnlyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedInUser := getLoggedInUserReq(r)
		if loggedInUser.ID != user.ID {
			t.Error("bad userid", loggedInUser.ID, user.ID)
		}
	})).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestGetCollectionZero(t *testing.T) {
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, userData.Name)
	getCollections(w, r)
	testCode(t, w, 200)
	collections := []service.CollectionSummaryT{}
	testJSONBody(t, w, &collections)
	if len(collections) != 0 {
		t.Error(w.Body, collections)
	}
}

func TestCreateCollectionSuccess(t *testing.T) {
	createCollectionSuccess(t, userData.Name, &collectionData)
}

func TestGetCollectionsOne(t *testing.T) {
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, userData.Name)
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
}

func TestGetCollection(t *testing.T) {
	w, r := postJSON(nil)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(getCollection))).ServeHTTP(w, r)
	var collection collectionT
	testJSONBody(t, w, &collection)
	if collection.Name != collectionData.Name {
		t.Error(collection)
	}
}

func TestUpdateCollection(t *testing.T) {
	collName := "newname"
	collection := collectionT{
		Name: collName,
	}
	createCollectionSuccess(t, userData.Name, &collection)
	collection.Name = "newname2"
	w, r := postJSON(collection)
	r = setCollectionName(r, userData.Name, collName)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(updateCollection))).ServeHTTP(w, r)
	var updated collectionT
	testCode(t, w, 200)
	testJSONBody(t, w, &updated)
	if collection.Name != updated.Name || collection.ID != updated.ID {
		t.Error(collection, updated)
	}
}

func TestDeleteCollection(t *testing.T) {
	collection := collectionT{
		Name: "newname",
	}
	createCollectionSuccess(t, userData.Name, &collection)
	w, r := postJSON(nil)
	r = setCollectionName(r, userData.Name, collection.Name)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(deleteCollection))).ServeHTTP(w, r)
	var collectionID string
	testJSONBody(t, w, &collectionID)
	if collectionID != collection.ID {
		t.Error(collectionID, collection.ID)
	}
}

func TestCreateSession(t *testing.T) {
	sessionData.CollectionID = collectionData.ID
	w, r := postJSON(sessionData)
	r.Header.Set("User-Agent", userAgent)
	createSession(w, r)
	testCode(t, w, 200)
}

func TestCollectionBaseHandler(t *testing.T) {
	w, r := postJSON(nil)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collection := getCollectionCtx(r.Context())
		if collection.ID != collectionData.ID {
			t.Error("Bad collection id")
		}
	}))).ServeHTTP(w, r)
}

func TestCollectionReadAccessHandler(t *testing.T) {
	log.Println(userData)
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, userData.Name)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(collectionReadAccessHandler(getNullHandler()))).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestCollectionWriteAccessHandler(t *testing.T) {
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, userData.Name)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(collectionWriteAccessHandler(getNullHandler()))).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestCollectionCreateAccessHandler(t *testing.T) {
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, userData.Name)
	r = setUserName(r, userData.Name)
	userBaseHandler(collectionCreateAccessHandler(getNullHandler())).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestCollectionCreateAccessHandlerNoRight(t *testing.T) {
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, user2Data.Name)
	r = setUserName(r, userData.Name)
	userBaseHandler(collectionCreateAccessHandler(getNoHandler(t))).ServeHTTP(w, r)
	testCode(t, w, 403)
}

func TestCollectionReadAccessHandlerNoRight(t *testing.T) {
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, user2Data.Name)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(collectionReadAccessHandler(getNoHandler(t)))).ServeHTTP(w, r)
	testCode(t, w, 403)
}

func TestCollectionWriteAccessHandlerNoRight(t *testing.T) {
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, user2Data.Name)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(collectionWriteAccessHandler(getNoHandler(t)))).ServeHTTP(w, r)
	testCode(t, w, 403)
}

func TestAddTeammate(t *testing.T) {
	notfoundTeammate := service.TeammateT{Email: "notfound@irl.hu"}
	teammate := service.TeammateT{Email: user2Data.Email}

	w, r := postJSON(notfoundTeammate)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(addTeammate))).ServeHTTP(w, r)
	testCode(t, w, 404)
	testBody(t, w, "User not exist ("+notfoundTeammate.Email+")\n")

	addTeammateSuccess(t, userData.Name, collectionData.Name, &teammate)

	w, r = postJSON(teammate)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(addTeammate))).ServeHTTP(w, r)
	testCode(t, w, 403)
	testBody(t, w, "Teammate exist ("+user2Data.Email+")\n")
}

func addTeammateSuccess(t *testing.T, userName, collectionName string, teammate *service.TeammateT) {
	w, r := postJSON(teammate)
	r = setCollectionName(r, userName, collectionName)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(addTeammate))).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestGetCollaborators(t *testing.T) {
	w, r := postJSON(nil)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(getTeammates))).ServeHTTP(w, r)

	var teammates []*service.TeammateT
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
	r = setLoggedInUserWithNameReq(r, user2Data.Name)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(collectionReadAccessHandler(getNullHandler()))).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestTeammateCollectionWriteAccessNoRight(t *testing.T) {
	w, r := postJSON(nil)
	r = setLoggedInUserWithNameReq(r, user2Data.Name)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(collectionWriteAccessHandler(getNoHandler(t)))).ServeHTTP(w, r)
	testCode(t, w, 403)
}

func TestRemoveTeammate(t *testing.T) {
	w, r := postJSON(nil)
	r = getReqWithRouteContext(r, kv{"email": user2Data.Email, "name": userData.Name, "collectionName": collectionData.Name})
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(removeTeammate))).ServeHTTP(w, r)
	testCode(t, w, 200)

	w, r = postJSON(nil)
	r = getReqWithRouteContext(r, kv{"email": user2Data.Email, "name": userData.Name, "collectionName": collectionData.Name})
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(removeTeammate))).ServeHTTP(w, r)
	testCode(t, w, 404)
	testBody(t, w, "User not exist ("+user2Data.Email+")\n")

	w, r = postJSON(nil)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(getTeammates))).ServeHTTP(w, r)

	var teammates []*service.TeammateT
	testJSONBody(t, w, &teammates)
	if len(teammates) != 0 {
		t.Error(len(teammates), "!=", 0)
	}
}

func TestDeleteUserAndTeammate(t *testing.T) {
	user1 := service.CreateUserT{
		Email:    "deleteuser1@irl.hu",
		Name:     "deleteuser1",
		Password: "deleteuser1",
	}
	user2 := service.CreateUserT{
		Email:    "deleteuser2@irl.hu",
		Name:     "deleteuser2",
		Password: "deleteuser2",
	}
	collection := collectionT{
		Name: "azaz.org",
	}

	testRegisterUserSuccess(t, user1)
	testRegisterUserSuccess(t, user2)
	createCollectionSuccess(t, user1.Name, &collection)

	teammate := service.TeammateT{Email: user2.Email}
	addTeammateSuccess(t, user1.Name, collection.Name, &teammate)

	input := deleteUserInputT{Password: user2.Password}
	w, r := postJSON(input)
	r = setUserName(r, user2.Name)
	userBaseHandler(http.HandlerFunc(deleteUser)).ServeHTTP(w, r)
	testCode(t, w, 200)

	/*coll, err := db.GetCollection(collection.ID)
	if err != nil {
		t.Error(err)
	}
	if db.GetTeammate(coll, user2.ID) != nil {
		t.Errorf("%s is teammate already", user2.ID)
	}*/

	input = deleteUserInputT{Password: user1.Password}
	w, r = postJSON(input)
	r = setUserName(r, user1.Name)
	userBaseHandler(http.HandlerFunc(deleteUser)).ServeHTTP(w, r)
	testCode(t, w, 200)
}

func TestGetSessions(t *testing.T) {
	w, r := postJSON(collectionInput)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(getSessions))).ServeHTTP(w, r)
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
	pageViewData.CollectionID = collectionData.ID
	pageViewData.SessionKey = sessionKey
	w, r := postJSON(pageViewData)
	r.Header.Set("User-Agent", userAgent)
	createPageview(w, r)
	testCode(t, w, 200)
}

func TestUpdateSession(t *testing.T) {
	sessionUpdateData.CollectionID = collectionData.ID
	sessionUpdateData.SessionKey = sessionKey
	w, r := postJSON(sessionUpdateData)
	updateSession(w, r)
	testCode(t, w, 200)
}

func TestGetPageViews(t *testing.T) {
	pageviewInput.SessionKey = sessionKey
	w, r := postJSON(pageviewInput)
	r = setCollectionName(r, userData.Name, collectionData.Name)
	userBaseHandler(collectionBaseHandler(http.HandlerFunc(getPageviews))).ServeHTTP(w, r)
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

func setCollectionName(r *http.Request, userName, collectionName string) *http.Request {
	return getReqWithRouteContext(r, kv{"name": userName, "collectionName": collectionName})
}

func setUserName(r *http.Request, name string) *http.Request {
	return getReqWithRouteContext(r, kv{"name": name})
}

func setLoggedInUserReq(r *http.Request, user *db.User) *http.Request {
	return r.WithContext(setLoggedInUserCtx(r.Context(), user))
}

func setLoggedInUserWithNameReq(r *http.Request, name string) *http.Request {
	user := getDbUserByName(name)
	return r.WithContext(setLoggedInUserCtx(r.Context(), user))
}

func getLoggedInUserReq(r *http.Request) *db.User {
	return getLoggedInUserCtx(r.Context())
}

func getDbUserByName(name string) *db.User {
	user, err := db.GetUserByName(name)
	if err != nil {
		panic(err)
	}
	return user
}
