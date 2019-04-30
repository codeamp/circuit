package mongo

type MongoAPI interface {
}

type Mongoer interface {
}

type Mongo struct {
	MongoAPI
}

func (x *Mongo) GetMongoInterface() MongoAPI {
	return nil
}
