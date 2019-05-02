package mongo

import (
	"context"

	"github.com/Clever/atlas-api-client/digestauth"
	atlas "github.com/Clever/atlas-api-client/gen-go/client"
	atlas_models "github.com/Clever/atlas-api-client/gen-go/models"
)

type MongoAtlasClient interface {
	GetDatabaseUsers(context.Context, string) (*atlas_models.GetDatabaseUsersResponse, error)
	GetDatabaseUser(context.Context, *atlas_models.GetDatabaseUserInput) (*atlas_models.DatabaseUser, error)

	CreateDatabaseUser(context.Context, *atlas_models.CreateDatabaseUserInput) (*atlas_models.DatabaseUser, error)
	DeleteDatabaseUser(context.Context, *atlas_models.DeleteDatabaseUserInput) error
}

type MongoAtlasClientBuilder interface {
	New(string, string, string) MongoAtlasClient
}

type mongoAtlasClient struct {
}

func (x *mongoAtlasClient) New(apiEndpoint string, publicKey string, privateKey string) MongoAtlasClient {
	atlasAPI := atlas.New(apiEndpoint)
	digestT := digestauth.NewTransport(
		publicKey,
		privateKey,
	)
	atlasAPI.SetTransport(&digestT)

	return atlasAPI
}
