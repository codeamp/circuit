package mongo_test

type MockMongo struct {
	mongo.MongoAPI
}

func (x *MockMongo) GetMongoInterface() mongo.MongoAPI {
	return nil
}
