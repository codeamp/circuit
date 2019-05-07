package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3API interface {
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

type IAMAPI interface {
	AddUserToGroup(*iam.AddUserToGroupInput) (*iam.AddUserToGroupOutput, error)
	RemoveUserFromGroup(*iam.RemoveUserFromGroupInput) (*iam.RemoveUserFromGroupOutput, error)

	CreateAccessKey(*iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error)
	ListAccessKeys(*iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error)
	DeleteAccessKey(*iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error)

	CreateUser(*iam.CreateUserInput) (*iam.CreateUserOutput, error)
	GetUser(*iam.GetUserInput) (*iam.GetUserOutput, error)
	DeleteUser(*iam.DeleteUserInput) (*iam.DeleteUserOutput, error)
	WaitUntilUserExists(*iam.GetUserInput) error

	DeleteUserPolicy(*iam.DeleteUserPolicyInput) (*iam.DeleteUserPolicyOutput, error)
	PutUserPolicy(*iam.PutUserPolicyInput) (*iam.PutUserPolicyOutput, error)
}

type S3Interfacer interface {
	New() S3Interfacer
	GetIAMServiceInterface(*S3Data) IAMAPI
	GetS3ServiceInterface(*S3Data, *iam.AccessKey) S3API
}

type S3Interface struct {
	IAMSvc IAMAPI
	S3Svc  S3API
}

func (x *S3Interface) New() S3Interfacer {
	return x
}

// Provide IAM interface for mock/testing purposes
func (x *S3Interface) GetIAMServiceInterface(data *S3Data) IAMAPI {
	x.IAMSvc = iam.New(session.New(&aws.Config{
		Region:      &data.AWSRegion,
		Credentials: credentials.NewStaticCredentials(data.AWSAccessKeyID, data.AWSSecretKey, ""),
	}))

	return x.IAMSvc
}

// Provide S3 interface for mock/testing purposes
func (x *S3Interface) GetS3ServiceInterface(data *S3Data, testingAccessKey *iam.AccessKey) S3API {
	x.S3Svc = s3.New(session.New(&aws.Config{
		Region:      &data.AWSRegion,
		Credentials: credentials.NewStaticCredentials(*testingAccessKey.AccessKeyId, *testingAccessKey.SecretAccessKey, ""),
	}))

	return x.S3Svc
}
