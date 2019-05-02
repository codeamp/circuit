package mongo

import atlas "github.com/Clever/atlas-api-client/gen-go/client"

type MongoAPI interface {
}

type Mongo struct {
	MongoAPI
	AtlasClient
}

func (x *Mongo) GetMongoInterface() MongoAPI {
	return nil
}

type AtlasClient interface {
	New(string) atlas.Client
}

type Atlas struct {
}

func (x *Mongo) New(apiEndpoint string) atlas.Client {
	return atlas.New(apiEndpoint)
}
