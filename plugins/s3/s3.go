package s3

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/aws/aws-sdk-go/service/s3"
)

const s3UserPolicyTemplate = `{
	"Version": "2012-10-17",
	"Statement": [
	    {
	        "Effect": "Allow",
	        "Action": [
	        	"s3:GetObject*",
	        	"s3:PutObject*",
	        	"s3:DeleteObject*",
	        	"s3:ListObject*",
	        	"s3:AbortMultipartUpload"
	         ],
	        "Resource": [
	            "arn:aws:s3:::%s/%s/*"
	        ]
	    },
	    {
	        "Effect": "Allow",
	        "Action": [
	        	"s3:List*"
	         ],
	        "Resource": [
	            "arn:aws:s3:::%s"
	        ]
	    }
	]
}`

type S3 struct {
	events chan transistor.Event
	data   S3Data

	S3Interfaces S3Interfacer
}

type S3Data struct {
	AWSAccessKeyID         string
	AWSSecretKey           string
	AWSRegion              string
	AWSBucket              string
	AWSGeneratedUserPrefix string
	AWSUserGroupName       string
	AWSCredentialsTimeout  int
}

func init() {
	transistor.RegisterPlugin("s3", func() transistor.Plugin {
		return &S3{S3Interfaces: &S3Interface{}}
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
// to use this storage, we will need to generate
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
// in addition to the region the bucket is in
//
// Accepts:
//		aws_access_key_id			- Access to create users and update policies
//		aws_secret_key
//		aws_region					- The region of the bucket
// 		aws_bucket					- Which bucket to be shared
//		aws_generated_user_prefix	- What should the IAM users be prefixed with
//		aws_user_group_name			- For organizational purposes, group users together to easily find later
// 		aws_credentials_timeout		- How long should we wait to see if the IAM credentials were successfully created
//
func (x *S3) Process(e transistor.Event) error {
	var err error
	if e.Matches("project:s3") {
		log.InfoWithFields(fmt.Sprintf("Process S3 event: %s", e.Event()), log.Fields{})
		switch e.Action {
		case transistor.GetAction("create"):
			err = x.createS3(e)
		case transistor.GetAction("update"):
			err = x.updateS3(e)
		case transistor.GetAction("delete"):
			err = x.deleteS3(e)
		default:
			log.Warn(fmt.Sprintf("Unhandled S3 event: %s", e.Event()))
		}

		if err != nil {
			log.Error("Sending error from process")
			x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("%v (Action: %v, Step: S3)", err.Error(), e.State))
			return nil
		}
	}

	return nil
}

// Wraper for sending an event back thruogh the messaging system for standardization and brevity
func (x *S3) sendS3Response(e transistor.Event, action transistor.Action, state transistor.State, stateMessage string, artifacts []transistor.Artifact) {
	event := e.NewEvent(action, state, stateMessage)
	event.Artifacts = artifacts

	x.events <- event
}

// Pull all the artifacts out from the event that we will need
// in order to service these requests. Stuff them into a local storage object.
func (x *S3) extractArtifacts(e transistor.Event) (*S3Data, error) {
	var data S3Data

	// Access Key ID
	awsAccessKeyID, err := e.GetArtifact("aws_access_key_id")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.AWSAccessKeyID = awsAccessKeyID.String()

	// Secret Key
	awsSecretKey, err := e.GetArtifact("aws_secret_key")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.AWSSecretKey = awsSecretKey.String()

	// Region
	awsRegion, err := e.GetArtifact("aws_region")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.AWSRegion = awsRegion.String()

	// Bucket
	awsBucket, err := e.GetArtifact("aws_bucket")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.AWSBucket = awsBucket.String()

	// Generated User Prefix
	awsGeneratedUserPrefix, err := e.GetArtifact("aws_generated_user_prefix")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.AWSGeneratedUserPrefix = awsGeneratedUserPrefix.String()

	// User Group Name
	awsUserGroupName, err := e.GetArtifact("aws_user_group_name")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.AWSUserGroupName = awsUserGroupName.String()

	// Credentials Timeout
	awsCredentialsTimeout, err := e.GetArtifact("aws_credentials_timeout")
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	credentialsTimeout, err := strconv.Atoi(awsCredentialsTimeout.String())
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
	}
	data.AWSCredentialsTimeout = credentialsTimeout

	return &data, nil
}

// Creates an IAM user with the given userName if one does not currently exist
func (x *S3) createIAMUserIfNotExist(data *S3Data, userName string) error {
	iamSvc := x.S3Interfaces.GetIAMServiceInterface(data)

	// create the user if it doesn't exist yet
	_, err := iamSvc.GetUser(&iam.GetUserInput{UserName: &userName})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchEntity") {
			// User does not exist. Create it!
			log.Info("NO SUCH ENTITY - Creating User! ", userName)

			// create the user
			_, err := iamSvc.CreateUser(&iam.CreateUserInput{UserName: &userName})
			if err != nil {
				log.Error(err.Error())
			}

			// wait for user to exist
			if err := iamSvc.WaitUntilUserExists(&iam.GetUserInput{UserName: &userName}); err != nil {
				log.Error(err.Error())
			} else {
				log.Info("user has been created! ", userName)
			}
		} else {
			log.Error(err.Error())
			return err
		}
	} else {
		log.Error("S3 IAM USER EXISTS")
		return errors.New("User Exists - Duplicate Extension?")
	}

	return nil
}

// Adds a IAM user to a specified IAM group
func (x *S3) addIAMUserToGroup(data *S3Data, userName string) error {
	iamSvc := x.S3Interfaces.GetIAMServiceInterface(data)

	if _, err := iamSvc.AddUserToGroup(&iam.AddUserToGroupInput{
		UserName:  &userName,
		GroupName: &data.AWSUserGroupName,
	}); err != nil {
		log.Error(err.Error())
		return err
	}

	log.Info(fmt.Sprintf("User '%s' was added to group '%s'", userName, data.AWSUserGroupName))
	return nil
}

// Assign the below policy
func (x *S3) assignUserIAMPolicyForS3(data *S3Data, userName string, prefix string) error {
	iamSvc := x.S3Interfaces.GetIAMServiceInterface(data)

	userPolicy := fmt.Sprintf(s3UserPolicyTemplate, data.AWSBucket, prefix, data.AWSBucket)
	_, err := iamSvc.PutUserPolicy(&iam.PutUserPolicyInput{
		UserName:       &userName,
		PolicyName:     aws.String("codeamp-storage"),
		PolicyDocument: &userPolicy,
	})

	if err != nil {
		log.Error(err.Error())
	}

	return err
}

// Verify that the credentials that we have generated will allow the user to do
// the basic object create operation.
//
// Further sanity checks could include the ability to get and delete this file afterwards
// but a simple PutObject request should suffice for now
func (x *S3) verifyS3CredentialsValid(e transistor.Event, data *S3Data, userName string, accessKey *iam.AccessKey, prefix string) error {

	testS3Svc := x.S3Interfaces.GetS3ServiceInterface(data, accessKey)
	input := &s3.PutObjectInput{
		Body:   strings.NewReader("This is a test file written by CodeAmp. You may delete it."),
		Bucket: &data.AWSBucket,
		Key:    aws.String(fmt.Sprintf("%s/%s", prefix, "codeamp-write-test.txt")),
	}

	startedTime := time.Now()
	currentTime := time.Now()
	waitInterval := 5
	for true {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("running"), fmt.Sprintf("Testing Permissions by Writing Sample File (%v elapsed)", currentTime.Sub(startedTime)), nil)

		_, err := testS3Svc.PutObject(input)
		if err != nil {
			log.Warn(err.Error())
		} else {
			break
		}

		time.Sleep(time.Duration(waitInterval) * time.Second)
		currentTime = time.Now()

		if currentTime.Sub(startedTime) >= (time.Duration(data.AWSCredentialsTimeout) * time.Second) {
			x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("Timed out when verifying permissions (%ds)", data.AWSCredentialsTimeout), nil)
			return errors.New("Timed out when verifying permissions")
		}
	}

	return nil
}

// Generate an IAM access key from an existing IAM user
func (x *S3) generateAccessCredentials(data *S3Data, userName string) (*iam.AccessKey, error) {
	iamSvc := x.S3Interfaces.GetIAMServiceInterface(data)

	accessKeyResponse, err := iamSvc.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: &userName,
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return accessKeyResponse.AccessKey, nil
}

// Prepare artifacts array to pass important information back to the front-end
func (x *S3) buildResultArtifacts(data *S3Data, prefix string, accessKey *iam.AccessKey) []transistor.Artifact {
	// Stuff new credentials into artifacts to be used by the front-end
	return []transistor.Artifact{
		transistor.Artifact{Key: "s3_aws_access_key_id", Value: accessKey.AccessKeyId, Secret: false},
		transistor.Artifact{Key: "s3_aws_secret_key", Value: accessKey.SecretAccessKey, Secret: false},
		transistor.Artifact{Key: "s3_aws_region", Value: data.AWSRegion, Secret: false},
		transistor.Artifact{Key: "s3_aws_bucket", Value: data.AWSBucket, Secret: false},
		transistor.Artifact{Key: "s3_aws_prefix", Value: fmt.Sprintf("%s/", prefix), Secret: false},
	}
}

// Handle the 'create' message when received
// Provision storage for a project. Return an error if there is a user that already exists.
func (x *S3) createS3(e transistor.Event) error {
	var data *S3Data
	var err error

	// Pull the required data from the event's artifacts
	if data, err = x.extractArtifacts(e); err != nil {
		log.Error(err.Error())
		return err
	}

	// Report process has begun
	x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("running"), "Creating S3 Configuration", nil)

	payload := e.Payload.(plugins.ProjectExtension)
	userName := fmt.Sprintf("%s%s", data.AWSGeneratedUserPrefix, payload.Project.Slug)

	// Creates the user if they do not exist
	err = x.createIAMUserIfNotExist(data, userName)
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return err
	}

	// Add the user to the group
	err = x.addIAMUserToGroup(data, userName)
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return err
	}

	// Generate credentials
	accessCredentials, err := x.generateAccessCredentials(data, userName)
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return err
	}

	// Assign the user a policy that includes the bucket and prefix
	err = x.assignUserIAMPolicyForS3(data, userName, payload.Project.Slug)
	if err != nil {
		x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return err
	}

	// Test that the API key has write access to the bucket
	log.Warn("Testing Permissions by Writing Sample File")

	x.sendS3Response(e, transistor.GetAction("status"), transistor.GetState("running"), "Waiting for AWS to propagate credentials", nil)
	err = x.verifyS3CredentialsValid(e, data, userName, accessCredentials, payload.Project.Slug)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	// Report Success
	log.Info(fmt.Sprintf("Success Generating Credentials! %s (S3 EXTENSION)", payload.Project.Slug))
	x.sendS3Response(e, transistor.GetAction("status"),
		transistor.GetState("complete"),
		"S3 Provisioning Complete.\nRemoving this extension does not delete any data.",
		x.buildResultArtifacts(data, payload.Project.Slug, accessCredentials))

	return nil
}

// This is a helper function for x.updateS3 so that we don't have to deal with the
// multi-value return statement when shoving the data back into another array of artifacts
func (x *S3) getArtifactIgnoreError(e *transistor.Event, artifactName string) transistor.Artifact {
	artifact, err := e.GetArtifact(artifactName)
	if err != nil {
		return transistor.Artifact{}
	}

	return artifact
}

// Handle the 'update' message when received
func (x *S3) updateS3(e transistor.Event) error {
	ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Nothing to Update. Removing this extension does not delete any data.")

	// Grab the artifacts from the previous run
	ev.Artifacts = []transistor.Artifact{
		x.getArtifactIgnoreError(&e, "s3_aws_access_key_id"),
		x.getArtifactIgnoreError(&e, "s3_aws_secret_key"),
		x.getArtifactIgnoreError(&e, "s3_aws_region"),
		x.getArtifactIgnoreError(&e, "s3_aws_bucket"),
		x.getArtifactIgnoreError(&e, "s3_aws_prefix"),
	}

	x.events <- ev
	return nil
}

// Delete any dependent resources that are necessary to ensure
// completion of the DeleteUser request
//
// Currently this includes:
//		AccessKeys
//		Assigned User Policies
//		User Group Assignment
//
func (x *S3) deleteUserDependencies(data *S3Data, userName string) error {
	iamSvc := x.S3Interfaces.GetIAMServiceInterface(data)

	// We need to know what access keys there are in order to enumerate and delete them
	// there is no delete all access keys call
	listAccessKeys, err := iamSvc.ListAccessKeys(&iam.ListAccessKeysInput{UserName: &userName})
	if err != nil {
		log.Error(err.Error())
	} else {
		for _, accessKey := range listAccessKeys.AccessKeyMetadata {
			_, err := iamSvc.DeleteAccessKey(&iam.DeleteAccessKeyInput{UserName: &userName, AccessKeyId: accessKey.AccessKeyId})
			if err != nil {
				log.Error(err.Error())
			}
		}
	}

	_, err = iamSvc.RemoveUserFromGroup(&iam.RemoveUserFromGroupInput{UserName: &userName, GroupName: &data.AWSUserGroupName})
	if err != nil {
		log.Error(err.Error())
	}

	_, err = iamSvc.DeleteUserPolicy(&iam.DeleteUserPolicyInput{UserName: &userName, PolicyName: aws.String("codeamp-storage")})
	if err != nil {
		log.Error(err.Error())
	}

	return nil
}

// When the extension is deleted, revoke access to the path that was created for this project
// Since this data may need to be retained, and is not quickly deletable, we will simply revoke
// access and worry about cleaning up stale/old data in another process
func (x *S3) deleteS3(e transistor.Event) error {
	var data *S3Data
	var err error

	if data, err = x.extractArtifacts(e); err != nil {
		log.Error(err.Error())
		return err
	}

	iamSvc := x.S3Interfaces.GetIAMServiceInterface(data)

	payload := e.Payload.(plugins.ProjectExtension)
	userName := fmt.Sprintf("%s%s", data.AWSGeneratedUserPrefix, payload.Project.Slug)

	// Check to see if user exists before we delete
	_, err = iamSvc.GetUser(&iam.GetUserInput{UserName: &userName})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchEntity") {
			log.Error("No User to Delete (S3 EXTENSION)")
		}

		return nil
	}

	log.Warn("DELETING AWS USER (S3 EXTENSION): ", userName)
	x.deleteUserDependencies(data, userName)
	_, err = iamSvc.DeleteUser(&iam.DeleteUserInput{UserName: &userName})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}
