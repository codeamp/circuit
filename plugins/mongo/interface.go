package mongo

type MongoAPI interface {
}

type MongoAtlasAPI interface {
	New(string) MongoAtlasAPI
}

type Mongo struct {
	MongoAPI
	MongoAtlasAPI
}

func (x *Mongo) GetMongoInterface() MongoAPI {
	return nil
}

func (x *Mongo) GetMongoAtlasInterface() MongoAtlasAPI {
	return nil
}
