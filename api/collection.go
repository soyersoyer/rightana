package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/k20a/errors"
	"github.com/soyersoyer/k20a/models"
)

type collectionT struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getCollectionsE(w http.ResponseWriter, r *http.Request) error {
	ownerEmail := GetUserEmail(r.Context())
	summary, err := models.GetCollectionSummaryByUserEmail(ownerEmail)
	if err != nil {
		return err
	}
	return respond(w, summary)
}

var getCollections = handleError(getCollectionsE)

func createCollectionE(w http.ResponseWriter, r *http.Request) error {
	var input collectionT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	ownerEmail := GetUserEmail(r.Context())
	collection, err := models.CreateCollection(ownerEmail, input.Name)
	if err != nil {
		return err
	}
	return respond(w, collectionT{
		Name: collection.Name,
		ID:   collection.ID,
	})
}

var createCollection = handleError(createCollectionE)

func SetCollection(ctx context.Context, collection *models.Collection) context.Context {
	return context.WithValue(ctx, keyCollection, collection)
}

func GetCollection(ctx context.Context) *models.Collection {
	return ctx.Value(keyCollection).(*models.Collection)
}

func collectionBaseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			collectionID := chi.URLParam(r, "collectionID")
			collection, err := models.GetCollection(collectionID)
			if err != nil {
				return err
			}
			ctx := SetCollection(r.Context(), collection)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		}))
}

func collectionReadAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			userEmail := GetUserEmail(r.Context())
			collection := GetCollection(r.Context())
			if err := models.CollectionReadAccessCheck(collection, userEmail); err != nil {
				return err
			}
			next.ServeHTTP(w, r)
			return nil
		}))
}

func collectionWriteAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			userEmail := GetUserEmail(r.Context())
			collection := GetCollection(r.Context())
			if err := models.CollectionWriteAccessCheck(collection, userEmail); err != nil {
				return err
			}
			next.ServeHTTP(w, r)
			return nil
		}))
}

func getCollectionE(w http.ResponseWriter, r *http.Request) error {
	collection := GetCollection(r.Context())
	return respond(w, collectionT{
		collection.ID,
		collection.Name,
	})
}

var getCollection = handleError(getCollectionE)

func updateCollectionE(w http.ResponseWriter, r *http.Request) error {
	collection := GetCollection(r.Context())
	var input collectionT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	if err := models.UpdateCollection(collection, input.Name); err != nil {
		return err
	}
	return respond(w, collectionT{
		collection.ID,
		collection.Name,
	})
}

var updateCollection = handleError(updateCollectionE)

func deleteCollectionE(w http.ResponseWriter, r *http.Request) error {
	collection := GetCollection(r.Context())

	if err := models.DeleteCollection(collection); err != nil {
		return err
	}
	return respond(w, collection.ID)
}

var deleteCollection = handleError(deleteCollectionE)

func getCollectionShardsE(w http.ResponseWriter, r *http.Request) error {
	collection := GetCollection(r.Context())
	shards, err := models.GetCollectionShards(collection)
	if err != nil {
		return err
	}
	return respond(w, shards)
}

var getCollectionShards = handleError(getCollectionShardsE)

func deleteCollectionShardE(w http.ResponseWriter, r *http.Request) error {
	collection := GetCollection(r.Context())
	shardID := chi.URLParam(r, "shardID")
	if err := models.DeleteCollectionShard(collection, shardID); err != nil {
		return err
	}
	return respond(w, shardID)
}

var deleteCollectionShard = handleError(deleteCollectionShardE)

type teammateT struct {
	Email string `json:"email"`
}

func getTeammatesE(w http.ResponseWriter, r *http.Request) error {
	collection := GetCollection(r.Context())
	teammates := []*teammateT{}
	for _, v := range collection.Teammates {
		teammates = append(teammates, &teammateT{v.Email})
	}
	return respond(w, teammates)
}

var getTeammates = handleError(getTeammatesE)

func addTeammateE(w http.ResponseWriter, r *http.Request) error {
	collection := GetCollection(r.Context())
	var input teammateT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}
	if err := models.AddTeammate(collection, input.Email); err != nil {
		return err
	}
	return respond(w, input)
}

var addTeammate = handleError(addTeammateE)

func removeTeammateE(w http.ResponseWriter, r *http.Request) error {
	collection := GetCollection(r.Context())
	email := chi.URLParam(r, "email")
	if err := models.RemoveTeammate(collection, email); err != nil {
		return err
	}
	return respond(w, email)
}

var removeTeammate = handleError(removeTeammateE)

func getCollectionDataE(w http.ResponseWriter, r *http.Request) error {
	var input models.CollectionDataInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	collection := GetCollection(r.Context())
	data, err := models.GetCollectionData(collection, &input)
	if err != nil {
		return err
	}
	return respond(w, data)
}

var getCollectionData = handleError(getCollectionDataE)

func getCollectionStatDataE(w http.ResponseWriter, r *http.Request) error {
	var input models.CollectionDataInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	collection := GetCollection(r.Context())
	data, err := models.GetCollectionStatData(collection, &input)
	if err != nil {
		return err
	}
	return respond(w, data)
}

var getCollectionStatData = handleError(getCollectionStatDataE)

func getSessionsE(w http.ResponseWriter, r *http.Request) error {
	var input models.CollectionDataInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	collection := GetCollection(r.Context())
	sessions, err := models.GetSessions(collection, &input)
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

	collection := GetCollection(r.Context())
	data, err := models.GetPageviews(collection, input.SessionKey)
	if err != nil {
		return err
	}
	return respond(w, data)
}

var getPageviews = handleError(getPageviewsE)
