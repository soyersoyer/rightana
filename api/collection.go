package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/k20a/errors"
	"github.com/soyersoyer/k20a/service"
)

type collectionT struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getCollectionSummariesE(w http.ResponseWriter, r *http.Request) error {
	ownerEmail := getUserEmailCtx(r.Context())
	summary, err := service.GetCollectionSummariesByUserEmail(ownerEmail)
	if err != nil {
		return err
	}
	return respond(w, summary)
}

var getCollectionSummaries = handleError(getCollectionSummariesE)

func createCollectionE(w http.ResponseWriter, r *http.Request) error {
	var input collectionT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	ownerEmail := getUserEmailCtx(r.Context())
	collection, err := service.CreateCollection(ownerEmail, input.Name)
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
			collectionID := chi.URLParam(r, "collectionID")
			collection, err := service.GetCollection(collectionID)
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
			userEmail := getUserEmailCtx(r.Context())
			collection := getCollectionCtx(r.Context())
			if err := service.CollectionReadAccessCheck(collection, userEmail); err != nil {
				return err
			}
			next.ServeHTTP(w, r)
			return nil
		}))
}

func collectionWriteAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			userEmail := getUserEmailCtx(r.Context())
			collection := getCollectionCtx(r.Context())
			if err := service.CollectionWriteAccessCheck(collection, userEmail); err != nil {
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
		return errors.InputDecodeFailed.Wrap(err)
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

type teammateT struct {
	Email string `json:"email"`
}

func getTeammatesE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())
	teammates := []*teammateT{}
	for _, v := range collection.Teammates {
		teammates = append(teammates, &teammateT{v.Email})
	}
	return respond(w, teammates)
}

var getTeammates = handleError(getTeammatesE)

func addTeammateE(w http.ResponseWriter, r *http.Request) error {
	collection := getCollectionCtx(r.Context())
	var input teammateT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}
	if err := service.AddTeammate(collection, input.Email); err != nil {
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
		return errors.InputDecodeFailed.Wrap(err)
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
		return errors.InputDecodeFailed.Wrap(err)
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
		return errors.InputDecodeFailed.Wrap(err)
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
		return errors.InputDecodeFailed.Wrap(err)
	}

	collection := getCollectionCtx(r.Context())
	data, err := service.GetPageviews(collection, input.SessionKey)
	if err != nil {
		return err
	}
	return respond(w, data)
}

var getPageviews = handleError(getPageviewsE)
