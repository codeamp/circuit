package mongo_test

import "github.com/codeamp/circuit/plugins/mongo"

type MockMongoAPI struct {
	MongoAPI mongo.MongoAPI
}

type MockMongoAtlasAPI struct {
	MongoAtlasAPI mongo.MongoAtlasAPI
}

func (x *MockMongo) GetMongoInterface() mongo.MongoAPI {
	return nil
}
