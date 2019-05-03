package mongo

import (
	"context"

	"github.com/Clever/atlas-api-client/digestauth"
	atlas "github.com/Clever/atlas-api-client/gen-go/client"
	atlas_models "github.com/Clever/atlas-api-client/gen-go/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoClientNamespacer interface {
	NewClient(...*options.ClientOptions) (MongoClienter, error)
}

type MongoClienter interface {
	Connect(context.Context) error
	Ping(context.Context, *readpref.ReadPref) error
	ListDatabaseNames(context.Context, interface{}, ...*options.ListDatabasesOptions) ([]string, error)
}

type MongoAtlasClienter interface {
	GetDatabaseUsers(context.Context, string) (*atlas_models.GetDatabaseUsersResponse, error)
	GetDatabaseUser(context.Context, *atlas_models.GetDatabaseUserInput) (*atlas_models.DatabaseUser, error)

	CreateDatabaseUser(context.Context, *atlas_models.CreateDatabaseUserInput) (*atlas_models.DatabaseUser, error)
	DeleteDatabaseUser(context.Context, *atlas_models.DeleteDatabaseUserInput) error
}

type MongoAtlasClientNamespacer interface {
	New(string, string, string) MongoAtlasClienter
}

type MongoAtlasClientNamespace struct {
}

func (x *MongoAtlasClientNamespace) New(apiEndpoint string, publicKey string, privateKey string) MongoAtlasClienter {
	atlasAPI := atlas.New(apiEndpoint)
	digestT := digestauth.NewTransport(
		publicKey,
		privateKey,
	)
	atlasAPI.SetTransport(&digestT)

	return atlasAPI
}

type MongoClientNamespace struct {
}

func (x *MongoClientNamespace) NewClient(opts ...*options.ClientOptions) (MongoClienter, error) {
	return mongo.NewClient(opts...)
}
