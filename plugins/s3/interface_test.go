package s3_test

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	amzS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/davecgh/go-spew/spew"

	"github.com/codeamp/circuit/plugins/s3"
)

type MockS3Interface struct {
	IAMSvc MockIAMClient
	S3Svc  MockS3Client
}

func (x *MockS3Interface) New() s3.S3Interfacer {
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
	if _, ok := x.users[*input.UserName]; ok == true {
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
	if _, ok := x.users[*input.UserName]; ok == false {
		return nil, errors.New("NoSuchEntity - RemoveUserFromGroup")
	}

	if _, ok := x.groupMembers[*input.GroupName]; ok == false {
		return nil, errors.New("NoSuchEntity - RemoveUserFromGroup")
	}

	if _, ok := x.groupMembers[*input.GroupName][*input.UserName]; ok == false {
		return nil, errors.New("NoSuchEntity - RemoveUserFromGroup")
	}

	delete(x.groupMembers[*input.GroupName], *input.UserName)
	return &iam.RemoveUserFromGroupOutput{}, nil
}

func (x *MockIAMClient) CreateAccessKey(input *iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error) {
	if _, ok := x.users[*input.UserName]; ok == true {
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
	if _, ok := x.users[*input.UserName]; ok != true {
		return nil, errors.New("NoSuchEntity")
	}

	if metadata, ok := x.accessKeys[*input.UserName]; ok == true {
		isTruncated := false
		return &iam.ListAccessKeysOutput{
			AccessKeyMetadata: metadata,
			IsTruncated:       &isTruncated,
		}, nil
	}

	return &iam.ListAccessKeysOutput{
		AccessKeyMetadata: []*iam.AccessKeyMetadata{},
	}, nil
}

func (x *MockIAMClient) DeleteAccessKey(input *iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error) {
	if _, ok := x.users[*input.UserName]; ok != true {
		return nil, errors.New("NoSuchEntity")
	}

	if _, ok := x.accessKeys[*input.UserName]; ok == true {
		delete(x.accessKeys, *input.UserName)
	}

	spew.Dump(x.accessKeys)
	return &iam.DeleteAccessKeyOutput{}, nil
}

func (x *MockIAMClient) CreateUser(input *iam.CreateUserInput) (*iam.CreateUserOutput, error) {
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
	if user, ok := x.users[*input.UserName]; ok == true {
		return &iam.GetUserOutput{User: user}, nil
	}

	return nil, errors.New("NoSuchEntity - GetUser")
}
func (x *MockIAMClient) DeleteUser(input *iam.DeleteUserInput) (*iam.DeleteUserOutput, error) {
	// Check to see if user stuff is deleted
	if _, ok := x.userPolicies[*input.UserName]; ok == true {
		if len(x.userPolicies[*input.UserName]) > 0 {
			return nil, errors.New("DeleteConflict - UserPolicies")
		}
	}

	if _, ok := x.accessKeys[*input.UserName]; ok == true {
		spew.Dump(x.accessKeys)
		return nil, errors.New("DeleteConflict - AccessKeys")
	}

	for idx, group := range x.groupMembers {
		spew.Dump(group, idx)
		for idx, member := range group {
			spew.Dump(member, idx)
		}
	}

	if _, ok := x.users[*input.UserName]; ok != true {
		return nil, errors.New("NoSuchEntity")
	}

	delete(x.users, *input.UserName)
	return &iam.DeleteUserOutput{}, nil
}
func (x *MockIAMClient) WaitUntilUserExists(input *iam.GetUserInput) error {
	return nil
}

func (x *MockIAMClient) DeleteUserPolicy(input *iam.DeleteUserPolicyInput) (*iam.DeleteUserPolicyOutput, error) {
	if _, ok := x.users[*input.UserName]; ok == false {
		return nil, errors.New("NoSuchEntity - DeleteUserPolicy")
	}

	if _, ok := x.userPolicies[*input.UserName]; ok == false {
		return nil, errors.New("NoSuchEntity - DeleteUserPolicy")
	}

	if _, ok := x.userPolicies[*input.UserName][*input.PolicyName]; ok == false {
		return nil, errors.New("NoSuchEntity - DeleteUserPolicy")
	}

	delete(x.userPolicies[*input.UserName], *input.PolicyName)

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
