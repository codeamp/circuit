package s3

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3 struct {
	events                 chan transistor.Event
	awsAccessKeyID         string
	awsSecretKey           string
	awsRegion              string
	awsBucket              string
	awsGeneratedUserPrefix string
	awsUserGroupName       string
}

func init() {
	transistor.RegisterPlugin("s3", func() transistor.Plugin {
		return &S3{}
	}, plugins.ProjectExtension{})
}

func (x *S3) Description() string {
	return "Provision S3 Assets for Project Use"
}

func (x *S3) SampleConfig() string {
	return ` `
}

func (x *S3) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started S3")
	return nil
}

func (x *S3) Stop() {
	log.Info("Stopping S3")
}

func (x *S3) Subscribe() []string {
	return []string{
		"project:s3:create",
		"project:s3:update",
		"project:s3:delete",
	}
}

// How does this work?
// A bucket is shared by other staging projects
// in order to create access for a new project
// to use htis storage, we will need to generate
// an IAM user to use this new prefix for this bucket
//
// We want to do this so that we can utilize the same bucket
// but isolate each applications logical access
//
// The artifacts that should be returned after an S3
// extension is successfuly created should be
// the credentials for which the project will
// be using to access the bucket, as well as the prefix
// that has been assigned for this application to use
//
// Accepts:
//		aws_access_key_id
//		aws_secret_key
//		aws_region
//		#aws_policy
// 		aws_bucket
//		aws_generated_user_prefix
//
func (x *S3) Process(e transistor.Event) error {
	var err error
	log.Debug("Processing S3 event")

	if e.Matches("project:s3") {
		switch e.Action {
		case transistor.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process S3 event: %s", e.Event()), log.Fields{})
			err = x.createS3(e)
		case transistor.GetAction("update"):
			log.InfoWithFields(fmt.Sprintf("Process S3 event: %s", e.Event()), log.Fields{})
			err = x.updateS3(e)
		case transistor.GetAction("delete"):
			log.InfoWithFields(fmt.Sprintf("Process S3 event: %s", e.Event()), log.Fields{})
			err = x.deleteS3(e)
		default:
			log.WarnWithFields(fmt.Sprintf("Unhandled S3 event: %s", e.Event()), log.Fields{})
		}

		if err != nil {
			log.Error("Sending error from process")
			x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("%v (Action: %v, Step: S3", err.Error(), e.State))
			return nil
		}
	}

	return nil
}

func (x *S3) handleCredentials(e transistor.Event) {
	log.Warn("handleCredentials")
}

func (x *S3) writeTestFile(e transistor.Event) {
	log.Warn("writeTestFile")
}

func (x *S3) sendS3Response(e transistor.Event, action transistor.Action, state transistor.State, stateMessage string, lbPayload plugins.ProjectExtension) {
	projectExtension := plugins.ProjectExtension{
		Environment: lbPayload.Environment,
		Project:     lbPayload.Project,
		ID:          lbPayload.Project.ID,
	}

	event := e.NewEvent(action, state, stateMessage)
	event.SetPayload(projectExtension)
	x.events <- event
}

func (x *S3) extractArtifacts(e transistor.Event) error {
	payload := e.Payload.(plugins.ProjectExtension)

	awsAccessKeyID, err := e.GetArtifact("aws_access_key_id")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), payload)
		return err
	}
	x.awsAccessKeyID = awsAccessKeyID.String()

	awsSecretKey, err := e.GetArtifact("aws_secret_key")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), payload)
		return err
	}
	x.awsSecretKey = awsSecretKey.String()

	awsRegion, err := e.GetArtifact("aws_region")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), payload)
		return err
	}
	x.awsRegion = awsRegion.String()

	awsBucket, err := e.GetArtifact("aws_bucket")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), payload)
		return err
	}
	x.awsBucket = awsBucket.String()

	awsGeneratedUserPrefix, err := e.GetArtifact("aws_generated_user_prefix")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), payload)
		return err
	}
	x.awsGeneratedUserPrefix = awsGeneratedUserPrefix.String()

	awsUserGroupName, err := e.GetArtifact("aws_user_group_name")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), payload)
		return err
	}
	x.awsUserGroupName = awsUserGroupName.String()

	return nil
}

func (x *S3) createS3(e transistor.Event) error {
	log.Warn("Creating S3 Extension")

	if err := x.extractArtifacts(e); err != nil {
		log.Error(err.Error())
		return err
	}

	iamSvc := iam.New(session.New(&aws.Config{
		Region:      &x.awsRegion,
		Credentials: credentials.NewStaticCredentials(x.awsAccessKeyID, x.awsSecretKey, ""),
	}))

	payload := e.Payload.(plugins.ProjectExtension)
	userName := fmt.Sprintf("%s%s", x.awsGeneratedUserPrefix, payload.Project.Slug)

	// create the user if it doesn't exist yet
	_, err := iamSvc.GetUser(&iam.GetUserInput{UserName: &userName})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchEntity") {
			// User does not exist. Create it!
			log.Error("NO SUCH ENTITY - Creating User! ", userName)

			// create the user
			_, err := iamSvc.CreateUser(&iam.CreateUserInput{UserName: &userName})
			if err != nil {
				log.Error(err.Error())
			}

			// wait for user to exist
			if err := iamSvc.WaitUntilUserExists(&iam.GetUserInput{UserName: &userName}); err != nil {
				log.Error(err.Error())
			} else {
				log.Error("user has been created! ", userName)
			}
		} else {
			log.Error(err.Error())
			return err
		}
	} else {
		log.Error("THERE WAS AN ERROR - USER EXISTS")
		return errors.New("There was an error during setup")
	}

	// Add the user to the group
	if _, err := iamSvc.AddUserToGroup(&iam.AddUserToGroupInput{
		UserName:  &userName,
		GroupName: &x.awsUserGroupName,
	}); err != nil {
		log.Error(err.Error())
		return err
	} else {
		log.Error("User was added to group: ", x.awsUserGroupName)
	}

	// Assign the user a policy that includes the bucket and prefix
	userPolicyTemplate := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
            	"s3:GetObject*",
            	"s3:PutObject*",
            	"s3:DeleteObject*",
            	"s3:List*",
            	"s3:AbortMultipartUpload"
             ],
            "Resource": [
                "arn:aws:s3:::%s/%s/*"
            ]
        }
    ]
	}`

	userPolicy := fmt.Sprintf(userPolicyTemplate, x.awsBucket, payload.Project.Slug)
	_, err = iamSvc.PutUserPolicy(&iam.PutUserPolicyInput{
		UserName:       &userName,
		PolicyName:     aws.String("codeamp-storage"),
		PolicyDocument: &userPolicy,
	})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	// Generate an API key for the user
	accessKeyResponse, err := iamSvc.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: &userName,
	})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	time.Sleep(time.Second * 10)

	// Test that the API key has write access to the bucket
	testSession := session.New(&aws.Config{
		Region:      &x.awsRegion,
		Credentials: credentials.NewStaticCredentials(*accessKeyResponse.AccessKey.AccessKeyId, *accessKeyResponse.AccessKey.SecretAccessKey, ""),
	})

	testS3Svc := s3.New(testSession)
	input := &s3.PutObjectInput{
		Body:   strings.NewReader("This is a test file written by CodeAmp. You may delete it."),
		Bucket: &x.awsBucket,
		Key:    aws.String(fmt.Sprintf("%s/%s", payload.Project.Slug, "codeamp-write-test.txt")),
	}

	_, err = testS3Svc.PutObject(input)
	if err != nil {
		log.Error(err.Error())
		return errors.New("There was an error writing with S3")
	}

	log.Error("S3 Success!")
	ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), fmt.Sprintf("S3 message"))
	x.events <- ev

	x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("complete"), "S3 Completed", payload)

	return nil
}

func (x *S3) updateS3(e transistor.Event) error {
	log.Warn("Updating S3 Extension")

	log.Error("S3 Success!")
	ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), fmt.Sprintf("S3 message"))
	x.events <- ev

	return nil
}

func (x *S3) deleteS3(e transistor.Event) error {
	if err := x.extractArtifacts(e); err != nil {
		log.Error(err.Error())
		return err
	}

	iamSvc := iam.New(session.New(&aws.Config{
		Region:      &x.awsRegion,
		Credentials: credentials.NewStaticCredentials(x.awsAccessKeyID, x.awsSecretKey, ""),
	}))

	payload := e.Payload.(plugins.ProjectExtension)
	userName := fmt.Sprintf("%s%s", x.awsGeneratedUserPrefix, payload.Project.Slug)

	log.Warn("DELETING USER: ", userName)

	listAccessKeys, err := iamSvc.ListAccessKeys(&iam.ListAccessKeysInput{UserName: &userName})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	for _, accessKey := range listAccessKeys.AccessKeyMetadata {
		_, err := iamSvc.DeleteAccessKey(&iam.DeleteAccessKeyInput{UserName: &userName, AccessKeyId: accessKey.AccessKeyId})
		if err != nil {
			log.Error(err.Error())
		}
	}

	_, err = iamSvc.RemoveUserFromGroup(&iam.RemoveUserFromGroupInput{UserName: &userName, GroupName: &x.awsUserGroupName})
	if err != nil {
		log.Error(err.Error())
	}

	_, err = iamSvc.DeleteUserPolicy(&iam.DeleteUserPolicyInput{UserName: &userName, PolicyName: aws.String("codeamp-storage")})
	if err != nil {
		log.Error(err.Error())
	}

	_, err = iamSvc.DeleteUser(&iam.DeleteUserInput{UserName: &userName})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}
