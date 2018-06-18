package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/rightana/service"
)

type collectionT struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getCollectionSummariesE(w http.ResponseWriter, r *http.Request) error {
	user := getUserCtx(r.Context())
	loggedInUser := getLoggedInUserCtx(r.Context())
	summary, err := service.GetCollectionSummariesByUserID(user.ID, loggedInUser.ID)
	if err != nil {
		return err
	}
	return respond(w, summary)
}

var getCollectionSummaries = handleError(getCollectionSummariesE)

func createCollectionE(w http.ResponseWriter, r *http.Request) error {
	var input collectionT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	user := getUserCtx(r.Context())
	collection, err := service.CreateCollection(user.ID, input.Name)
	if err != nil {
		return err
	}
	return respond(w, collectionT{
		Name: collection.Name,
		ID:   collection.ID,
	})
}

var createCollection = handleError(createCollectionE)

func setCollectionCtx(ctx context.Context, collection *service.Collection) context.Context {
	return context.WithValue(ctx, keyCollection, collection)
}

func getCollectionCtx(ctx context.Context) *service.Collection {
	return ctx.Value(keyCollection).(*service.Collection)
}

func collectionBaseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			collectionName := chi.URLParam(r, "collectionName")
			user := getUserCtx(r.Context())
			collection, err := service.GetCollectionByName(user, collectionName)
			if err != nil {
				return err
			}
			ctx := setCollectionCtx(r.Context(), collection)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		}))
}

func collectionReadAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			loggedInUser := getLoggedInUserCtx(r.Context())
			collection := getCollectionCtx(r.Context())
			if err := service.CollectionReadAccessCheck(collection, loggedInUser); err != nil {
				return err
			}
			next.ServeHTTP(w, r)
			return nil
		}))
}

func collectionWriteAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			loggedInUser := getLoggedInUserCtx(r.Context())
			collection := getCollectionCtx(r.Context())
			if err := service.CollectionWriteAccessCheck(collection, loggedInUser); err != nil {
				return err
			}
			next.ServeHTTP(w, r)
			return nil
		}))
}

func collectionCreateAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			user := getUserCtx(r.Context())
			loggedInUser := getLoggedInUserCtx(r.Context())
			if err := service.CollectionCreateAccessCheck(user, loggedInUser); err != nil {
				return err
			}
			next.ServeHTTP(w, r)
			return nil
		}))
}

func getCollectionE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())
	return respond(w, collectionT{
		collection.ID,
		collection.Name,
	})
}

var getCollection = handleError(getCollectionE)

func updateCollectionE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())
	var input collectionT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	if err := service.UpdateCollection(collection, input.Name); err != nil {
		return err
	}
	return respond(w, collectionT{
		collection.ID,
		collection.Name,
	})
}

var updateCollection = handleError(updateCollectionE)

func deleteCollectionE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())

	if err := service.DeleteCollection(collection); err != nil {
		return err
	}
	return respond(w, collection.ID)
}

var deleteCollection = handleError(deleteCollectionE)

func getCollectionShardsE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())
	shards, err := service.GetCollectionShards(collection)
	if err != nil {
		return err
	}
	return respond(w, shards)
}

var getCollectionShards = handleError(getCollectionShardsE)

func deleteCollectionShardE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())
	shardID := chi.URLParam(r, "shardID")
	if err := service.DeleteCollectionShard(collection, shardID); err != nil {
		return err
	}
	return respond(w, shardID)
}

var deleteCollectionShard = handleError(deleteCollectionShardE)

func getTeammatesE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())
	teammates, err := service.GetCollectionTeammates(collection)
	if err != nil {
		return err
	}
	return respond(w, teammates)
}

var getTeammates = handleError(getTeammatesE)

func addTeammateE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())
	var input service.TeammateT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}
	if err := service.AddTeammate(collection, input); err != nil {
		return err
	}
	return respond(w, input)
}

var addTeammate = handleError(addTeammateE)

func removeTeammateE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())
	email := chi.URLParam(r, "email")
	if err := service.RemoveTeammate(collection, email); err != nil {
		return err
	}
	return respond(w, email)
}

var removeTeammate = handleError(removeTeammateE)

func getCollectionDataE(w http.ResponseWriter, r *http.Request) error {
	var input service.CollectionDataInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	collection := getCollectionCtx(r.Context())
	data, err := service.GetCollectionData(collection, &input)
	if err != nil {
		return err
	}
	return respond(w, data)
}

var getCollectionData = handleError(getCollectionDataE)

func getCollectionStatDataE(w http.ResponseWriter, r *http.Request) error {
	var input service.CollectionDataInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	collection := getCollectionCtx(r.Context())
	data, err := service.GetCollectionStatData(collection, &input)
	if err != nil {
		return err
	}
	return respond(w, data)
}

var getCollectionStatData = handleError(getCollectionStatDataE)

func getSessionsE(w http.ResponseWriter, r *http.Request) error {
	var input service.CollectionDataInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	collection := getCollectionCtx(r.Context())
	sessions, err := service.GetSessions(collection, &input)
	if err != nil {
		return err
	}
	return respond(w, sessions)
}

var getSessions = handleError(getSessionsE)

type pageviewInputT struct {
	SessionKey string `json:"session_key"`
}

func getPageviewsE(w http.ResponseWriter, r *http.Request) error {
	var input pageviewInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	collection := getCollectionCtx(r.Context())
	data, err := service.GetPageviews(collection, input.SessionKey)
	if err != nil {
		return err
	}
	return respond(w, data)
}

var getPageviews = handleError(getPageviewsE)
