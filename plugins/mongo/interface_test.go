package mongo_test

import (
	"context"
	"errors"

	atlas_models "github.com/Clever/atlas-api-client/gen-go/models"
	"github.com/codeamp/circuit/plugins/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MockMongoAtlasClientNamespace struct {
	MockMongoAtlasClient
}

func (x *MockMongoAtlasClientNamespace) New(apiEndpoint string, publicKey string, privateKey string) mongo.MongoAtlasClienter {
	return &x.MockMongoAtlasClient
}

type MockMongoAtlasClient struct {
	Users []*atlas_models.DatabaseUser
}

func (x *MockMongoAtlasClient) Clear() {
	x.Users = make([]*atlas_models.DatabaseUser, 0, 10)
}

func (x *MockMongoAtlasClient) GetDatabaseUsers(context.Context, string) (*atlas_models.GetDatabaseUsersResponse, error) {
	response := &atlas_models.GetDatabaseUsersResponse{
		Results:    x.Users,
		TotalCount: int64(len(x.Users)),
	}

	return response, nil
}

func (x *MockMongoAtlasClient) GetDatabaseUser(ctx context.Context, input *atlas_models.GetDatabaseUserInput) (*atlas_models.DatabaseUser, error) {
	for _, dbUser := range x.Users {
		if dbUser.GroupID == input.GroupID {
			if dbUser.Username == input.Username {
				return dbUser, nil
			}
		}
	}

	return nil, errors.New("No user with username")
}

func (x *MockMongoAtlasClient) CreateDatabaseUser(ctx context.Context, input *atlas_models.CreateDatabaseUserInput) (*atlas_models.DatabaseUser, error) {

	additionalUser := &atlas_models.DatabaseUser{
		DatabaseName: input.CreateDatabaseUserRequest.DatabaseName,
		GroupID:      input.GroupID,
		Links:        nil,
		Roles:        nil,
		Username:     input.CreateDatabaseUserRequest.Username,
	}

	x.Users = append(x.Users, additionalUser)

	return additionalUser, nil
}

func (x *MockMongoAtlasClient) DeleteDatabaseUser(ctx context.Context, input *atlas_models.DeleteDatabaseUserInput) error {
	for idx, dbUser := range x.Users {
		if dbUser.GroupID == input.GroupID {
			if dbUser.Username == input.Username {
				x.Users = append(x.Users[:idx], x.Users[idx+1:]...)
				return nil
			}
		}
	}

	return errors.New("No user with username")
}

type MockMongoClientNamespace struct {
}

func (x *MockMongoClientNamespace) NewClient(opts ...*options.ClientOptions) (mongo.MongoClienter, error) {
	return &MockMongoClient{}, nil
}

type MockMongoClient struct {
}

func (x *MockMongoClient) Connect(context context.Context) error {
	return nil
}

func (x *MockMongoClient) Ping(context context.Context, readpref *readpref.ReadPref) error {
	return nil
}

func (x *MockMongoClient) ListDatabaseNames(context context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) ([]string, error) {
	return nil, nil
}
