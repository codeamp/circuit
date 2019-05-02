package mongo

import (
	"github.com/Clever/atlas-api-client/digestauth"
	atlas "github.com/Clever/atlas-api-client/gen-go/client"
)

type MongoAtlasClientBuilder interface {
	New(string, string, string) atlas.Client
}

type MongoAPI interface {
}

type Mongo struct {
	MongoAPI
	MongoAtlasClient
}

func (x *Mongo) GetMongoInterface() MongoAPI {
	return nil
}

type MongoAtlasClient struct {
}

func (x *MongoAtlasClient) New(apiEndpoint string, publicKey string, privateKey string) atlas.Client {
	atlasAPI := atlas.New(apiEndpoint)
	digestT := digestauth.NewTransport(
		publicKey,
		privateKey,
	)
	atlasAPI.SetTransport(&digestT)

	return atlasAPI
}
