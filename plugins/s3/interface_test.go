package s3_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/iam"
	amzS3 "github.com/aws/aws-sdk-go/service/s3"

	"github.com/codeamp/circuit/plugins/s3"
)

type MockS3Interface struct {
	IAMSvc MockIAMClient
	S3Svc  MockS3Client
}

// Provide IAM interface for mock/testing purposes
func (x *MockS3Interface) GetIAMServiceInterface(data *s3.S3Data) s3.IAMAPI {
	return &x.IAMSvc
}

// Provide S3 interface for mock/testing purposes
func (x *MockS3Interface) GetS3ServiceInterface(data *s3.S3Data, testingAccessKey *iam.AccessKey) s3.S3API {
	return &x.S3Svc
}

type MockIAMClient struct {
}

func (x *MockIAMClient) AddUserToGroup(*iam.AddUserToGroupInput) (*iam.AddUserToGroupOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}
func (x *MockIAMClient) RemoveUserFromGroup(*iam.RemoveUserFromGroupInput) (*iam.RemoveUserFromGroupOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}

func (x *MockIAMClient) CreateAccessKey(*iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}
func (x *MockIAMClient) ListAccessKeys(*iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}
func (x *MockIAMClient) DeleteAccessKey(*iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}

func (x *MockIAMClient) CreateUser(*iam.CreateUserInput) (*iam.CreateUserOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}
func (x *MockIAMClient) GetUser(*iam.GetUserInput) (*iam.GetUserOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}
func (x *MockIAMClient) DeleteUser(*iam.DeleteUserInput) (*iam.DeleteUserOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}
func (x *MockIAMClient) WaitUntilUserExists(*iam.GetUserInput) error {
	return errors.New("Stub Function: Implement!")
}

func (x *MockIAMClient) DeleteUserPolicy(*iam.DeleteUserPolicyInput) (*iam.DeleteUserPolicyOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}
func (x *MockIAMClient) PutUserPolicy(*iam.PutUserPolicyInput) (*iam.PutUserPolicyOutput, error) {
	return nil, errors.New("Stub Function: Implement!")
}

type MockS3Client struct {
}

func (x *MockS3Client) PutObject(input *amzS3.PutObjectInput) (*amzS3.PutObjectOutput, error) {
	return nil, errors.New("Put Object Failed!")
}
