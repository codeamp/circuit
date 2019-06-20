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
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"

	_ "github.com/davecgh/go-spew/spew"
)

// Secret Resolver Mutation
type ReleaseResolverMutation struct {
	// DB
	DB *gorm.DB
	// Events
	Events chan transistor.Event
}

type ReleaseComponents struct {
	Project           *model.Project
	Environment       *model.Environment
	Services          []model.Service
	Secrets           []model.Secret
	ProjectExtensions []model.ProjectExtension
	HeadFeature       model.Feature
	TailFeature       model.Feature
}

// CreateRelease
// Workflows to support:
// 1. A fresh Release   					✓
// 2. A redeploy
// 3. A redeploy AND rebuild
// 4. A queued release   					✓
// 5. A rollback to an existing release
func (r *ReleaseResolverMutation) CreateRelease(ctx context.Context, args *struct{ Release *model.ReleaseInput }) (*ReleaseResolver, error) {
	// Exit Early Under the following conditions:
	// 1. User is not authed
	// 2. The project does not exist
	// 3. The environment does not exist
	// 4. The release ID to rollback to exists
	// 5. Project does not have permission to create a release for an environment
	// 6. Project does not have any extensions
	// X. Project does not have any workflow extensions
	// 7. No other waiting releases with exact same configuration
	// 8. Environment, TailFeature, and HeadFeature are all valid
	/******************************************
	*
	*	1. Check User Auth for Endpoint
	*
	*******************************************/
	log.Warn("Create RelEase 1")
	userID, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		log.Error("User not allowed to utilize endpoint")
		return nil, err
	}

	releaseComponents, err := r.PrepRelease(args.Release.ProjectID, args.Release.EnvironmentID, args.Release.HeadFeatureID, args.Release.ID)
	if err != nil {
		return nil, err
	}

	/******************************************
	*
	*	Create the Release
	*
	*******************************************/
	log.Warn("Create RelEase 9")
	forceRebuild := false
	release, err := r.createRelease(userID, args.Release.ProjectID, args.Release.EnvironmentID, args.Release.HeadFeatureID, forceRebuild,
		releaseComponents.Secrets, releaseComponents.Services, releaseComponents.ProjectExtensions)
	if err != nil {
		return nil, err
	}

	/******************************************
	*
	*	Dispatch Release event
	*
	*******************************************/
	releaseEvent, _ := r.BuildReleaseEvent(release, releaseComponents)
	r.Events <- *releaseEvent

	log.Warn("Create RelEase 10")

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

func (r *ReleaseResolverMutation) PrepRelease(projectID string, environmentID string,
	headFeatureID string, releaseID *string) (*ReleaseComponents, error) {
	/******************************************
	*
	*	2. Verify Project Existence
	*
	*******************************************/
	log.Warn("Create PrepRelease 2")
	var project model.Project
	if err := r.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			log.ErrorWithFields("Project not found", log.Fields{
				"id": projectID,
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
	log.Warn("Create PrepRelease 3")
	var rollbackRelease model.Release
	log.Warn("RELEASE ID PROVIDED TO PREP RELEASE: ", releaseID)
	if releaseID != nil && *releaseID != "" && r.DB.Where("id = ?", string(*releaseID)).First(&rollbackRelease).RecordNotFound() {
		log.ErrorWithFields("Could not find existing release", log.Fields{
			"id": *releaseID,
		})
		return nil, errors.New("Release not found")
	}

	/******************************************
	*
	*	4. Check Project Auth for Release Permissions
	*
	*******************************************/
	log.Warn("Create PrepRelease 4")
	// Check if project can create release in environment
	if r.isAuthorizedReleaseForEnvironment(projectID, environmentID) == false {
		log.Error("Project not allowed to create release in given environment")
		return nil, errors.New("Project not allowed to create release in given environment")
	}

	/******************************************
	*
	*	5. Ensure Project has Extensions
	*
	*******************************************/
	log.Warn("Create PrepRelease 5")
	projectExtensions, err := r.getProjectExtensions(projectID, environmentID)
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
	services, _ := r.gatherAndBuildServices(projectID, environmentID)

	/******************************************
	*
	*	Prepare Secrets
	*
	******************************************/
	secrets, _ := r.gatherAndBuildSecrets(projectID, environmentID)

	/************************************
	*
	* 	6. Ensure no other waiting releases with same
	*	secrets and services signatures
	*
	*************************************/
	log.Warn("Create RelEase 6")
	// ADB This is currently broken.
	// if r.isReleasePending(projectID, environmentID, headFeatureID, secrets, services) {
	// 	// same release so return
	// 	log.Warn("Found a waiting release with the same services signature, secrets signature and head feature hash. Aborting.")
	// 	// , log.Fields{
	// 	// 	"services_sig":      servicesSig,
	// 	// 	"secrets_sig":       secretsSig,
	// 	// 	"head_feature_hash": waitingReleaseHeadFeature.Hash,
	// 	// })
	// 	return nil, errors.New("Found a waiting release with the same properties. Aborting.")
	// }
	if len(services) > 0 {
		log.Warn("Mid Prep: ", services[0].Count)
	} else {
		log.Warn(services)
	}

	/************************************
	*
	* 	7. Validate Environment, HeadFeature, TailFeature
	*
	*************************************/
	log.Warn("Create RelEase 7")
	// Ensure Environment, HeadFeature, and TailFeature all exist
	var environment model.Environment
	if err := r.DB.Where("id = ?", environmentID).Find(&environment).Error; err != nil {
		log.InfoWithFields("no env found", log.Fields{
			"id": environmentID,
		})
		return nil, errors.New("Environment not found")
	}

	var headFeature model.Feature
	if err := r.DB.Where("id = ?", headFeatureID).First(&headFeature).Error; err != nil {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": headFeatureID,
		})
		return nil, errors.New("head feature not found")
	}

	var tailFeature model.Feature
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
	log.Warn("Create PrepRelease 8")
	secrets = r.injectReleaseEnvVars(secrets, &project, &headFeature)

	// If this is a new release, then generate a new secrets and services config
	// if not then reuse the old one on a previous RelEase
	if releaseID != nil {
		log.Warn("GETTING RELEASE CONFIGURATION FROM OLD RELEASE")
		_secrets, _services, _projectExtensions, err := r.getReleaseConfiguration(*releaseID)
		if err != nil {
			log.Error(err)
		} else {
			services = _services
			secrets = _secrets
			projectExtensions = _projectExtensions
		}
	}

	if len(services) > 0 {
		log.Warn("End Prep: ", services[0].Count)
	} else {
		log.Warn(services)
	}
	return &ReleaseComponents{&project, &environment, services, secrets, projectExtensions, headFeature, tailFeature}, nil
}

func (r *ReleaseResolverMutation) BuildRelease(userID string, projectID string, environmentID string,
	headFeatureID string, forceRebuild bool, secrets []model.Secret, services []model.Service, projectExtensions []model.ProjectExtension) (*model.Release, error) {
	// the tail feature id is the current release's head feature id
	// this is incorrect when the same commit is deployed multiple times
	// or when there is a rollback condition
	currentRelease := model.Release{}
	tailFeatureID := headFeatureID
	if err := r.DB.Where("state = ? and project_id = ? and environment_id = ?", transistor.GetState("complete"), projectID, environmentID).Find(&currentRelease).Order("created_at desc").Limit(1).Error; err == nil {
		tailFeatureID = currentRelease.HeadFeatureID.String()
	}

	// Convert model.Services to plugin.Services
	pluginServices, err := r.setupServices(services)
	if err != nil {
		return nil, err
	}

	/******************************************
	*
	*	Gather Services & Configure
	*
	*******************************************/
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

	return &release, nil
}

func (r *ReleaseResolverMutation) createRelease(userID string, projectID string, environmentID string,
	headFeatureID string, forceRebuild bool, secrets []model.Secret, services []model.Service, projectExtensions []model.ProjectExtension) (*model.Release, error) {

	release, err := r.BuildRelease(userID, projectID, environmentID,
		headFeatureID, forceRebuild, secrets, services, projectExtensions)
	if err != nil {
		return nil, err
	}

	if err := r.DB.Create(&release).Error; err != nil {
		return nil, err
	}

	return release, nil
}

func (r *ReleaseResolverMutation) gatherAndBuildSecrets(projectID string, environmentID string) ([]model.Secret, error) {
	// Gather all env vars / "secrets" for this service
	secrets := []model.Secret{}

	err := r.DB.Where("environment_id = ? AND ((project_id = ? AND scope = ?) OR (scope = ?))", environmentID, projectID, "project", "global").Find(&secrets).Error
	if err != nil {
		log.Error(err)
	}

	for _, secret := range secrets {
		if err := r.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secret.Value).Error; err != nil {
			log.Error(err)
		}
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
	r.DB.LogMode(true)
	if err := r.DB.Where("project_id = ? and environment_id = ?", projectID, environmentID).Find(&services).Error; err != nil {
		return nil, err
	}
	r.DB.LogMode(false)

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
		if err := r.DB.Where("service_id = ?", service.Model.ID).Find(&ports).Error; err != nil {
			log.Error(err)
		}
		services[i].Ports = ports

		// Deployment Strategy
		deploymentStrategy := model.ServiceDeploymentStrategy{}
		if err := r.DB.Where("service_id = ?", service.Model.ID).Find(&deploymentStrategy).Error; err != nil {
			log.Error(err)
			return nil, err
		}
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

	}

	return services, nil
}

func (r *ReleaseResolverMutation) BuildReleaseEvent(release *model.Release, releaseComponents *ReleaseComponents) (*transistor.Event, error) {
	// get the branch set for this environment and project from project settings
	// var branch string
	var projectSettings model.ProjectSettings
	var branch string
	if err := r.DB.Where("environment_id = ? and project_id = ?", release.EnvironmentID.String(), release.ProjectID.String()).First(&projectSettings).Error; err != nil {
		log.ErrorWithFields("no env project branch found", log.Fields{})
		return nil, fmt.Errorf("no env project branch found")
	} else {
		branch = projectSettings.GitBranch
	}

	var pluginSecrets []plugins.Secret
	for _, secret := range releaseComponents.Secrets {
		pluginSecrets = append(pluginSecrets, plugins.Secret{
			Key:   secret.Key,
			Value: secret.Value.Value,
			Type:  secret.Type,
		})
	}

	pluginServices, err := r.setupServices(releaseComponents.Services)
	if err != nil {
		return nil, err
	}

	releaseEventPayload := plugins.Release{
		ID:          release.Model.ID.String(),
		Environment: releaseComponents.Environment.Name,
		HeadFeature: plugins.Feature{
			ID:         releaseComponents.HeadFeature.Model.ID.String(),
			Hash:       releaseComponents.HeadFeature.Hash,
			ParentHash: releaseComponents.HeadFeature.ParentHash,
			User:       releaseComponents.HeadFeature.User,
			Message:    releaseComponents.HeadFeature.Message,
			Created:    releaseComponents.HeadFeature.Created,
		},
		TailFeature: plugins.Feature{
			ID:         releaseComponents.TailFeature.Model.ID.String(),
			Hash:       releaseComponents.TailFeature.Hash,
			ParentHash: releaseComponents.TailFeature.ParentHash,
			User:       releaseComponents.TailFeature.User,
			Message:    releaseComponents.TailFeature.Message,
			Created:    releaseComponents.TailFeature.Created,
		},
		User: release.User.Email,
		Project: plugins.Project{
			ID:         releaseComponents.Project.Model.ID.String(),
			Slug:       releaseComponents.Project.Slug,
			Repository: releaseComponents.Project.Repository,
		},
		Git: plugins.Git{
			Url:           releaseComponents.Project.GitUrl,
			Branch:        branch,
			RsaPrivateKey: releaseComponents.Project.RsaPrivateKey,
		},
		Secrets:  pluginSecrets,
		Services: pluginServices,
	}

	event := transistor.NewEvent(transistor.EventName("release"), transistor.GetAction("create"), releaseEventPayload)
	return &event, nil
}

func (r *ReleaseResolverMutation) getReleaseConfiguration(releaseID string) ([]model.Secret, []model.Service, []model.ProjectExtension, error) {
	log.Warn(fmt.Sprintf("Existing Release. Rolling back %s", releaseID))

	var existingRelease model.Release
	if err := r.DB.Find(&existingRelease).Where("id = ?", releaseID).Error; err != nil {
		log.Error(err)
		return nil, nil, nil, err
	} else {
		var projectExtensions []model.ProjectExtension
		var secrets []model.Secret
		var services []model.Service

		// Rollback
		secretsJsonb := existingRelease.Secrets
		servicesJsonb := existingRelease.Services
		projectExtensionsJsonb := existingRelease.ProjectExtensions

		// unmarshal projectExtensionsJsonb and servicesJsonb into project extensions
		err := json.Unmarshal(projectExtensionsJsonb.RawMessage, &projectExtensions)
		if err != nil {
			return nil, nil, nil, errors.New("Could not unmarshal project extensions")
		}

		err = json.Unmarshal(servicesJsonb.RawMessage, &services)
		if err != nil {
			return nil, nil, nil, errors.New("Could not unmarshal services")
		}

		err = json.Unmarshal(secretsJsonb.RawMessage, &secrets)
		if err != nil {
			return nil, nil, nil, errors.New("Could not unmarshal secrets")
		}

		return secrets, services, projectExtensions, nil
	}
}

func (r *ReleaseResolverMutation) isReleasePending(projectID string, environmentID string, headFeatureID string, secrets []model.Secret, services []model.Service) bool {
	secretsJsonb := r.makeJsonb(secrets)
	servicesJsonb := r.makeJsonb(services)

	type Comparator struct {
		test int
	}

	// check if there's a previous release in waiting state that
	// has the same secrets and services signatures
	secretsSha1 := sha1.New()
	// log.Warn(string(secretsJsonb.RawMessage[:]))
	secretsSha1.Write(secretsJsonb.RawMessage)
	secretsSig := secretsSha1.Sum(nil)

	servicesSha1 := sha1.New()
	servicesSha1.Write(servicesJsonb.RawMessage)
	servicesSig := servicesSha1.Sum(nil)

	currentReleaseHeadFeature := model.Feature{}

	// Gather HeadFeature
	if err := r.DB.Where("id = ?", headFeatureID).First(&currentReleaseHeadFeature).Error; err != nil {
		log.Error(err)
	}

	waitingRelease := model.Release{}

	// Gather a release in waiting state for same project and environment
	r.DB.Where("state in (?) and project_id = ? and environment_id = ?", []string{string(transistor.GetState("waiting")),
		string(transistor.GetState("running"))}, projectID, environmentID).Order("created_at desc").First(&waitingRelease)

	// Convert sercrets and services into sha for comparison
	wrSecretsSha1 := sha1.New()
	// log.Warn(string(waitingRelease.Services.RawMessage[:]))
	// log.Warn(string(servicesJsonb.RawMessage[:]))

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
		log.Warn("There IS a release pending")
		return true
	}

	log.Warn("there is NO release pending")
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
	if err := r.DB.Where("id = ?", userID).Find(&user).Error; err != nil {
		log.Error(err)
	}

	var release model.Release
	var releaseExtensions []model.ReleaseExtension

	// Warn if no release extensions are found
	r.DB.Where("release_id = ?", args.ID).Find(&releaseExtensions)
	if len(releaseExtensions) < 1 {
		log.WarnWithFields("No release extensions found for release", log.Fields{
			"id": args.ID,
		})
		return nil, errors.New("No release extensions found for release")
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
	if err := r.DB.Save(&release).Error; err != nil {
		log.Error(err)
	}

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

func (r *ReleaseResolverMutation) injectReleaseEnvVars(secrets []model.Secret, project *model.Project, headFeature *model.Feature) []model.Secret {
	return append(secrets, []model.Secret{
		{
			Key: "CODEAMP_SLUG",
			Value: model.SecretValue{
				Value: project.Slug,
			},
			Type: plugins.GetType("env"),
		},
		{
			Key: "CODEAMP_HASH",
			Value: model.SecretValue{
				Value: headFeature.Hash[0:7],
			},
			Type: plugins.GetType("env"),
		},
		{
			Key: "CODEAMP_CREATED_AT",
			Value: model.SecretValue{
				Value: time.Now().Format(time.RFC3339),
			},
			Type: plugins.GetType("env"),
		},
		{
			Key: "CODEFLOW_SLUG",
			Value: model.SecretValue{
				Value: project.Slug,
			},
			Type: plugins.GetType("env"),
		},
		{
			Key: "CODEFLOW_HASH",
			Value: model.SecretValue{
				Value: headFeature.Hash[0:7],
			},
			Type: plugins.GetType("env"),
		},
		{
			Key: "CODEFLOW_CREATED_AT",
			Value: model.SecretValue{
				Value: time.Now().Format(time.RFC3339),
			},
			Type: plugins.GetType("env"),
		},
	}...)
}

func (r *ReleaseResolverMutation) setupServices(services []model.Service) ([]plugins.Service, error) {
	var pluginServices []plugins.Service
	for _, service := range services {
		var spec model.ServiceSpec
		if r.DB.Where("service_id = ?", service.Model.ID).First(&spec).RecordNotFound() {
			log.ErrorWithFields("servicespec not found", log.Fields{
				"service_id": service.Model.ID,
			})
			return nil, fmt.Errorf("ServiceSpec not found")
		}

		pluginServices = AppendPluginService(pluginServices, service, spec)
	}

	return pluginServices, nil
}
