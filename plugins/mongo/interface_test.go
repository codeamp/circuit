package mongo_test

import (
	atlas "github.com/Clever/atlas-api-client/gen-go/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type MockMongoAtlasClient struct {
	suite.Suite
}

func (x *MockMongoAtlasClient) New(apiEndpoint string, publicKey string, privateKey string) atlas.Client {
	controller := gomock.NewController(x.T())
	atlasAPI := atlas.NewMockClient(controller)

	return atlasAPI
}
