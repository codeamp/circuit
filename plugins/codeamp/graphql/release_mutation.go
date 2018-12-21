package graphql_resolver

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
)

// Secret Resolver Mutation
type ReleaseResolverMutation struct {
	// DB
	DB *gorm.DB
	// Events
	Events chan transistor.Event
}

// CreateRelease
// Workflows to support:
// 1. A fresh Release
// 2. A redeploy
// 3. A redeploy AND rebuild
// 4. A queued release
// 5. A rollback to an existing release
func (r *ReleaseResolverMutation) CreateRelease(ctx context.Context, args *struct{ Release *model.ReleaseInput }) (*ReleaseResolver, error) {
	// Exit Early Under the following conditions:
	// 1. User is not authed
	// 2. The project does not exist
	// 3. The release ID to rollback to exists
	// 4. Project does not have permission to create a release for an environment
	// 5. Project does not have any extensions
	// X. Project does not have any workflow extensions
	// 6. No other waiting releases with exact same configuration
	// 7. Environment, TailFeature, and HeadFeature are all valid

	/******************************************
	*
	*	1. Check User Auth for Endpoint
	*
	*******************************************/
	userID, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		log.Error("User not allowed to utilize endpoint")
		return nil, err
	}

	/******************************************
	*
	*	2. Verify Project Existence
	*
	*******************************************/
	var project model.Project
	if err := r.DB.Where("id = ?", args.Release.ProjectID).First(&project).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			log.ErrorWithFields("Project not found", log.Fields{
				"id": args.Release.ProjectID,
			})
		} else {
			err = errors.New("Project not found")
		}

		return nil, err
	}

	/******************************************
	*
	*	3. Rollback release exists, if requested
	*
	*******************************************/
	var rollbackRelease model.Release
	if args.Release.ID != nil && *args.Release.ID != "" && r.DB.Where("id = ?", string(*args.Release.ID)).First(&rollbackRelease).RecordNotFound() {
		log.ErrorWithFields("Could not find existing release", log.Fields{
			"id": *args.Release.ID,
		})
		return nil, errors.New("Release not found")
	}

	/******************************************
	*
	*	4. Check Project Auth for Release Permissions
	*
	*******************************************/
	// Check if project can create release in environment
	if r.isAuthorizedReleaseForEnvironment(args.Release.ProjectID, args.Release.EnvironmentID) == false {
		log.Error("Project not allowed to create release in given environment")
		return nil, errors.New("Project not allowed to create release in given environment")
	}

	/******************************************
	*
	*	5. Ensure Project has Extensions
	*
	*******************************************/
	projectExtensions, err := r.getProjectExtensions(args.Release.ProjectID, args.Release.EnvironmentID)
	if err != nil {
		log.Error(err.Error())
		return nil, errors.New(err.Error())
	}
	if len(projectExtensions) == 0 {
		log.Error("No project extensions found for release")
		return nil, errors.New("No project extensions found")
	}

	/******************************************
	*
	*	Prepare Services
	*
	*******************************************/
	services, _ := r.gatherAndBuildServices(args.Release.ProjectID, args.Release.EnvironmentID)

	/******************************************
	*
	*	Prepare Secrets
	*
	******************************************/
	secrets, _ := r.gatherAndBuildSecrets(args.Release.ProjectID, args.Release.EnvironmentID)

	/************************************
	*
	* 	6. Ensure no other waiting releases with same
	*	secrets and services signatures
	*
	*************************************/
	if r.isReleasePending(args.Release.ProjectID, args.Release.EnvironmentID, args.Release.HeadFeatureID, secrets, services) {
		// same release so return
		log.Warn("Found a waiting release with the same services signature, secrets signature and head feature hash. Aborting.")
		// , log.Fields{
		// 	"services_sig":      servicesSig,
		// 	"secrets_sig":       secretsSig,
		// 	"head_feature_hash": waitingReleaseHeadFeature.Hash,
		// })
		return nil, errors.New("Found a waiting release with the same properties. Aborting.")
	}

	/************************************
	*
	* 	7. Validate Environment, HeadFeature, TailFeature
	*
	*************************************/
	// Ensure Environment, HeadFeature, and TailFeature all exist
	var environment model.Environment
	if r.DB.Where("id = ?", args.Release.EnvironmentID).Find(&environment).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.Release.EnvironmentID,
		})
		return nil, errors.New("Environment not found")
	}

	var headFeature model.Feature
	if r.DB.Where("id = ?", args.Release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": args.Release.HeadFeatureID,
		})
		return nil, errors.New("head feature not found")
	}

	// var tailFeature model.Feature
	// if r.DB.Where("id = ?", args.Release.TailFeatureID).First(&tailFeature).RecordNotFound() {
	// 	log.InfoWithFields("tail feature not found", log.Fields{
	// 		"id": args.Release.TailFeatureID,
	// 	})
	// 	return nil, errors.New("Tail feature not found")
	// }

	/******************************************
	*
	*	Insert Environment Variables
	*
	*******************************************/
	// secrets = r.injectReleaseEnvVars(secrets, &project, headFeature, tailFeature)

	// var pluginSecrets []plugins.Secret
	// for _, secret := range secrets {
	// 	pluginSecrets = append(pluginSecrets, plugins.Secret{
	// 		Key:   secret.Key,
	// 		Value: secret.Value.Value,
	// 		Type:  secret.Type,
	// 	})
	// }

	// If this is a new release, then generate a new secrets and services config
	// if not then reuse the old one on a previous release
	// This is a new release because no previous release ID was provided
	if args.Release.ID != nil {
		r.createRollback(*args.Release.ID)
	}

	/******************************************
	*
	*	Create the Release
	*
	*******************************************/
	forceRebuild := false
	release, err := r.createRelease(userID, args.Release.ProjectID, args.Release.EnvironmentID, args.Release.HeadFeatureID, forceRebuild,
		secrets, services, projectExtensions)
	if err != nil {
		return nil, err
	}

	/******************************************
	*
	*	Dispatch Release event
	*
	*******************************************/
	releaseEvent := r.buildReleaseEvent(release)
	spew.Dump(releaseEvent)

	// All releases should be queued
	return &ReleaseResolver{DBReleaseResolver: &db_resolver.ReleaseResolver{DB: r.DB, Release: *release}}, nil
}

func (r *ReleaseResolverMutation) makeUUIDFromString(source string) uuid.UUID {
	uuid_res, err := uuid.FromString(source)
	if err != nil {
		log.Error("Couldn't parse field in makeUUIDFromString")
		return uuid.FromStringOrNil("0")
	}

	return uuid_res
}

func (r *ReleaseResolverMutation) createRelease(userID string, projectID string, environmentID string,
	headFeatureID string, forceRebuild bool, secrets []model.Secret, services []model.Service, projectExtensions []model.ProjectExtension) (*model.Release, error) {
	// the tail feature id is the current release's head feature id
	// this is incorrect when the same commit is deployed multiple times
	// or when there is a rollback condition
	currentRelease := model.Release{}
	tailFeatureID := headFeatureID
	if err := r.DB.Where("state = ? and project_id = ? and environment_id = ?", transistor.GetState("complete"), projectID, environmentID).Find(&currentRelease).Order("created_at desc").Limit(1).Error; err == nil {
		tailFeatureID = currentRelease.HeadFeatureID.String()
	}

	// get the branch set for this environment and project from project settings
	// var branch string
	var projectSettings model.ProjectSettings
	var branch string
	if r.DB.Where("environment_id = ? and project_id = ?", environmentID, projectID).First(&projectSettings).RecordNotFound() {
		log.ErrorWithFields("no env project branch found", log.Fields{})
		return nil, fmt.Errorf("no env project branch found")
	} else {
		branch = projectSettings.GitBranch
	}

	log.Warn("Branch = ", branch)

	// Convert model.Services to plugin.Services
	// Why? Serialized later?
	pluginServices, err := r.setupServices(services)
	if err != nil {
		return nil, err
	}

	/******************************************
	*
	*	Gather Services & Configure
	*
	*******************************************/
	// check if any project extensions that are not 'once' exists
	isRollback := false
	release := model.Release{
		State:             transistor.GetState("waiting"),
		StateMessage:      "Release created",
		ProjectID:         r.makeUUIDFromString(projectID),
		EnvironmentID:     r.makeUUIDFromString(environmentID),
		UserID:            uuid.FromStringOrNil(userID),
		HeadFeatureID:     r.makeUUIDFromString(headFeatureID),
		TailFeatureID:     r.makeUUIDFromString(tailFeatureID),
		Secrets:           r.makeJsonb(secrets),
		Services:          r.makeJsonb(pluginServices),
		ProjectExtensions: r.makeJsonb(projectExtensions),
		ForceRebuild:      forceRebuild,
		IsRollback:        isRollback,
	}

	if err := r.DB.Create(&release).Error; err != nil {
		return nil, err
	}

	return &release, nil
}

func (r *ReleaseResolverMutation) gatherAndBuildSecrets(projectID string, environmentID string) ([]model.Secret, error) {
	// Gather all env vars / "secrets" for this service
	secrets := []model.Secret{}

	projectSecrets := []model.Secret{}
	r.DB.Where("environment_id = ? AND project_id = ? AND scope = ?", environmentID, projectID, "project").Find(&projectSecrets)
	for _, secret := range projectSecrets {
		var secretValue model.SecretValue
		r.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		secret.Value = secretValue
		secrets = append(secrets, secret)
	}

	globalSecrets := []model.Secret{}
	r.DB.Where("environment_id = ? AND scope = ?", environmentID, "global").Find(&globalSecrets)
	for _, secret := range globalSecrets {
		var secretValue model.SecretValue
		r.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		secret.Value = secretValue
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func (r *ReleaseResolverMutation) makeJsonb(in interface{}) postgres.Jsonb {
	rawBytes, err := json.Marshal(in)
	if err != nil {
		log.Error(err.Error())
	}

	return postgres.Jsonb{rawBytes}
}

func (r *ReleaseResolverMutation) gatherAndBuildServices(projectID string, environmentID string) ([]model.Service, error) {
	var services []model.Service

	// Find all services
	if err := r.DB.Where("project_id = ? and environment_id = ?", projectID, environmentID).Find(&services).Error; err != nil {
		return nil, err
	}

	if len(services) == 0 {
		log.ErrorWithFields("No services found", log.Fields{
			"project_id": projectID,
		})
		return nil, errors.New("No services found")
	}

	// Build service configuration
	for i, service := range services {
		// Ports
		ports := []model.ServicePort{}
		r.DB.Where("service_id = ?", service.Model.ID).Find(&ports)
		services[i].Ports = ports

		// Deployment Strategy
		deploymentStrategy := model.ServiceDeploymentStrategy{}
		r.DB.Where("service_id = ?", service.Model.ID).Find(&deploymentStrategy)
		services[i].DeploymentStrategy = deploymentStrategy

		// Readiness
		readinessProbe := model.ServiceHealthProbe{}
		err := r.DB.Where("service_id = ? and type = ?", service.Model.ID, "readinessProbe").Find(&readinessProbe).Error
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		readinessHeaders := []model.ServiceHealthProbeHttpHeader{}
		err = r.DB.Where("health_probe_id = ?", readinessProbe.ID).Find(&readinessHeaders).Error
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		readinessProbe.HttpHeaders = readinessHeaders
		services[i].ReadinessProbe = readinessProbe

		// Liveness
		livenessProbe := model.ServiceHealthProbe{}
		err = r.DB.Where("service_id = ? and type = ?", service.Model.ID, "livenessProbe").Find(&livenessProbe).Error
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}

		livenessHeaders := []model.ServiceHealthProbeHttpHeader{}
		err = r.DB.Where("health_probe_id = ?", livenessProbe.ID).Find(&livenessHeaders).Error
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		livenessProbe.HttpHeaders = livenessHeaders
		services[i].LivenessProbe = livenessProbe

		spew.Dump(service)
	}

	return services, nil
}

func (r *ReleaseResolverMutation) buildReleaseEvent(release *model.Release) *plugins.Release {
	releaseEvent := plugins.Release{}
	// 	IsRollback:  isRollback,
	// 	ID:          release.Model.ID.String(),
	// 	Environment: environment.Key,
	// 	HeadFeature: plugins.Feature{
	// 		ID:         headFeature.Model.ID.String(),
	// 		Hash:       headFeature.Hash,
	// 		ParentHash: headFeature.ParentHash,
	// 		User:       headFeature.User,
	// 		Message:    headFeature.Message,
	// 		Created:    headFeature.Created,
	// 	},
	// 	TailFeature: plugins.Feature{
	// 		ID:         tailFeature.Model.ID.String(),
	// 		Hash:       tailFeature.Hash,
	// 		ParentHash: tailFeature.ParentHash,
	// 		User:       tailFeature.User,
	// 		Message:    tailFeature.Message,
	// 		Created:    tailFeature.Created,
	// 	},
	// 	User: release.User.Email,
	// 	Project: plugins.Project{
	// 		ID:         project.Model.ID.String(),
	// 		Slug:       project.Slug,
	// 		Repository: project.Repository,
	// 	},
	// 	Git: plugins.Git{
	// 		Url:           project.GitUrl,
	// 		Branch:        branch,
	// 		RsaPrivateKey: project.RsaPrivateKey,
	// 	},
	// 	Secrets:  pluginSecrets,
	// 	Services: pluginServices,
	// }

	return &releaseEvent
}

func (r *ReleaseResolverMutation) createRollback(releaseID string) {
	/******************************************
	*
	*	Existing Release, ReleaseID Provided
	*	Rollback to this Release
	*
	*******************************************/
	log.Info(fmt.Sprintf("Existing Release. Rolling back %s", releaseID))
	// // Rollback
	// isRollback = true
	// existingRelease := model.Release{}

	// secretsJsonb := existingRelease.Secrets
	// servicesJsonb := existingRelease.Services
	// projectExtensionsJsonb := existingRelease.ProjectExtensions

	// // unmarshal projectExtensionsJsonb and servicesJsonb into project extensions
	// err := json.Unmarshal(projectExtensionsJsonb.RawMessage, &projectExtensions)
	// if err != nil {
	// 	return nil, errors.New("Could not unmarshal project extensions")
	// }

	// err = json.Unmarshal(servicesJsonb.RawMessage, &services)
	// if err != nil {
	// 	return nil, errors.New("Could not unmarshal services")
	// }

	// err = json.Unmarshal(secretsJsonb.RawMessage, &secrets)
	// if err != nil {
	// 	return nil, errors.New("Could not unmarshal secrets")
	// }
}

func (r *ReleaseResolverMutation) isReleasePending(projectID string, environmentID string, headFeatureID string, secrets []model.Secret, services []model.Service) bool {
	secretsJsonb := r.makeJsonb(secrets)
	servicesJsonb := r.makeJsonb(services)

	// check if there's a previous release in waiting state that
	// has the same secrets and services signatures
	secretsSha1 := sha1.New()
	log.Warn(string(secretsJsonb.RawMessage[:]))
	secretsSha1.Write(secretsJsonb.RawMessage)
	secretsSig := secretsSha1.Sum(nil)

	servicesSha1 := sha1.New()
	servicesSha1.Write(servicesJsonb.RawMessage)
	servicesSig := servicesSha1.Sum(nil)

	currentReleaseHeadFeature := model.Feature{}

	// Gather HeadFeature
	r.DB.Where("id = ?", headFeatureID).First(&currentReleaseHeadFeature)

	waitingRelease := model.Release{}

	// Gather a release in waiting state for same project and environment
	r.DB.Where("state in (?) and project_id = ? and environment_id = ?", []string{string(transistor.GetState("waiting")),
		string(transistor.GetState("running"))}, projectID, environmentID).Order("created_at desc").First(&waitingRelease)
	spew.Dump(waitingRelease)

	// Convert sercrets and services into sha for comparison
	wrSecretsSha1 := sha1.New()
	log.Warn(string(waitingRelease.Services.RawMessage[:]))
	wrSecretsSha1.Write(waitingRelease.Services.RawMessage)
	waitingReleaseSecretsSig := wrSecretsSha1.Sum(nil)

	wrServicesSha1 := sha1.New()
	wrServicesSha1.Write(waitingRelease.Services.RawMessage)
	waitingReleaseServicesSig := wrServicesSha1.Sum(nil)

	waitingReleaseHeadFeature := model.Feature{}

	// Gather HeadFeature for the "waiting" release
	r.DB.Where("id = ?", waitingRelease.HeadFeatureID).First(&waitingReleaseHeadFeature)

	// If we have found another release rolling back in the exact same configuration
	// abort this release process.
	log.Warn(strings.Compare(string(secretsSig[:]), string(waitingReleaseSecretsSig[:])))
	log.Warn(bytes.Equal(secretsSig, waitingReleaseSecretsSig))
	log.Warn(bytes.Equal(servicesSig, waitingReleaseServicesSig))
	log.Warn(strings.Compare(currentReleaseHeadFeature.Hash, waitingReleaseHeadFeature.Hash))

	if bytes.Equal(secretsSig, waitingReleaseSecretsSig) &&
		bytes.Equal(servicesSig, waitingReleaseServicesSig) &&
		strings.Compare(currentReleaseHeadFeature.Hash, waitingReleaseHeadFeature.Hash) == 0 {
		return true
	}

	return false
}

// Checks to see if Project has permission to create a Release for the given Environment
func (r *ReleaseResolverMutation) isAuthorizedReleaseForEnvironment(projectID string, environmentID string) bool {
	if r.DB.Where("environment_id = ? and project_id = ?", environmentID, projectID).Find(&model.ProjectEnvironment{}).RecordNotFound() {
		return false
	}

	return true
}

func (r *ReleaseResolverMutation) getProjectExtensions(projectID string, environmentID string) ([]model.ProjectExtension, error) {
	var projectExtensions []model.ProjectExtension
	if err := r.DB.Where("project_id = ? AND environment_id = ? AND state = ?", projectID, environmentID, transistor.GetState("complete")).Find(&projectExtensions).Error; err != nil {
		return nil, err
	}

	for _, projectExtension := range projectExtensions {
		extension := model.Extension{}
		if r.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.ErrorWithFields("extension spec not found", log.Fields{
				"id": projectExtension.ExtensionID,
			})
			return nil, fmt.Errorf("extension spec not found")
		}
	}

	if len(projectExtensions) == 0 {
		log.ErrorWithFields("project has no extensions", log.Fields{
			"project_id":     projectID,
			"environment_id": environmentID,
		})
		return nil, fmt.Errorf("no project extensions found")
	}

	return projectExtensions, nil
}

func (r *ReleaseResolverMutation) StopRelease(ctx context.Context, args *struct{ ID graphql.ID }) (*ReleaseResolver, error) {
	/******************************************
	*
	*	Check User Auth for Endpoint
	*
	*******************************************/
	userID, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return nil, err
	}

	// Find User from Auth ID
	var user model.User
	r.DB.Where("id = ?", userID).Find(&user)

	var release model.Release
	var releaseExtensions []model.ReleaseExtension

	// Warn if no release extensions are found
	r.DB.Where("release_id = ?", args.ID).Find(&releaseExtensions)
	if len(releaseExtensions) < 1 {
		log.WarnWithFields("No release extensions found for release: %s", log.Fields{
			"id": args.ID,
		})
	}

	// Error if the release we're stopping cannot be found
	if r.DB.Where("id = ?", args.ID).Find(&release).RecordNotFound() {
		log.WarnWithFields("Release not found", log.Fields{
			"id": args.ID,
		})

		return nil, errors.New("Release Not Found")
	}

	// Mark release as 'canceled' state
	release.State = transistor.GetState("canceled")
	release.StateMessage = fmt.Sprintf("Release canceled by %s", user.Email)
	r.DB.Save(&release)

	// Iterate thorugh release extensions
	// Send 'canceled' messages out to the extensions
	// that have not been marked as running
	for _, releaseExtension := range releaseExtensions {
		var projectExtension model.ProjectExtension
		if r.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
			log.WarnWithFields("Associated project extension not found", log.Fields{
				"id": args.ID,
				"release_extension_id": releaseExtension.ID,
				"project_extension_id": releaseExtension.ProjectExtensionID,
			})

			return nil, fmt.Errorf("project extension %s not found", releaseExtension.ProjectExtensionID)
		}

		// Find Extensions to match the searched for Project Extensions
		extension := model.Extension{}
		if r.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.InfoWithFields("extension not found", log.Fields{
				"id": projectExtension.Model.ID,
				"release_extension_id": releaseExtension.Model.ID,
			})
			return nil, fmt.Errorf("extension %s not found", projectExtension.ExtensionID)
		}

		// Send 'canceled' messages out to the extensions that re not yet running
		if releaseExtension.State == transistor.GetState("waiting") {
			releaseExtensionEvent := plugins.ReleaseExtension{
				ID:      releaseExtension.ID.String(),
				Project: plugins.Project{},
				Release: plugins.Release{
					ID: releaseExtension.ReleaseID.String(),
				},
				Environment: "",
			}

			// Update the release extension
			event := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), transistor.GetAction("create"), releaseExtensionEvent)
			event.State = transistor.GetState("canceled")
			event.StateMessage = fmt.Sprintf("Deployment Stopped By User %s", user.Email)
			r.Events <- event
		}
	}

	return &ReleaseResolver{DBReleaseResolver: &db_resolver.ReleaseResolver{DB: r.DB, Release: release}}, nil
}

func (r *ReleaseResolverMutation) injectReleaseEnvVars(pluginSecrets []plugins.Secret, project *model.Project, headFeature model.Feature, tailFeature model.Feature) []plugins.Secret {
	// insert CodeAmp envs
	slugSecret := plugins.Secret{
		Key:   "CODEAMP_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, slugSecret)

	hashSecret := plugins.Secret{
		Key:   "CODEAMP_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, hashSecret)

	timeSecret := plugins.Secret{
		Key:   "CODEAMP_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, timeSecret)

	// insert Codeflow envs - remove later
	_slugSecret := plugins.Secret{
		Key:   "CODEFLOW_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _slugSecret)

	_hashSecret := plugins.Secret{
		Key:   "CODEFLOW_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _hashSecret)

	_timeSecret := plugins.Secret{
		Key:   "CODEFLOW_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _timeSecret)

	return pluginSecrets
}

func (r *ReleaseResolverMutation) setupServices(services []model.Service) ([]plugins.Service, error) {
	var pluginServices []plugins.Service
	for _, service := range services {
		var spec model.ServiceSpec
		if r.DB.Where("service_id = ?", service.Model.ID).First(&spec).RecordNotFound() {
			log.ErrorWithFields("servicespec not found", log.Fields{
				"service_id": service.Model.ID,
			})
			return []plugins.Service{}, fmt.Errorf("ServiceSpec not found")
		}

		pluginServices = AppendPluginService(pluginServices, service, spec)
	}

	return pluginServices, nil
}
