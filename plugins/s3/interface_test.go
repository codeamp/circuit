package s3_test

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	amzS3 "github.com/aws/aws-sdk-go/service/s3"

	"github.com/codeamp/circuit/plugins/s3"
	log "github.com/codeamp/logger"
)

type MockS3Interface struct {
	IAMSvc MockIAMClient
	S3Svc  MockS3Client
}

func (x *MockS3Interface) New() *MockS3Interface {
	x.IAMSvc = *x.IAMSvc.New()
	return x
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
	accessKeys   map[string][]*iam.AccessKeyMetadata
	users        map[string]*iam.User
	userPolicies map[string]map[string]*string
	groupMembers map[string]map[string]string
}

func (x *MockIAMClient) New() *MockIAMClient {
	x.users = make(map[string]*iam.User)
	x.userPolicies = make(map[string]map[string]*string)
	x.accessKeys = make(map[string][]*iam.AccessKeyMetadata)
	x.groupMembers = make(map[string]map[string]string)
	return x
}

func (x *MockIAMClient) AddUserToGroup(input *iam.AddUserToGroupInput) (*iam.AddUserToGroupOutput, error) {
	if _, ok := x.users[*input.UserName]; ok != false {
		// Assume the group exists for testing purposes
		if _, ok := x.groupMembers[*input.GroupName]; ok == false {
			x.groupMembers[*input.GroupName] = make(map[string]string, 0)
		}

		x.groupMembers[*input.GroupName][*input.UserName] = *input.UserName
		return &iam.AddUserToGroupOutput{}, nil
	}

	return nil, errors.New("NoSuchEntity - AddUserToGroup")
}

func (x *MockIAMClient) RemoveUserFromGroup(input *iam.RemoveUserFromGroupInput) (*iam.RemoveUserFromGroupOutput, error) {
	return nil, errors.New("RemoveUserFromGroup Stub Function: Implement!")
}

func (x *MockIAMClient) CreateAccessKey(input *iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error) {
	if _, ok := x.users[*input.UserName]; ok != false {
		if _, ok := x.accessKeys[*input.UserName]; ok == false {
			x.accessKeys[*input.UserName] = make([]*iam.AccessKeyMetadata, 0, 1)
		}

		createDate := time.Now()
		x.accessKeys[*input.UserName] = append(x.accessKeys[*input.UserName], &iam.AccessKeyMetadata{
			AccessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE"),
			UserName:    input.UserName,
			Status:      aws.String("Active"),
			CreateDate:  &createDate,
		})

		return &iam.CreateAccessKeyOutput{
			AccessKey: &iam.AccessKey{
				AccessKeyId:     aws.String("AKIAIOSFODNN7EXAMPLE"),
				CreateDate:      &createDate,
				SecretAccessKey: aws.String("AKIAIOSFODNN7EXAMPLE"),
				Status:          aws.String("Active"),
				UserName:        input.UserName,
			},
		}, nil
	}
	return nil, errors.New("NoSuchEntity - CreateAccessKey")
}

func (x *MockIAMClient) ListAccessKeys(input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {
	if metadata, ok := x.accessKeys[*input.UserName]; ok != false {
		isTruncated := false
		return &iam.ListAccessKeysOutput{
			AccessKeyMetadata: metadata,
			IsTruncated:       &isTruncated,
		}, nil
	}

	return nil, errors.New("NoSuchEntity - ListAccessKeys")
}

func (x *MockIAMClient) DeleteAccessKey(input *iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error) {
	if _, ok := x.users[*input.UserName]; ok != false {
		return nil, nil
	}

	return nil, errors.New("NoSuchEntity - DeleteAccessKey")
}

func (x *MockIAMClient) CreateUser(input *iam.CreateUserInput) (*iam.CreateUserOutput, error) {
	log.Warn("CreateUser")
	if _, ok := x.users[*input.UserName]; ok != true {
		x.users[*input.UserName] = &iam.User{
			UserName: input.UserName,
		}

		return &iam.CreateUserOutput{User: x.users[*input.UserName]}, nil
	} else {
		// There was a user here already
		return &iam.CreateUserOutput{User: x.users[*input.UserName]}, errors.New("EntityAlreadyExists - CreateUser")
	}

	return nil, errors.New("InvalidInput")
}

func (x *MockIAMClient) GetUser(input *iam.GetUserInput) (*iam.GetUserOutput, error) {
	if user, ok := x.users[*input.UserName]; ok != false {
		return &iam.GetUserOutput{User: user}, nil
	}

	return nil, errors.New("NoSuchEntity - GetUser")
}
func (x *MockIAMClient) DeleteUser(input *iam.DeleteUserInput) (*iam.DeleteUserOutput, error) {
	return nil, errors.New("DeleteUser Stub Function: Implement!")
}
func (x *MockIAMClient) WaitUntilUserExists(input *iam.GetUserInput) error {
	return nil
}

func (x *MockIAMClient) DeleteUserPolicy(input *iam.DeleteUserPolicyInput) (*iam.DeleteUserPolicyOutput, error) {
	if _, ok := x.users[*input.UserName]; ok == false {
		return nil, errors.New("NoSuchEntity - DeleteUserPolicy")
	}

	return nil, errors.New("NoSuchEntity - DeleteUserPolicy")
}
func (x *MockIAMClient) PutUserPolicy(input *iam.PutUserPolicyInput) (*iam.PutUserPolicyOutput, error) {
	if _, ok := x.users[*input.UserName]; ok == false {
		return nil, errors.New("NoSuchEntity - PutUserPolicy")
	}

	if _, ok := x.userPolicies[*input.UserName]; ok == false {
		x.userPolicies[*input.UserName] = make(map[string]*string)
	}

	x.userPolicies[*input.UserName][*input.PolicyName] = input.PolicyDocument
	return &iam.PutUserPolicyOutput{}, nil
}

type MockS3Client struct {
}

func (x *MockS3Client) PutObject(input *amzS3.PutObjectInput) (*amzS3.PutObjectOutput, error) {
	return &amzS3.PutObjectOutput{
		ETag: aws.String("EASLDKJASLDKJASD"),
	}, nil
}
