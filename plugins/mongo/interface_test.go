package mongo_test

import (
	"context"
	"errors"

	atlas_models "github.com/Clever/atlas-api-client/gen-go/models"
	"github.com/codeamp/circuit/plugins/mongo"
)

type MockMongoAtlasClientBuilder struct {
	MockMongoAtlasClient
}

func (x *MockMongoAtlasClientBuilder) New(apiEndpoint string, publicKey string, privateKey string) mongo.MongoAtlasClient {
	return &x.MockMongoAtlasClient
}

type MockMongoAtlasClient struct {
	Users []*atlas_models.DatabaseUser
}

func (x *MockMongoAtlasClient) Clear() {
	x.Users = make([]*atlas_models.DatabaseUser, 0, 10)
}

func (x *MockMongoAtlasClient) GetDatabaseUsers(context.Context, string) (*atlas_models.GetDatabaseUsersResponse, error) {
	return nil, errors.New("Stub Function!")
}

func (x *MockMongoAtlasClient) GetDatabaseUser(context.Context, *atlas_models.GetDatabaseUserInput) (*atlas_models.DatabaseUser, error) {
	return nil, errors.New("Stub Function!")
}

func (x *MockMongoAtlasClient) CreateDatabaseUser(context.Context, *atlas_models.CreateDatabaseUserInput) (*atlas_models.DatabaseUser, error) {
	return nil, errors.New("Stub Function!")
}

func (x *MockMongoAtlasClient) DeleteDatabaseUser(context.Context, *atlas_models.DeleteDatabaseUserInput) error {
	return errors.New("Stub Function!")
}
