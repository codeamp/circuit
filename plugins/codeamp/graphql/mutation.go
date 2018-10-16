package graphql_resolver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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
	"github.com/extemporalgenome/slug"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
)

// CreateProject Create project
func (r *Resolver) CreateProject(ctx context.Context, args *struct {
	Project *model.ProjectInput
}) (*ProjectResolver, error) {
	var userId string
	var err error
	if userId, err = auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var project model.Project

	protocol := "HTTPS"
	switch args.Project.GitProtocol {
	case "private", "PRIVATE", "ssh", "SSH":
		protocol = "SSH"
	case "public", "PUBLIC", "https", "HTTPS":
		protocol = "HTTPS"
	}

	// Check if project already exists with same name
	existingProject := model.Project{}
	res := plugins.GetRegexParams("(?P<host>(git@|https?:\\/\\/)([\\w\\.@]+)(\\/|:))(?P<owner>[\\w,\\-,\\_]+)\\/(?P<repo>[\\w,\\-,\\_]+)(.git){0,1}((\\/){0,1})", args.Project.GitUrl)
	repository := fmt.Sprintf("%s/%s", res["owner"], res["repo"])
	if r.DB.Where("repository = ?", repository).First(&existingProject).RecordNotFound() {
		log.WarnWithFields("[+] Project not found", log.Fields{
			"repository": repository,
		})
	} else {
		return nil, fmt.Errorf("This repository already exists. Try again with a different git url.")
	}

	project = model.Project{
		GitProtocol: protocol,
		GitUrl:      args.Project.GitUrl,
		Secret:      transistor.RandomString(30),
	}
	project.Name = repository
	project.Repository = repository
	project.Slug = slug.Slug(repository)

	deletedProject := model.Project{}
	if err := r.DB.Unscoped().Where("repository = ?", repository).First(&deletedProject).Error; err != nil {
		project.Model.ID = deletedProject.Model.ID
	}

	// priv *rsa.PrivateKey;
	priv, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		return nil, err
	}

	err = priv.Validate()
	if err != nil {
		return nil, err
	}

	// Get der format. priv_der []byte
	priv_der := x509.MarshalPKCS1PrivateKey(priv)

	// pem.Block
	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   priv_der,
	}

	// Public Key generation
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}

	project.RsaPrivateKey = string(pem.EncodeToMemory(&priv_blk))
	project.RsaPublicKey = string(ssh.MarshalAuthorizedKey(pub))

	r.DB.Create(&project)

	// Create git branch for env per env
	environments := []model.Environment{}
	r.DB.Find(&environments)
	if len(environments) == 0 {
		log.InfoWithFields("No envs found.", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("No envs found")
	}

	for _, environment := range environments {
		r.DB.Create(&model.ProjectSettings{
			EnvironmentID: environment.Model.ID,
			ProjectID:     project.Model.ID,
			GitBranch:     "master",
		})

		// Create ProjectEnvironment rows for default envs
		if environment.IsDefault {
			r.DB.Create(&model.ProjectEnvironment{
				EnvironmentID: environment.Model.ID,
				ProjectID:     project.Model.ID,
			})
		}
	}

	// Create user permission for project
	userPermission := model.UserPermission{
		UserID: uuid.FromStringOrNil(userId),
		Value:  fmt.Sprintf("projects/%s", project.Repository),
	}
	r.DB.Create(&userPermission)

	return &ProjectResolver{DBProjectResolver: &db_resolver.ProjectResolver{DB: r.DB, Project: project}}, nil
}

// UpdateProject Update project
func (r *Resolver) UpdateProject(args *struct {
	Project *model.ProjectInput
}) (*ProjectResolver, error) {
	var project model.Project

	if args.Project.ID == nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	if r.DB.Where("id = ?", args.Project.ID).First(&project).RecordNotFound() {
		log.WarnWithFields("Project not found", log.Fields{
			"id": args.Project.ID,
		})
		return nil, fmt.Errorf("Project not found.")
	}

	if args.Project.GitUrl != "" {
		project.GitUrl = args.Project.GitUrl
	}

	switch args.Project.GitProtocol {
	case "private", "PRIVATE", "ssh", "SSH":
		project.GitProtocol = "SSH"
		if strings.HasPrefix(project.GitUrl, "http") {
			project.GitUrl = fmt.Sprintf("git@%s:%s.git", strings.Split(strings.Split(project.GitUrl, "://")[1], "/")[0], project.Repository)
		}
	case "public", "PUBLIC", "https", "HTTPS":
		project.GitProtocol = "HTTPS"
		if strings.HasPrefix(project.GitUrl, "git@") {
			project.GitUrl = fmt.Sprintf("https://%s/%s.git", strings.Split(strings.Split(project.GitUrl, "@")[1], ":")[0], project.Repository)
		}
	}

	if args.Project.GitBranch != nil {
		projectID, err := uuid.FromString(*args.Project.ID)
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse project ID")
		}

		environmentID, err := uuid.FromString(*args.Project.EnvironmentID)
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse environment ID")
		}

		var projectSettings model.ProjectSettings
		if r.DB.Where("environment_id = ? and project_id = ?", environmentID, projectID).First(&projectSettings).RecordNotFound() {
			projectSettings.EnvironmentID = environmentID
			projectSettings.ProjectID = projectID
			projectSettings.GitBranch = *args.Project.GitBranch
			projectSettings.ContinuousDeploy = *args.Project.ContinuousDeploy

			r.DB.Save(&projectSettings)
		} else {
			projectSettings.GitBranch = *args.Project.GitBranch
			projectSettings.ContinuousDeploy = *args.Project.ContinuousDeploy

			r.DB.Save(&projectSettings)
		}
	}

	r.DB.Save(&project)

	return &ProjectResolver{DBProjectResolver: &db_resolver.ProjectResolver{DB: r.DB, Project: project}}, nil
}

// StopRelease
func (r *Resolver) StopRelease(ctx context.Context, args *struct{ ID graphql.ID }) (*ReleaseResolver, error) {
	userID, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return &ReleaseResolver{}, err
	}

	var user model.User

	r.DB.Where("id = ?", userID).Find(&user)

	var release model.Release
	var releaseExtensions []model.ReleaseExtension

	r.DB.Where("release_id = ?", args.ID).Find(&releaseExtensions)
	if len(releaseExtensions) < 1 {
		return nil, fmt.Errorf("No release extensions found for release: %s", args.ID)
	}

	if r.DB.Where("id = ?", args.ID).Find(&release).RecordNotFound() {
		log.WarnWithFields("Release not found", log.Fields{
			"id": args.ID,
		})

		return nil, errors.New("Release Not Found")
	}

	release.State = transistor.GetState("canceled")
	release.StateMessage = fmt.Sprintf("Release canceled by %s", user.Email)
	r.DB.Save(&release)

	for _, releaseExtension := range releaseExtensions {
		var projectExtension model.ProjectExtension
		if r.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
			log.WarnWithFields("Associated project extension not found", log.Fields{
				"id": args.ID,
				"release_extension_id": releaseExtension.ID,
				"project_extension_id": releaseExtension.ProjectExtensionID,
			})

			return nil, errors.New("Project Extension Not Found")
		}

		// find associated ProjectExtension Extension
		var extension model.Extension
		if r.DB.Where("id = ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.WarnWithFields("Associated extension not found", log.Fields{
				"id": args.ID,
				"release_extension_id": releaseExtension.ID,
				"project_extension_id": releaseExtension.ProjectExtensionID,
				"extension_id":         projectExtension.ExtensionID,
			})

			return nil, errors.New("Extension Not Found")
		}

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

// CreateRelease
func (r *Resolver) CreateRelease(ctx context.Context, args *struct{ Release *model.ReleaseInput }) (*ReleaseResolver, error) {
	var project model.Project
	var secrets []model.Secret
	var services []model.Service
	var projectExtensions []model.ProjectExtension
	var secretsJsonb postgres.Jsonb
	var servicesJsonb postgres.Jsonb
	var projectExtensionsJsonb postgres.Jsonb

	isRollback := false

	userID, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return nil, err
	}

	// Check if project can create release in environment
	if r.DB.Where("environment_id = ? and project_id = ?", args.Release.EnvironmentID, args.Release.ProjectID).Find(&model.ProjectEnvironment{}).RecordNotFound() {
		return nil, errors.New("Project not allowed to create release in given environment")
	}

	if args.Release.ID == nil {
		projectSecrets := []model.Secret{}
		// get all the env vars related to this release and store
		r.DB.Where("environment_id = ? AND project_id = ? AND scope = ?", args.Release.EnvironmentID, args.Release.ProjectID, "project").Find(&projectSecrets)
		for _, secret := range projectSecrets {
			var secretValue model.SecretValue
			r.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
			secret.Value = secretValue
			secrets = append(secrets, secret)
		}

		globalSecrets := []model.Secret{}
		r.DB.Where("environment_id = ? AND scope = ?", args.Release.EnvironmentID, "global").Find(&globalSecrets)
		for _, secret := range globalSecrets {
			var secretValue model.SecretValue
			r.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
			secret.Value = secretValue
			secrets = append(secrets, secret)
		}

		secretsMarshaled, err := json.Marshal(secrets)
		if err != nil {
			return &ReleaseResolver{}, err
		}

		secretsJsonb = postgres.Jsonb{secretsMarshaled}

		r.DB.Where("project_id = ? and environment_id = ?", args.Release.ProjectID, args.Release.EnvironmentID).Find(&services)
		if len(services) == 0 {
			log.InfoWithFields("no services found", log.Fields{
				"project_id": args.Release.ProjectID,
			})
		}

		for i, service := range services {
			ports := []model.ServicePort{}
			r.DB.Where("service_id = ?", service.Model.ID).Find(&ports)
			services[i].Ports = ports

			deploymentStrategy := model.ServiceDeploymentStrategy{}
			r.DB.Where("service_id = ?", service.Model.ID).Find(&deploymentStrategy)
			services[i].DeploymentStrategy = deploymentStrategy

			readinessProbe := model.ServiceHealthProbe{}
			err = r.DB.Where("service_id = ? and type = ?", service.Model.ID, "readinessProbe").Find(&readinessProbe).Error
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

		servicesMarshaled, err := json.Marshal(services)
		if err != nil {
			return &ReleaseResolver{}, err
		}

		servicesJsonb = postgres.Jsonb{servicesMarshaled}
		// check if any project extensions that are not 'once' exists
		r.DB.Where("project_id = ? AND environment_id = ? AND state = ?", args.Release.ProjectID, args.Release.EnvironmentID, transistor.GetState("complete")).Find(&projectExtensions)

		if len(projectExtensions) == 0 {
			log.InfoWithFields("project has no extensions", log.Fields{
				"project_id":     args.Release.ProjectID,
				"environment_id": args.Release.EnvironmentID,
			})
			return nil, fmt.Errorf("no project extensions found")
		}

		projectExtensionsMarshaled, err := json.Marshal(projectExtensions)
		if err != nil {
			return &ReleaseResolver{}, err
		}

		projectExtensionsJsonb = postgres.Jsonb{projectExtensionsMarshaled}
	} else {
		log.Info(fmt.Sprintf("Existing Release. Rolling back %d", args.Release.ID))
		// Rollback
		isRollback = true
		existingRelease := model.Release{}

		if r.DB.Where("id = ?", string(*args.Release.ID)).First(&existingRelease).RecordNotFound() {
			log.ErrorWithFields("Could not find existing release", log.Fields{
				"id": *args.Release.ID,
			})
			return &ReleaseResolver{}, errors.New("Release not found")
		}

		secretsJsonb = existingRelease.Secrets
		servicesJsonb = existingRelease.Services
		projectExtensionsJsonb = existingRelease.ProjectExtensions

		// unmarshal projectExtensionsJsonb and servicesJsonb into project extensions
		err := json.Unmarshal(projectExtensionsJsonb.RawMessage, &projectExtensions)
		if err != nil {
			return &ReleaseResolver{}, errors.New("Could not unmarshal project extensions")
		}

		err = json.Unmarshal(servicesJsonb.RawMessage, &services)
		if err != nil {
			return &ReleaseResolver{}, errors.New("Could not unmarshal services")
		}

		err = json.Unmarshal(secretsJsonb.RawMessage, &secrets)
		if err != nil {
			return &ReleaseResolver{}, errors.New("Could not unmarshal secrets")
		}
	}

	// check if there's a previous release in waiting state that
	// has the same secrets and services signatures
	secretsSha1 := sha1.New()
	secretsSha1.Write(secretsJsonb.RawMessage)
	secretsSig := secretsSha1.Sum(nil)

	servicesSha1 := sha1.New()
	servicesSha1.Write(servicesJsonb.RawMessage)
	servicesSig := servicesSha1.Sum(nil)

	currentReleaseHeadFeature := model.Feature{}

	r.DB.Where("id = ?", args.Release.HeadFeatureID).First(&currentReleaseHeadFeature)

	waitingRelease := model.Release{}

	r.DB.Where("state in (?) and project_id = ? and environment_id = ?", []string{string(transistor.GetState("waiting")),
		string(transistor.GetState("running"))}, args.Release.ProjectID, args.Release.EnvironmentID).Order("created_at desc").First(&waitingRelease)

	wrSecretsSha1 := sha1.New()
	wrSecretsSha1.Write(waitingRelease.Services.RawMessage)
	waitingReleaseSecretsSig := wrSecretsSha1.Sum(nil)

	wrServicesSha1 := sha1.New()
	wrServicesSha1.Write(waitingRelease.Services.RawMessage)
	waitingReleaseServicesSig := wrServicesSha1.Sum(nil)

	waitingReleaseHeadFeature := model.Feature{}

	r.DB.Where("id = ?", waitingRelease.HeadFeatureID).First(&waitingReleaseHeadFeature)

	if bytes.Equal(secretsSig, waitingReleaseSecretsSig) &&
		bytes.Equal(servicesSig, waitingReleaseServicesSig) &&
		strings.Compare(currentReleaseHeadFeature.Hash, waitingReleaseHeadFeature.Hash) == 0 {

		// same release so return
		log.InfoWithFields("Found a waiting release with the same services signature, secrets signature and head feature hash. Aborting", log.Fields{
			"services_sig":      servicesSig,
			"secrets_sig":       secretsSig,
			"head_feature_hash": waitingReleaseHeadFeature.Hash,
		})
		return &ReleaseResolver{}, fmt.Errorf("Found a waiting release with the same properties. Aborting.")
	}

	projectID, err := uuid.FromString(args.Release.ProjectID)
	if err != nil {
		log.InfoWithFields("Couldn't parse projectID", log.Fields{
			"args": args,
		})
		return &ReleaseResolver{}, fmt.Errorf("Couldn't parse projectID")
	}

	headFeatureID, err := uuid.FromString(args.Release.HeadFeatureID)
	if err != nil {
		log.InfoWithFields("Couldn't parse headFeatureID", log.Fields{
			"args": args,
		})
		return &ReleaseResolver{}, fmt.Errorf("Couldn't parse headFeatureID")
	}

	environmentID, err := uuid.FromString(args.Release.EnvironmentID)
	if err != nil {
		log.InfoWithFields("Couldn't parse environmentID", log.Fields{
			"args": args,
		})
		return &ReleaseResolver{}, fmt.Errorf("Couldn't parse environmentID")
	}

	// the tail feature id is the current release's head feature id
	currentRelease := model.Release{}
	tailFeatureID := headFeatureID
	if err = r.DB.Where("state = ? and project_id = ? and environment_id = ?", transistor.GetState("complete"), projectID, environmentID).Find(&currentRelease).Order("created_at desc").Limit(1).Error; err == nil {
		tailFeatureID = currentRelease.HeadFeatureID
	}

	if r.DB.Where("id = ?", projectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": projectID,
		})
		return &ReleaseResolver{}, errors.New("Project not found")
	}

	// get all branches relevant for the project
	var branch string
	var projectSettings model.ProjectSettings

	if r.DB.Where("environment_id = ? and project_id = ?", environmentID, projectID).First(&projectSettings).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	var environment model.Environment
	if r.DB.Where("id = ?", environmentID).Find(&environment).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": environmentID,
		})
		return &ReleaseResolver{}, errors.New("Environment not found")
	}

	var headFeature model.Feature
	if r.DB.Where("id = ?", headFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": headFeatureID,
		})
		return &ReleaseResolver{}, errors.New("head feature not found")
	}

	var tailFeature model.Feature
	if r.DB.Where("id = ?", tailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": tailFeatureID,
		})
		return &ReleaseResolver{}, errors.New("Tail feature not found")
	}

	var pluginServices []plugins.Service
	pluginServices, err = r.setupServices(services)
	if err != nil {
		return &ReleaseResolver{}, err
	}

	var pluginSecrets []plugins.Secret
	for _, secret := range secrets {
		pluginSecrets = append(pluginSecrets, plugins.Secret{
			Key:   secret.Key,
			Value: secret.Value.Value,
			Type:  secret.Type,
		})
	}

	// Create/Emit Release ProjectExtensions
	willCreateReleaseExtension := false
	for _, projectExtension := range projectExtensions {
		extension := model.Extension{}
		if r.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.ErrorWithFields("extension spec not found", log.Fields{
				"id": projectExtension.ExtensionID,
			})
			return &ReleaseResolver{}, fmt.Errorf("extension spec not found")
		}

		if plugins.Type(extension.Type) == plugins.GetType("workflow") || plugins.Type(extension.Type) == plugins.GetType("deployment") {
			willCreateReleaseExtension = true
			break
		}
	}

	var release model.Release
	if willCreateReleaseExtension == true {
		// Create Release
		release = model.Release{
			State:             transistor.GetState("waiting"),
			StateMessage:      "Release created",
			ProjectID:         projectID,
			EnvironmentID:     environmentID,
			UserID:            uuid.FromStringOrNil(userID),
			HeadFeatureID:     headFeatureID,
			TailFeatureID:     tailFeatureID,
			Secrets:           secretsJsonb,
			Services:          servicesJsonb,
			ProjectExtensions: projectExtensionsJsonb,
			ForceRebuild:      args.Release.ForceRebuild,
			IsRollback:        isRollback,
		}

		r.DB.Create(&release)
	} else {
		return nil, fmt.Errorf("No release extensions found")
	}

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

	releaseEvent := plugins.Release{
		IsRollback:  isRollback,
		ID:          release.Model.ID.String(),
		Environment: environment.Key,
		HeadFeature: plugins.Feature{
			ID:         headFeature.Model.ID.String(),
			Hash:       headFeature.Hash,
			ParentHash: headFeature.ParentHash,
			User:       headFeature.User,
			Message:    headFeature.Message,
			Created:    headFeature.Created,
		},
		TailFeature: plugins.Feature{
			ID:         tailFeature.Model.ID.String(),
			Hash:       tailFeature.Hash,
			ParentHash: tailFeature.ParentHash,
			User:       tailFeature.User,
			Message:    tailFeature.Message,
			Created:    tailFeature.Created,
		},
		User: release.User.Email,
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Git: plugins.Git{
			Url:           project.GitUrl,
			Branch:        branch,
			RsaPrivateKey: project.RsaPrivateKey,
		},
		Secrets:  pluginSecrets,
		Services: pluginServices, // ADB Added this
	}

	for _, projectExtension := range projectExtensions {
		extension := model.Extension{}
		if r.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.ErrorWithFields("extension spec not found", log.Fields{
				"id": projectExtension.ExtensionID,
			})
			return &ReleaseResolver{}, errors.New("extension spec not found")
		}

		if plugins.Type(extension.Type) == plugins.GetType("workflow") || plugins.Type(extension.Type) == plugins.GetType("deployment") {
			var headFeature model.Feature
			if r.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
				log.ErrorWithFields("head feature not found", log.Fields{
					"id": release.HeadFeatureID,
				})
				return &ReleaseResolver{}, errors.New("head feature not found")
			}

			// create ReleaseExtension
			releaseExtension := model.ReleaseExtension{
				State:              transistor.GetState("waiting"),
				StateMessage:       "",
				ReleaseID:          release.Model.ID,
				FeatureHash:        headFeature.Hash,
				ServicesSignature:  fmt.Sprintf("%x", servicesSig),
				SecretsSignature:   fmt.Sprintf("%x", secretsSig),
				ProjectExtensionID: projectExtension.Model.ID,
				Type:               extension.Type,
			}

			r.DB.Create(&releaseExtension)
		}
	}

	if waitingRelease.State != "" {
		log.Info(fmt.Sprintf("Release is already running, queueing %s", release.Model.ID.String()))
		return &ReleaseResolver{}, fmt.Errorf("Release is already running, queuing %s", release.Model.ID.String())
	} else {
		release.Started = time.Now()
		r.DB.Save(&release)

		r.Events <- transistor.NewEvent(transistor.EventName("release"), transistor.GetAction("create"), releaseEvent)

		return &ReleaseResolver{DBReleaseResolver: &db_resolver.ReleaseResolver{DB: r.DB, Release: release}}, nil
	}
}

// CreateService Create service
func (r *Resolver) CreateService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	// Check service name length
	if len(args.Service.Name) > 63 {
		return nil, fmt.Errorf("Service name cannot be longer than 63 characters.")
	}

	// Check if project can create service in environment
	if r.DB.Where("environment_id = ? and project_id = ?", args.Service.EnvironmentID, args.Service.ProjectID).Find(&model.ProjectEnvironment{}).RecordNotFound() {
		return nil, fmt.Errorf("Project not allowed to create service in given environment")
	}

	projectID, err := uuid.FromString(args.Service.ProjectID)
	if err != nil {
		return nil, err
	}

	environmentID, err := uuid.FromString(args.Service.EnvironmentID)
	if err != nil {
		return nil, err
	}

	serviceSpecID, err := uuid.FromString(args.Service.ServiceSpecID)
	if err != nil {
		return nil, err
	}

	var deploymentStrategy model.ServiceDeploymentStrategy
	if args.Service.DeploymentStrategy != nil {
		deploymentStrategy, err = validateDeploymentStrategyInput(args.Service.DeploymentStrategy)
		if err != nil {
			return nil, err
		}
	}

	var livenessProbe model.ServiceHealthProbe
	if args.Service.LivenessProbe != nil {
		probeType := plugins.GetType("livenessProbe")
		probe := args.Service.LivenessProbe
		probe.Type = &probeType
		livenessProbe, err = validateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var readinessProbe model.ServiceHealthProbe
	if args.Service.ReadinessProbe != nil {
		probeType := plugins.GetType("readinessProbe")
		probe := args.Service.ReadinessProbe
		probe.Type = &probeType
		readinessProbe, err = validateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var preStopHook string
	if args.Service.PreStopHook != nil {
		preStopHook = *args.Service.PreStopHook
	}

	service := model.Service{
		Name:               args.Service.Name,
		Command:            args.Service.Command,
		ServiceSpecID:      serviceSpecID,
		Type:               plugins.Type(args.Service.Type),
		Count:              args.Service.Count,
		ProjectID:          projectID,
		EnvironmentID:      environmentID,
		DeploymentStrategy: deploymentStrategy,
		LivenessProbe:      livenessProbe,
		ReadinessProbe:     readinessProbe,
		PreStopHook:        preStopHook,
	}

	r.DB.Create(&service)

	// Create Health Probe Headers
	if service.LivenessProbe.HttpHeaders != nil {
		for _, h := range service.LivenessProbe.HttpHeaders {
			h.HealthProbeID = service.LivenessProbe.ID
			r.DB.Create(&h)
		}
	}

	if service.ReadinessProbe.HttpHeaders != nil {
		for _, h := range service.ReadinessProbe.HttpHeaders {
			h.HealthProbeID = service.ReadinessProbe.ID
			r.DB.Create(&h)
		}
	}

	if args.Service.Ports != nil {
		for _, cp := range *args.Service.Ports {
			servicePort := model.ServicePort{
				ServiceID: service.ID,
				Port:      cp.Port,
				Protocol:  cp.Protocol,
			}
			r.DB.Create(&servicePort)
		}
	}

	return &ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: r.DB, Service: service}}, nil
}

// UpdateService Update Service
func (r *Resolver) UpdateService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	serviceID := uuid.FromStringOrNil(*args.Service.ID)
	serviceSpecID := uuid.FromStringOrNil(args.Service.ServiceSpecID)

	if serviceID == uuid.Nil || serviceSpecID == uuid.Nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	var service model.Service
	if r.DB.Where("id = ?", serviceID).Find(&service).RecordNotFound() {
		return nil, fmt.Errorf("Record not found with given argument id")
	}

	service.Command = args.Service.Command
	service.Name = args.Service.Name
	service.Type = plugins.Type(args.Service.Type)
	service.ServiceSpecID = serviceSpecID
	service.Count = args.Service.Count

	r.DB.Save(&service)

	// delete all previous container ports
	var servicePorts []model.ServicePort
	r.DB.Where("service_id = ?", serviceID).Find(&servicePorts)

	// delete all container ports
	// replace with current
	for _, cp := range servicePorts {
		r.DB.Delete(&cp)
	}

	if args.Service.Ports != nil {
		for _, cp := range *args.Service.Ports {
			servicePort := model.ServicePort{
				ServiceID: service.ID,
				Port:      cp.Port,
				Protocol:  cp.Protocol,
			}
			r.DB.Create(&servicePort)
		}
	}

	var livenessProbe = model.ServiceHealthProbe{}
	var err error
	if args.Service.LivenessProbe != nil {
		probeType := plugins.GetType("livenessProbe")
		probe := args.Service.LivenessProbe
		probe.Type = &probeType
		livenessProbe, err = validateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var readinessProbe = model.ServiceHealthProbe{}
	if args.Service.ReadinessProbe != nil {
		probeType := plugins.GetType("readinessProbe")
		probe := args.Service.ReadinessProbe
		probe.Type = &probeType
		readinessProbe, err = validateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var oldHealthProbes []model.ServiceHealthProbe
	r.DB.Where("service_id = ?", serviceID).Find(&oldHealthProbes)
	for _, probe := range oldHealthProbes {
		var headers []model.ServiceHealthProbeHttpHeader
		r.DB.Where("health_probe_id = ?", probe.ID).Find(&headers)
		for _, header := range headers {
			r.DB.Delete(&header)
		}
		r.DB.Delete(&probe)
	}

	var deploymentStrategy model.ServiceDeploymentStrategy
	r.DB.Where("service_id = ?", serviceID).Find(&deploymentStrategy)
	updatedDeploymentStrategy, err := validateDeploymentStrategyInput(args.Service.DeploymentStrategy)
	if err != nil {
		return nil, err
	}

	deploymentStrategy.Type = updatedDeploymentStrategy.Type
	deploymentStrategy.MaxUnavailable = updatedDeploymentStrategy.MaxUnavailable
	deploymentStrategy.MaxSurge = updatedDeploymentStrategy.MaxSurge

	r.DB.Save(&deploymentStrategy)
	service.DeploymentStrategy = deploymentStrategy
	service.ReadinessProbe = readinessProbe
	service.LivenessProbe = livenessProbe

	var preStopHook string
	if args.Service.PreStopHook != nil {
		preStopHook = *args.Service.PreStopHook
	}
	service.PreStopHook = preStopHook
	r.DB.Save(&service)

	// Create Health Probe Headers
	for _, h := range service.LivenessProbe.HttpHeaders {
		h.HealthProbeID = service.LivenessProbe.ID
		r.DB.Create(&h)
	}

	for _, h := range service.ReadinessProbe.HttpHeaders {
		h.HealthProbeID = service.ReadinessProbe.ID
		r.DB.Create(&h)
	}

	return &ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: r.DB, Service: service}}, nil
}

// DeleteService Delete service
func (r *Resolver) DeleteService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	serviceID, err := uuid.FromString(*args.Service.ID)

	if err != nil {
		return nil, err
	}

	var service model.Service

	r.DB.Where("id = ?", serviceID).Find(&service)
	r.DB.Delete(&service)

	// delete all previous container ports
	var servicePorts []model.ServicePort
	r.DB.Where("service_id = ?", serviceID).Find(&servicePorts)

	// delete all container ports
	for _, cp := range servicePorts {
		r.DB.Delete(&cp)
	}

	var healthProbes []model.ServiceHealthProbe
	r.DB.Where("service_id = ?", serviceID).Find(&healthProbes)
	for _, probe := range healthProbes {
		var headers []model.ServiceHealthProbeHttpHeader
		r.DB.Where("health_probe_id = ?", probe.ID).Find(&headers)
		for _, header := range headers {
			r.DB.Delete(&header)
		}
		r.DB.Delete(&probe)
	}

	var deploymentStrategy model.ServiceDeploymentStrategy
	r.DB.Where("service_id = ?", serviceID).Find(&deploymentStrategy)
	r.DB.Delete(&deploymentStrategy)

	return &ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: r.DB, Service: service}}, nil
}

func validateHealthProbe(input model.ServiceHealthProbeInput) (model.ServiceHealthProbe, error) {
	healthProbe := model.ServiceHealthProbe{}

	switch probeType := *input.Type; probeType {
	case plugins.GetType("livenessProbe"), plugins.GetType("readinessProbe"):
		healthProbe.Type = probeType
		if input.InitialDelaySeconds != nil {
			healthProbe.InitialDelaySeconds = *input.InitialDelaySeconds
		}
		if input.PeriodSeconds != nil {
			healthProbe.PeriodSeconds = *input.PeriodSeconds
		}
		if input.TimeoutSeconds != nil {
			healthProbe.TimeoutSeconds = *input.TimeoutSeconds
		}
		if input.SuccessThreshold != nil {
			healthProbe.SuccessThreshold = *input.SuccessThreshold
		}
		if input.FailureThreshold != nil {
			healthProbe.FailureThreshold = *input.FailureThreshold
		}
	default:
		return model.ServiceHealthProbe{}, fmt.Errorf("Unsuported Probe Type %s", string(*input.Type))
	}

	switch probeMethod := input.Method; probeMethod {
	case "default", "":
		return model.ServiceHealthProbe{}, nil
	case "exec":
		healthProbe.Method = input.Method
		if input.Command == nil {
			return model.ServiceHealthProbe{}, fmt.Errorf("Command is required if Probe method is exec")
		}
		healthProbe.Command = *input.Command
	case "http":
		healthProbe.Method = input.Method
		if input.Port == nil {
			return model.ServiceHealthProbe{}, fmt.Errorf("http probe require a port to be set")
		}
		healthProbe.Port = *input.Port
		if input.Path == nil {
			return model.ServiceHealthProbe{}, fmt.Errorf("http probe requires a path to be set")
		}
		healthProbe.Path = *input.Path

		// httpStr := "http"
		// httpsStr := "https"
		if input.Scheme == nil || (*input.Scheme != "http" && *input.Scheme != "https") {
			return model.ServiceHealthProbe{}, fmt.Errorf("http probe requires scheme to be set to either http or https")
		}
		healthProbe.Scheme = *input.Scheme
	case "tcp":
		healthProbe.Method = input.Method
		if input.Port == nil {
			return model.ServiceHealthProbe{}, fmt.Errorf("tcp probe requires a port to be set")
		}
		healthProbe.Port = *input.Port
	default:
		return model.ServiceHealthProbe{}, fmt.Errorf("Unsuported Probe Method %s", string(input.Method))
	}

	if input.HttpHeaders != nil {
		for _, headerInput := range *input.HttpHeaders {
			header := model.ServiceHealthProbeHttpHeader{
				Name:          headerInput.Name,
				Value:         headerInput.Value,
				HealthProbeID: healthProbe.ID,
			}
			healthProbe.HttpHeaders = append(healthProbe.HttpHeaders, header)
		}

	}

	return healthProbe, nil
}

func validateDeploymentStrategyInput(input *model.DeploymentStrategyInput) (model.ServiceDeploymentStrategy, error) {
	switch strategy := input.Type; strategy {
	case plugins.GetType("default"), plugins.GetType("recreate"):
		return model.ServiceDeploymentStrategy{Type: plugins.Type(input.Type)}, nil
	case plugins.GetType("rollingUpdate"):
		if input.MaxUnavailable == 0 {
			return model.ServiceDeploymentStrategy{}, fmt.Errorf("RollingUpdate DeploymentStrategy requires a valid maxUnavailable parameter")
		}

		if input.MaxSurge == 0 {
			return model.ServiceDeploymentStrategy{}, fmt.Errorf("RollingUpdate DeploymentStrategy requires a valid maxSurge parameter")
		}
	default:
		return model.ServiceDeploymentStrategy{}, fmt.Errorf("Unsuported Deployment Strategy %s", input.Type)
	}

	deploymentStrategy := model.ServiceDeploymentStrategy{
		Type:           plugins.Type(input.Type),
		MaxUnavailable: input.MaxUnavailable,
		MaxSurge:       input.MaxSurge,
	}

	return deploymentStrategy, nil
}

func (r *Resolver) CreateServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := model.ServiceSpec{
		Name:                   args.ServiceSpec.Name,
		CpuRequest:             args.ServiceSpec.CpuRequest,
		CpuLimit:               args.ServiceSpec.CpuLimit,
		MemoryRequest:          args.ServiceSpec.MemoryRequest,
		MemoryLimit:            args.ServiceSpec.MemoryLimit,
		TerminationGracePeriod: args.ServiceSpec.TerminationGracePeriod,
	}

	r.DB.Create(&serviceSpec)

	return &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}}, nil
}

func (r *Resolver) UpdateServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := model.ServiceSpec{}

	serviceSpecID, err := uuid.FromString(*args.ServiceSpec.ID)
	if err != nil {
		return nil, fmt.Errorf("UpdateServiceSpec: Missing argument id")
	}

	if r.DB.Where("id=?", serviceSpecID).Find(&serviceSpec).RecordNotFound() {
		return nil, fmt.Errorf("ServiceSpec not found with given argument id")
	}

	serviceSpec.Name = args.ServiceSpec.Name
	serviceSpec.CpuLimit = args.ServiceSpec.CpuLimit
	serviceSpec.CpuRequest = args.ServiceSpec.CpuRequest
	serviceSpec.MemoryLimit = args.ServiceSpec.MemoryLimit
	serviceSpec.MemoryRequest = args.ServiceSpec.MemoryRequest
	serviceSpec.TerminationGracePeriod = args.ServiceSpec.TerminationGracePeriod

	r.DB.Save(&serviceSpec)

	//r.ServiceSpecUpdated(&serviceSpec)

	return &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}}, nil
}

func (r *Resolver) DeleteServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := model.ServiceSpec{}
	if r.DB.Where("id=?", args.ServiceSpec.ID).Find(&serviceSpec).RecordNotFound() {
		return nil, fmt.Errorf("ServiceSpec not found with given argument id")
	} else {
		services := []model.Service{}
		r.DB.Where("service_spec_id = ?", serviceSpec.Model.ID).Find(&services)
		if len(services) == 0 {
			r.DB.Delete(&serviceSpec)

			//r.ServiceSpecDeleted(&serviceSpec)

			return &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}}, nil
		} else {
			return nil, fmt.Errorf("Delete all project-services using this service spec first.")
		}
	}
}

func (r *Resolver) CreateEnvironment(ctx context.Context, args *struct{ Environment *model.EnvironmentInput }) (*EnvironmentResolver, error) {
	mut := EnvironmentResolverMutation{r.DB}
	return mut.CreateEnvironment(ctx, args)
}

func (r *Resolver) UpdateEnvironment(ctx context.Context, args *struct{ Environment *model.EnvironmentInput }) (*EnvironmentResolver, error) {
	mut := EnvironmentResolverMutation{r.DB}
	return mut.UpdateEnvironment(ctx, args)
}

func (r *Resolver) DeleteEnvironment(ctx context.Context, args *struct{ Environment *model.EnvironmentInput }) (*EnvironmentResolver, error) {
	mut := EnvironmentResolverMutation{r.DB}
	return mut.DeleteEnvironment(ctx, args)
}

func (r *Resolver) CreateSecret(ctx context.Context, args *struct{ Secret *model.SecretInput }) (*SecretResolver, error) {

	projectID := uuid.UUID{}
	var environmentID uuid.UUID
	var secretScope model.SecretScope

	if args.Secret.ProjectID != nil {
		// Check if project can create secret
		if r.DB.Where("environment_id = ? and project_id = ?", args.Secret.EnvironmentID, args.Secret.ProjectID).Find(&model.ProjectEnvironment{}).RecordNotFound() {
			return nil, errors.New("Project not allowed to create secret in given environment")
		}

		projectID = uuid.FromStringOrNil(*args.Secret.ProjectID)
	}

	secretScope = GetSecretScope(args.Secret.Scope)
	if secretScope == model.SecretScope("unknown") {
		return nil, fmt.Errorf("Invalid env var scope.")
	}

	environmentID, err := uuid.FromString(args.Secret.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse environmentID. Invalid format.")
	}

	userIDString, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return &SecretResolver{}, err
	}

	userID, err := uuid.FromString(userIDString)
	if err != nil {
		return &SecretResolver{}, err
	}

	var existingEnvVar model.Secret

	if r.DB.Where("key = ? and project_id = ? and deleted_at is null and environment_id = ? and type = ?", args.Secret.Key, projectID, environmentID, args.Secret.Type).Find(&existingEnvVar).RecordNotFound() {
		secret := model.Secret{
			Key:           args.Secret.Key,
			ProjectID:     projectID,
			Type:          plugins.GetType(args.Secret.Type),
			Scope:         secretScope,
			EnvironmentID: environmentID,
			IsSecret:      args.Secret.IsSecret,
		}
		r.DB.Create(&secret)

		secretValue := model.SecretValue{
			SecretID: secret.Model.ID,
			Value:    args.Secret.Value,
			UserID:   userID,
		}
		r.DB.Create(&secretValue)

		//r.SecretCreated(&secret)

		return &SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: r.DB, Secret: secret, SecretValue: secretValue}}, nil
	} else {
		return nil, fmt.Errorf("CreateSecret: key already exists")
	}

}

func (r *Resolver) UpdateSecret(ctx context.Context, args *struct{ Secret *model.SecretInput }) (*SecretResolver, error) {
	var secret model.Secret

	userIDString, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return &SecretResolver{}, err
	}

	userID, err := uuid.FromString(userIDString)
	if err != nil {
		return &SecretResolver{}, err
	}

	if r.DB.Where("id = ?", args.Secret.ID).Find(&secret).RecordNotFound() {
		return nil, fmt.Errorf("UpdateSecret: env var doesn't exist.")
	} else {
		secretValue := model.SecretValue{
			SecretID: secret.Model.ID,
			Value:    args.Secret.Value,
			UserID:   userID,
		}
		r.DB.Create(&secretValue)

		//r.SecretUpdated(&secret)

		return &SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: r.DB, Secret: secret, SecretValue: secretValue}}, nil
	}
}

func (r *Resolver) DeleteSecret(ctx context.Context, args *struct{ Secret *model.SecretInput }) (*SecretResolver, error) {
	var secret model.Secret

	if r.DB.Where("id = ?", args.Secret.ID).Find(&secret).RecordNotFound() {
		return nil, fmt.Errorf("DeleteSecret: key doesn't exist.")
	} else {
		// check if any configs are using the secret
		extensions := []model.Extension{}
		where := fmt.Sprintf(`config @> '{"config": [{"value": "%s"}]}'`, secret.Model.ID.String())
		r.DB.Where(where).Find(&extensions)
		if len(extensions) == 0 {
			versions := []model.SecretValue{}

			r.DB.Delete(&secret)
			r.DB.Where("secret_id = ?", secret.Model.ID).Delete(&versions)

			//r.SecretDeleted(&secret)

			return &SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: r.DB, Secret: secret}}, nil
		} else {
			return nil, fmt.Errorf("Remove Config values from Extensions where Secret is used before deleting.")
		}
	}
}

func (r *Resolver) CreateExtension(args *struct{ Extension *model.ExtensionInput }) (*ExtensionResolver, error) {
	mut := ExtensionResolverMutation{r.DB}
	return mut.CreateExtension(args)
}

func (r *Resolver) UpdateExtension(args *struct{ Extension *model.ExtensionInput }) (*ExtensionResolver, error) {
	mut := ExtensionResolverMutation{r.DB}
	return mut.UpdateExtension(args)	
}

func (r *Resolver) DeleteExtension(args *struct{ Extension *model.ExtensionInput }) (*ExtensionResolver, error) {
	mut := ExtensionResolverMutation{r.DB}
	return mut.DeleteExtension(args)
}

func (r *Resolver) CreateProjectExtension(ctx context.Context, args *struct{ ProjectExtension *model.ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var projectExtension model.ProjectExtension

	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	// Check if project can create project extension in environment
	if err := r.DB.Where("environment_id = ? and project_id = ?", args.ProjectExtension.EnvironmentID, args.ProjectExtension.ProjectID).Find(&model.ProjectEnvironment{}).Error; err != nil {
		return nil, errors.New("Project not allowed to install extensions in given environment")
	}

	extension := model.Extension{}
	if err := r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).Error; err != nil {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return nil, fmt.Errorf("No extension found for id: '%s'", args.ProjectExtension.ExtensionID)
	}

	project := model.Project{}
	if err := r.DB.Where("id = ?", args.ProjectExtension.ProjectID).Find(&project).Error; err != nil {
		log.InfoWithFields("no project found", log.Fields{
			"id": args.ProjectExtension.ProjectID,
		})
		return nil, fmt.Errorf("No project found: '%s'", args.ProjectExtension.ProjectID)
	}

	env := model.Environment{}
	if err := r.DB.Where("id = ?", args.ProjectExtension.EnvironmentID).Find(&env).Error; err != nil {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.ProjectExtension.EnvironmentID,
		})
		return nil, fmt.Errorf("No environment found: '%s'", args.ProjectExtension.ProjectID)
	}

	// check if extension already exists with project
	// ignore if the extension type is 'once' (installable many times)
	if extension.Type == plugins.GetType("once") || extension.Type == plugins.GetType("notification") || r.DB.Where("project_id = ? and extension_id = ? and environment_id = ?", args.ProjectExtension.ProjectID, args.ProjectExtension.ExtensionID, args.ProjectExtension.EnvironmentID).Find(&projectExtension).RecordNotFound() {
		if extension.Key == "route53" {
			err := r.handleExtensionRoute53(args, &projectExtension)
			if err != nil {
				return &ProjectExtensionResolver{}, err
			}
		}

		projectExtension = model.ProjectExtension{
			State:         transistor.GetState("waiting"),
			ExtensionID:   extension.Model.ID,
			ProjectID:     project.Model.ID,
			EnvironmentID: env.Model.ID,
			Config:        postgres.Jsonb{[]byte(args.ProjectExtension.Config.RawMessage)},
			CustomConfig:  postgres.Jsonb{[]byte(args.ProjectExtension.CustomConfig.RawMessage)},
		}

		r.DB.Save(&projectExtension)

		artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}

		projectExtensionEvent := plugins.ProjectExtension{
			ID: projectExtension.Model.ID.String(),
			Project: plugins.Project{
				ID:         project.Model.ID.String(),
				Slug:       project.Slug,
				Repository: project.Repository,
			},
			Environment: env.Key,
		}
		ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("project:%s", extension.Key)), transistor.GetAction("create"), projectExtensionEvent)
		ev.Artifacts = artifacts
		r.Events <- ev

		return &ProjectExtensionResolver{DBProjectExtensionResolver: &db_resolver.ProjectExtensionResolver{DB: r.DB, ProjectExtension: projectExtension}}, nil
	}

	return nil, errors.New("This extension is already installed in this project.")
}

func (r *Resolver) UpdateProjectExtension(args *struct{ ProjectExtension *model.ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var projectExtension model.ProjectExtension

	if r.DB.Where("id = ?", args.ProjectExtension.ID).First(&projectExtension).RecordNotFound() {
		log.InfoWithFields("no project extension found", log.Fields{
			"extension": args.ProjectExtension,
		})
		return nil, fmt.Errorf("No project extension found")
	}

	extension := model.Extension{}
	if r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return nil, fmt.Errorf("No extension found.")
	}

	project := model.Project{}
	if r.DB.Where("id = ?", args.ProjectExtension.ProjectID).Find(&project).RecordNotFound() {
		log.InfoWithFields("no project found", log.Fields{
			"id": args.ProjectExtension.ProjectID,
		})
		return nil, fmt.Errorf("No project found.")
	}

	env := model.Environment{}
	if r.DB.Where("id = ?", args.ProjectExtension.EnvironmentID).Find(&env).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.ProjectExtension.EnvironmentID,
		})
		return nil, fmt.Errorf("No environment found.")
	}

	if extension.Key == "route53" {
		err := r.handleExtensionRoute53(args, &projectExtension)
		if err != nil {
			return nil, err
		}
	}

	projectExtension.Config = postgres.Jsonb{args.ProjectExtension.Config.RawMessage}
	projectExtension.CustomConfig = postgres.Jsonb{args.ProjectExtension.CustomConfig.RawMessage}
	projectExtension.State = transistor.GetState("waiting")
	projectExtension.StateMessage = ""

	r.DB.Save(&projectExtension)

	artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	projectExtensionEvent := plugins.ProjectExtension{
		ID: projectExtension.Model.ID.String(),
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Environment: env.Key,
	}

	ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("project:%s", extension.Key)), transistor.GetAction("update"), projectExtensionEvent)
	ev.Artifacts = artifacts

	r.Events <- ev

	return &ProjectExtensionResolver{DBProjectExtensionResolver: &db_resolver.ProjectExtensionResolver{DB: r.DB, ProjectExtension: projectExtension}}, nil
}

func (r *Resolver) DeleteProjectExtension(args *struct{ ProjectExtension *model.ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var projectExtension model.ProjectExtension

	if r.DB.Where("id = ?", args.ProjectExtension.ID).First(&projectExtension).RecordNotFound() {
		log.InfoWithFields("no project extension found", log.Fields{
			"extension": args.ProjectExtension,
		})
		return nil, fmt.Errorf("No Project Extension Found")
	}

	extension := model.Extension{}
	if r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return nil, errors.New("No extension found.")
	}

	project := model.Project{}
	if r.DB.Where("id = ?", args.ProjectExtension.ProjectID).Find(&project).RecordNotFound() {
		log.InfoWithFields("no project found", log.Fields{
			"id": args.ProjectExtension.ProjectID,
		})
		return nil, errors.New("No project found.")
	}

	env := model.Environment{}
	if r.DB.Where("id = ?", args.ProjectExtension.EnvironmentID).Find(&env).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.ProjectExtension.EnvironmentID,
		})
		return nil, errors.New("No environment found.")
	}

	// ADB
	// Removed logic here that would delete all existing release extensions associated with this project extension
	// However, that's not really what we want. Doing this means we lose a part of our release history
	// What we really want is to just delete the project extension from future releases and leave the history unaffected

	r.DB.Delete(&projectExtension)

	artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	projectExtensionEvent := plugins.ProjectExtension{
		ID: projectExtension.Model.ID.String(),
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Environment: env.Key,
	}
	ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("project:%s", extension.Key)), transistor.GetAction("delete"), projectExtensionEvent)
	ev.Artifacts = artifacts
	r.Events <- ev

	return &ProjectExtensionResolver{DBProjectExtensionResolver: &db_resolver.ProjectExtensionResolver{DB: r.DB, ProjectExtension: projectExtension}}, nil
}

// UpdateUserPermissions
func (r *Resolver) UpdateUserPermissions(ctx context.Context, args *struct{ UserPermissions *model.UserPermissionsInput }) ([]string, error) {
	var err error
	var results []string

	if r.DB.Where("id = ?", args.UserPermissions.UserID).Find(&model.User{}).RecordNotFound() {
		return nil, errors.New("User not found")
	}

	for _, permission := range args.UserPermissions.Permissions {
		if _, err = auth.CheckAuth(ctx, []string{permission.Value}); err != nil {
			return nil, err
		}
	}

	for _, permission := range args.UserPermissions.Permissions {
		if permission.Grant == true {
			userPermission := model.UserPermission{
				UserID: uuid.FromStringOrNil(args.UserPermissions.UserID),
				Value:  permission.Value,
			}
			r.DB.Where(userPermission).FirstOrCreate(&userPermission)
			results = append(results, permission.Value)
		} else {
			r.DB.Where("user_id = ? AND value = ?", args.UserPermissions.UserID, permission.Value).Delete(&model.UserPermission{})
		}
	}

	return results, nil
}

// UpdateProjectEnvironments
func (r *Resolver) UpdateProjectEnvironments(ctx context.Context, args *struct {
	ProjectEnvironments *model.ProjectEnvironmentsInput
}) ([]*EnvironmentResolver, error) {
	var results []*EnvironmentResolver

	project := model.Project{}
	if r.DB.Where("id = ?", args.ProjectEnvironments.ProjectID).Find(&project).RecordNotFound() {
		return nil, errors.New("No project found with inputted projectID")
	}

	for _, permission := range args.ProjectEnvironments.Permissions {
		// Check if environment object exists
		environment := model.Environment{}
		if r.DB.Where("id = ?", permission.EnvironmentID).Find(&environment).RecordNotFound() {
			return nil, errors.New(fmt.Sprintf("No environment found for environmentID %s", permission.EnvironmentID))
		}

		if permission.Grant {
			// Grant permission by adding ProjectEnvironment row
			projectEnvironment := model.ProjectEnvironment{
				EnvironmentID: environment.Model.ID,
				ProjectID:     project.Model.ID,
			}
			r.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID, project.Model.ID).FirstOrCreate(&projectEnvironment)
			results = append(results, &EnvironmentResolver{DBEnvironmentResolver: &db_resolver.EnvironmentResolver{DB: r.DB, Environment: environment}})
		} else {
			r.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID, project.Model.ID).Delete(&model.ProjectEnvironment{})
		}
	}

	return results, nil
}

// GetGitCommits
func (r *Resolver) GetGitCommits(ctx context.Context, args *struct {
	ProjectID     graphql.ID
	EnvironmentID graphql.ID
	New           *bool
}) (bool, error) {
	if args.New != nil && *args.New {
		var err error
		project := model.Project{}
		env := model.Environment{}
		projectSettings := model.ProjectSettings{}
		latestFeature := model.Feature{}
		hash := ""

		if err = r.DB.Where("id = ?", args.ProjectID).First(&project).Error; err != nil {
			return false, err
		}

		if err = r.DB.Where("id = ?", args.EnvironmentID).First(&env).Error; err != nil {
			return false, err
		}

		if err = r.DB.Where("project_id = ? AND environment_id = ?", project.Model.ID, env.Model.ID).First(&projectSettings).Error; err != nil {
			return false, err
		}

		if err = r.DB.Where("project_id = ?", project.Model.ID).Order("created_at DESC").First(&latestFeature).Error; err == nil {
			hash = latestFeature.Hash
		}

		payload := plugins.GitSync{
			Project: plugins.Project{
				ID:         project.Model.ID.String(),
				Repository: project.Repository,
			},
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        projectSettings.GitBranch,
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			},
			From: hash,
		}

		r.Events <- transistor.NewEvent(plugins.GetEventName("gitsync"), transistor.GetAction("create"), payload)
		return true, nil
	}
	return true, nil
}

func (r *Resolver) BookmarkProject(ctx context.Context, args *struct{ ID graphql.ID }) (bool, error) {
	var projectBookmark model.ProjectBookmark

	_userID, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return false, err
	}

	userID, err := uuid.FromString(_userID)
	if err != nil {
		return false, err
	}

	projectID, err := uuid.FromString(string(args.ID))
	if err != nil {
		return false, err
	}

	if r.DB.Where("user_id = ? AND project_id = ?", userID, projectID).First(&projectBookmark).RecordNotFound() {
		projectBookmark = model.ProjectBookmark{
			UserID:    userID,
			ProjectID: projectID,
		}
		r.DB.Save(&projectBookmark)
		return true, nil
	} else {
		r.DB.Delete(&projectBookmark)
		return false, nil
	}
}

/* fills in Config by querying config ids and getting the actual value */
func ExtractArtifacts(projectExtension model.ProjectExtension, extension model.Extension, db *gorm.DB) ([]transistor.Artifact, error) {
	var artifacts []transistor.Artifact
	var err error

	extensionConfig := []model.ExtConfig{}
	if extension.Config.RawMessage != nil {
		err = json.Unmarshal(extension.Config.RawMessage, &extensionConfig)
		if err != nil {
			return nil, err
		}
	}

	projectConfig := []model.ExtConfig{}
	if projectExtension.Config.RawMessage != nil {
		err = json.Unmarshal(projectExtension.Config.RawMessage, &projectConfig)
		if err != nil {
			return nil, err
		}
	}

	existingArtifacts := []transistor.Artifact{}
	if projectExtension.Artifacts.RawMessage != nil {
		err = json.Unmarshal(projectExtension.Artifacts.RawMessage, &existingArtifacts)
		if err != nil {
			return nil, err
		}
	}

	for i, ec := range extensionConfig {
		for _, pc := range projectConfig {
			if ec.AllowOverride && ec.Key == pc.Key && pc.Value != "" {
				extensionConfig[i].Value = pc.Value
			}
		}

		var artifact transistor.Artifact
		// check if val is UUID. If so, query in environment variables for id
		secretID := uuid.FromStringOrNil(extensionConfig[i].Value)
		if secretID != uuid.Nil {
			secret := model.SecretValue{}
			if db.Where("secret_id = ?", secretID).Order("created_at desc").First(&secret).RecordNotFound() {
				log.InfoWithFields("secret not found", log.Fields{
					"secret_id": secretID,
				})
			}
			artifact.Key = ec.Key
			artifact.Value = secret.Value
		} else {
			artifact.Key = ec.Key
			artifact.Value = extensionConfig[i].Value
		}
		artifacts = append(artifacts, artifact)
	}

	for _, ea := range existingArtifacts {
		artifacts = append(artifacts, ea)
	}

	projectCustomConfig := make(map[string]interface{})
	if projectExtension.CustomConfig.RawMessage != nil {
		err = json.Unmarshal(projectExtension.CustomConfig.RawMessage, &projectCustomConfig)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	for key, val := range projectCustomConfig {
		var artifact transistor.Artifact
		artifact.Key = key
		artifact.Value = val
		artifact.Secret = false
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}
