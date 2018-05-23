package codeamp_resolvers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
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
	Project *ProjectInput
}) (*ProjectResolver, error) {
	var project Project

	protocol := "HTTPS"
	switch args.Project.GitProtocol {
	case "private", "PRIVATE", "ssh", "SSH":
		protocol = "SSH"
	case "public", "PUBLIC", "https", "HTTPS":
		protocol = "HTTPS"
	}

	// Check if project already exists with same name
	existingProject := Project{}
	res := plugins.GetRegexParams("(?P<host>(git@|https?:\\/\\/)([\\w\\.@]+)(\\/|:))(?P<owner>[\\w,\\-,\\_]+)\\/(?P<repo>[\\w,\\-,\\_]+)(.git){0,1}((\\/){0,1})", args.Project.GitUrl)
	repository := fmt.Sprintf("%s/%s", res["owner"], res["repo"])
	if r.DB.Where("repository = ?", repository).First(&existingProject).RecordNotFound() {
		log.InfoWithFields("[+] Project not found", log.Fields{
			"repository": repository,
		})
	} else {
		return nil, fmt.Errorf("This repository already exists. Try again with a different git url.")
	}

	project = Project{
		GitProtocol: protocol,
		GitUrl:      args.Project.GitUrl,
		Secret:      transistor.RandomString(30),
	}
	project.Name = repository
	project.Repository = repository
	project.Slug = slug.Slug(repository)

	deletedProject := Project{}
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
	environments := []Environment{}
	if r.DB.Find(&environments).RecordNotFound() {
		log.InfoWithFields("Environment doesn't exist.", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("No environments initialized.")
	}

	for _, env := range environments {
		r.DB.Create(&ProjectSettings{
			EnvironmentID: env.Model.ID,
			ProjectID:     project.Model.ID,
			GitBranch:     "master",
		})
		// Create ProjectEnvironment rows for default envs
		if env.IsDefault {
			r.DB.Create(&ProjectEnvironment{
				EnvironmentID: env.Model.ID,
				ProjectID:     project.Model.ID,
			})
		}
	}

	if userId, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	} else {
		// Create user permission for project
		userPermission := UserPermission{
			UserID: uuid.FromStringOrNil(userId),
			Value:  fmt.Sprintf("projects/%s", project.Repository),
		}
		r.DB.Create(&userPermission)
	}

	return &ProjectResolver{DB: r.DB, Project: project}, nil
}

// UpdateProject Update project
func (r *Resolver) UpdateProject(args *struct {
	Project *ProjectInput
}) (*ProjectResolver, error) {
	var project Project

	if args.Project.ID == nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	if r.DB.Where("id = ?", args.Project.ID).First(&project).RecordNotFound() {
		log.InfoWithFields("Project not found", log.Fields{
			"id": args.Project.ID,
		})
		return nil, fmt.Errorf("Project not found.")
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
			return &ProjectResolver{}, fmt.Errorf("Couldn't parse project ID")
		}

		environmentID, err := uuid.FromString(*args.Project.EnvironmentID)
		if err != nil {
			return &ProjectResolver{}, fmt.Errorf("Couldn't parse environment ID")
		}

		var projectSettings ProjectSettings
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

	return &ProjectResolver{DB: r.DB, Project: project}, nil
}

// StopRelease
func (r *Resolver) StopRelease(ctx context.Context, args *struct{ ID graphql.ID }) (*ReleaseResolver, error) {
	userID, err := CheckAuth(ctx, []string{})
	if err != nil {
		return &ReleaseResolver{}, err
	}

	var user User

	r.DB.Where("id = ?", userID).Find(&user)

	var release Release
	var releaseExtensions []ReleaseExtension

	r.DB.Where("release_id = ?", args.ID).Find(&releaseExtensions)
	if len(releaseExtensions) < 1 {
		log.Info("No release extensions found for release")
	}

	if r.DB.Where("id = ?", args.ID).Find(&release).RecordNotFound() {
		log.InfoWithFields("Release not found", log.Fields{
			"id": args.ID,
		})

		return nil, errors.New("Release Not Found")
	}

	release.State = plugins.GetState("failed")
	release.StateMessage = fmt.Sprintf("Release stopped by %s", user.Email)
	r.DB.Save(&release)

	for _, releaseExtension := range releaseExtensions {
		var projectExtension ProjectExtension
		if r.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
			log.InfoWithFields("Associated project extension not found", log.Fields{
				"id": args.ID,
				"release_extension_id": releaseExtension.ID,
				"project_extension_id": releaseExtension.ProjectExtensionID,
			})

			return nil, errors.New("Project Extension Not Found")
		}

		// find associated ProjectExtension Extension
		var extension Extension
		if r.DB.Where("id = ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.InfoWithFields("Associated extension not found", log.Fields{
				"id": args.ID,
				"release_extension_id": releaseExtension.ID,
				"project_extension_id": releaseExtension.ProjectExtensionID,
				"extension_id":         projectExtension.ExtensionID,
			})

			return nil, errors.New("Extension Not Found")
		}

		if releaseExtension.State == plugins.GetState("waiting") {
			releaseExtensionEvent := plugins.ReleaseExtension{
				ID:      releaseExtension.ID.String(),
				Slug:    extension.Key,
				Project: plugins.Project{},
				Release: plugins.Release{
					ID: releaseExtension.ReleaseID.String(),
				},
				Environment: "",
			}
			event := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), plugins.GetAction("create"), releaseExtensionEvent)
			event.State = plugins.GetState("failed")
			event.StateMessage = fmt.Sprintf("Deployment Stopped By User %s", user.Email)
			r.Events <- event
		}
	}

	return &ReleaseResolver{DB: r.DB, Release: release}, nil
}

// CreateRelease
func (r *Resolver) CreateRelease(ctx context.Context, args *struct{ Release *ReleaseInput }) (*ReleaseResolver, error) {
	var project Project
	var secrets []Secret
	var services []Service
	var projectExtensions []ProjectExtension
	var secretsJsonb postgres.Jsonb
	var servicesJsonb postgres.Jsonb
	var projectExtensionsJsonb postgres.Jsonb

	// Check if project can create release in environment
	if r.DB.Where("environment_id = ? and project_id = ?", args.Release.EnvironmentID, args.Release.ProjectID).Find(&ProjectEnvironment{}).RecordNotFound() {
		return nil, errors.New("Project not allowed to create release in given environment")
	}

	if args.Release.ID == nil {
		projectSecrets := []Secret{}
		// get all the env vars related to this release and store
		r.DB.Where("environment_id = ? AND project_id = ? AND scope = ?", args.Release.EnvironmentID, args.Release.ProjectID, "project").Find(&projectSecrets)
		for _, secret := range projectSecrets {
			var secretValue SecretValue
			r.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
			secret.Value = secretValue
			secrets = append(secrets, secret)
		}

		globalSecrets := []Secret{}
		r.DB.Where("environment_id = ? AND scope = ?", args.Release.EnvironmentID, "global").Find(&globalSecrets)
		for _, secret := range globalSecrets {
			var secretValue SecretValue
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
			ports := []ServicePort{}
			r.DB.Where("service_id = ?", service.Model.ID).Find(&ports)
			services[i].Ports = ports
		}

		servicesMarshaled, err := json.Marshal(services)
		if err != nil {
			return &ReleaseResolver{}, err
		}

		servicesJsonb = postgres.Jsonb{servicesMarshaled}
		// check if any project extensions that are not 'once' exists
		r.DB.Where("project_id = ? AND environment_id = ? AND state = ?", args.Release.ProjectID, args.Release.EnvironmentID, plugins.GetState("complete")).Find(&projectExtensions)

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
		release := Release{}

		if r.DB.Where("id = ?", string(*args.Release.ID)).Find(&release).RecordNotFound() {
			log.ErrorWithFields("Could not find release", log.Fields{
				"id": *args.Release.ID,
			})
			return &ReleaseResolver{}, errors.New("Release not found")
		}

		secretsJsonb = release.Secrets
		servicesJsonb = release.Services
		projectExtensionsJsonb = release.ProjectExtensions

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

	currentReleaseHeadFeature := Feature{}

	r.DB.Where("id = ?", args.Release.HeadFeatureID).First(&currentReleaseHeadFeature)

	waitingRelease := Release{}

	r.DB.Where("state in (?) and project_id = ? and environment_id = ?", []string{string(plugins.GetState("waiting")),
		string(plugins.GetState("running"))}, args.Release.ProjectID, args.Release.EnvironmentID).Order("created_at desc").First(&waitingRelease)

	wrSecretsSha1 := sha1.New()
	wrSecretsSha1.Write(waitingRelease.Services.RawMessage)
	waitingReleaseSecretsSig := wrSecretsSha1.Sum(nil)

	wrServicesSha1 := sha1.New()
	wrServicesSha1.Write(waitingRelease.Services.RawMessage)
	waitingReleaseServicesSig := wrServicesSha1.Sum(nil)

	waitingReleaseHeadFeature := Feature{}

	r.DB.Where("id = ?", waitingRelease.HeadFeatureID).First(&waitingReleaseHeadFeature)

	if fmt.Sprintf("%x", secretsSig) == fmt.Sprintf("%x", waitingReleaseSecretsSig) && fmt.Sprintf("%x", servicesSig) == fmt.Sprintf("%x", waitingReleaseServicesSig) && currentReleaseHeadFeature.Hash == waitingReleaseHeadFeature.Hash {
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
	currentRelease := Release{}
	tailFeatureID := headFeatureID
	if r.DB.Where("state = ? and project_id = ? and environment_id = ?", plugins.GetState("complete"), projectID, environmentID).Find(&currentRelease).Order("created_at desc").Limit(1).RecordNotFound() {
	} else {
		tailFeatureID = currentRelease.HeadFeatureID
	}

	userID, err := CheckAuth(ctx, []string{})
	if err != nil {
		return &ReleaseResolver{}, err
	}

	// Create Release
	release := Release{
		ProjectID:         projectID,
		EnvironmentID:     environmentID,
		UserID:            uuid.FromStringOrNil(userID),
		HeadFeatureID:     headFeatureID,
		TailFeatureID:     tailFeatureID,
		Secrets:           secretsJsonb,
		Services:          servicesJsonb,
		ProjectExtensions: projectExtensionsJsonb,
		ForceRebuild:      args.Release.ForceRebuild,
	}

	r.DB.Create(&release)

	if r.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": release.ProjectID,
		})
		return &ReleaseResolver{}, errors.New("Project not found")
	}

	// get all branches relevant for the project
	var branch string
	var projectSettings ProjectSettings

	if r.DB.Where("environment_id = ? and project_id = ?", release.EnvironmentID, release.ProjectID).First(&projectSettings).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	var environment Environment
	if r.DB.Where("id = ?", release.EnvironmentID).Find(&environment).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": release.EnvironmentID,
		})
		return &ReleaseResolver{}, errors.New("Environment not found")
	}

	var headFeature Feature
	if r.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return &ReleaseResolver{}, errors.New("head feature not found")
	}

	var tailFeature Feature
	if r.DB.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return &ReleaseResolver{}, errors.New("Tail feature not found")
	}

	var pluginServices []plugins.Service
	for _, service := range services {
		var spec ServiceSpec
		if r.DB.Where("id = ?", service.ServiceSpecID).First(&spec).RecordNotFound() {
			log.InfoWithFields("servicespec not found", log.Fields{
				"id": service.ServiceSpecID,
			})
			return &ReleaseResolver{}, errors.New("ServiceSpec not found")
		}

		count, _ := strconv.ParseInt(service.Count, 10, 64)
		terminationGracePeriod, _ := strconv.ParseInt(spec.TerminationGracePeriod, 10, 64)

		listeners := []plugins.Listener{}
		for _, l := range service.Ports {
			p, err := strconv.ParseInt(l.Port, 10, 32)
			if err != nil {
				panic(err)
			}
			listener := plugins.Listener{
				Port:     int32(p),
				Protocol: l.Protocol,
			}
			listeners = append(listeners, listener)
		}

		pluginServices = append(pluginServices, plugins.Service{
			ID:        service.Model.ID.String(),
			Name:      service.Name,
			Command:   service.Command,
			Listeners: listeners,
			Replicas:  count,
			Spec: plugins.ServiceSpec{
				ID:                            spec.Model.ID.String(),
				CpuRequest:                    fmt.Sprintf("%sm", spec.CpuRequest),
				CpuLimit:                      fmt.Sprintf("%sm", spec.CpuLimit),
				MemoryRequest:                 fmt.Sprintf("%sMi", spec.MemoryRequest),
				MemoryLimit:                   fmt.Sprintf("%sMi", spec.MemoryLimit),
				TerminationGracePeriodSeconds: terminationGracePeriod,
			},
			Type: string(service.Type),
		})
	}

	var pluginSecrets []plugins.Secret
	for _, secret := range secrets {
		pluginSecrets = append(pluginSecrets, plugins.Secret{
			Key:   secret.Key,
			Value: secret.Value.Value,
			Type:  secret.Type,
		})
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
		Secrets: pluginSecrets,
	}

	// Create/Emit Release ProjectExtensions
	for _, projectExtension := range projectExtensions {
		extension := Extension{}
		if r.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.ErrorWithFields("extension spec not found", log.Fields{
				"id": projectExtension.ExtensionID,
			})
			return &ReleaseResolver{}, errors.New("extension spec not found")
		}

		if plugins.Type(extension.Type) == plugins.GetType("workflow") || plugins.Type(extension.Type) == plugins.GetType("deployment") {
			var headFeature Feature
			if r.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
				log.ErrorWithFields("head feature not found", log.Fields{
					"id": release.HeadFeatureID,
				})
				return &ReleaseResolver{}, errors.New("head feature not found")
			}

			// create ReleaseExtension
			releaseExtension := ReleaseExtension{
				ReleaseID:          release.Model.ID,
				FeatureHash:        headFeature.Hash,
				ServicesSignature:  fmt.Sprintf("%x", servicesSig),
				SecretsSignature:   fmt.Sprintf("%x", secretsSig),
				ProjectExtensionID: projectExtension.Model.ID,
				Type:               extension.Type,
				State:              plugins.GetState("waiting"),
			}

			r.DB.Create(&releaseExtension)
		}
	}

	if waitingRelease.State != "" {
		log.Info(fmt.Sprintf("Release is already running, queueing %s", release.Model.ID.String()))
		return &ReleaseResolver{}, fmt.Errorf("Release is already running, queuing %s", release.Model.ID.String())
	} else {
		r.Events <- transistor.NewEvent(transistor.EventName("release"), plugins.GetAction("create"), releaseEvent)

		return &ReleaseResolver{DB: r.DB, Release: Release{}}, nil
	}
}

// CreateService Create service
func (r *Resolver) CreateService(args *struct{ Service *ServiceInput }) (*ServiceResolver, error) {
	// Check if project can create service in environment
	if r.DB.Where("environment_id = ? and project_id = ?", args.Service.EnvironmentID, args.Service.ProjectID).Find(&ProjectEnvironment{}).RecordNotFound() {
		return nil, errors.New("Project not allowed to create service in given environment")
	}

	projectID, err := uuid.FromString(args.Service.ProjectID)
	if err != nil {
		return &ServiceResolver{}, err
	}

	environmentID, err := uuid.FromString(args.Service.EnvironmentID)
	if err != nil {
		return &ServiceResolver{}, err
	}

	serviceSpecID, err := uuid.FromString(args.Service.ServiceSpecID)
	if err != nil {
		return &ServiceResolver{}, err
	}

	service := Service{
		Name:          args.Service.Name,
		Command:       args.Service.Command,
		ServiceSpecID: serviceSpecID,
		Type:          plugins.Type(args.Service.Type),
		Count:         args.Service.Count,
		ProjectID:     projectID,
		EnvironmentID: environmentID,
	}

	r.DB.Create(&service)

	if args.Service.Ports != nil {
		for _, cp := range *args.Service.Ports {
			servicePort := ServicePort{
				ServiceID: service.ID,
				Port:      cp.Port,
				Protocol:  cp.Protocol,
			}
			r.DB.Create(&servicePort)
		}
	}

	//r.ServiceCreated(&service)

	return &ServiceResolver{DB: r.DB, Service: service}, nil
}

// UpdateService Update Service
func (r *Resolver) UpdateService(args *struct{ Service *ServiceInput }) (*ServiceResolver, error) {
	serviceID := uuid.FromStringOrNil(*args.Service.ID)
	serviceSpecID := uuid.FromStringOrNil(args.Service.ServiceSpecID)

	if serviceID == uuid.Nil || serviceSpecID == uuid.Nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	var service Service
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
	var servicePorts []ServicePort
	r.DB.Where("service_id = ?", serviceID).Find(&servicePorts)

	// delete all container ports
	// replace with current

	for _, cp := range servicePorts {
		r.DB.Delete(&cp)
	}

	if args.Service.Ports != nil {
		for _, cp := range *args.Service.Ports {
			servicePort := ServicePort{
				ServiceID: service.ID,
				Port:      cp.Port,
				Protocol:  cp.Protocol,
			}
			r.DB.Create(&servicePort)
		}
	}

	//r.ServiceUpdated(&service)

	return &ServiceResolver{DB: r.DB, Service: service}, nil
}

// DeleteService Delete service
func (r *Resolver) DeleteService(args *struct{ Service *ServiceInput }) (*ServiceResolver, error) {
	serviceID, err := uuid.FromString(*args.Service.ID)

	if err != nil {
		return &ServiceResolver{}, err
	}

	var service Service

	r.DB.Where("id = ?", serviceID).Find(&service)
	r.DB.Delete(&service)

	// delete all previous container ports
	var servicePorts []ServicePort
	r.DB.Where("service_id = ?", serviceID).Find(&servicePorts)

	// delete all container ports
	// replace with current
	for _, cp := range servicePorts {
		r.DB.Delete(&cp)
	}

	//r.ServiceDeleted(&service)

	return &ServiceResolver{DB: r.DB, Service: service}, nil
}

func (r *Resolver) CreateServiceSpec(args *struct{ ServiceSpec *ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := ServiceSpec{
		Name:                   args.ServiceSpec.Name,
		CpuRequest:             args.ServiceSpec.CpuRequest,
		CpuLimit:               args.ServiceSpec.CpuLimit,
		MemoryRequest:          args.ServiceSpec.MemoryRequest,
		MemoryLimit:            args.ServiceSpec.MemoryLimit,
		TerminationGracePeriod: args.ServiceSpec.TerminationGracePeriod,
	}

	r.DB.Create(&serviceSpec)

	//r.ServiceSpecCreated(&serviceSpec)

	return &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}, nil
}

func (r *Resolver) UpdateServiceSpec(args *struct{ ServiceSpec *ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := ServiceSpec{}

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

	return &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}, nil
}

func (r *Resolver) DeleteServiceSpec(args *struct{ ServiceSpec *ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := ServiceSpec{}
	if r.DB.Where("id=?", args.ServiceSpec.ID).Find(&serviceSpec).RecordNotFound() {
		return nil, fmt.Errorf("ServiceSpec not found with given argument id")
	} else {
		services := []Service{}
		r.DB.Where("service_spec_id = ?", serviceSpec.Model.ID).Find(&services)
		if len(services) == 0 {
			r.DB.Delete(&serviceSpec)

			//r.ServiceSpecDeleted(&serviceSpec)

			return &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}, nil
		} else {
			return nil, fmt.Errorf("Delete all project-services using this service spec first.")
		}
	}
}

func (r *Resolver) CreateEnvironment(ctx context.Context, args *struct{ Environment *EnvironmentInput }) (*EnvironmentResolver, error) {
	var existingEnv Environment
	if r.DB.Where("key = ?", args.Environment.Key).Find(&existingEnv).RecordNotFound() {
		env := Environment{
			Name:      args.Environment.Name,
			Key:       args.Environment.Key,
			IsDefault: args.Environment.IsDefault,
			Color:     args.Environment.Color,
		}

		r.DB.Create(&env)

		//r.EnvironmentCreated(&env)

		return &EnvironmentResolver{DB: r.DB, Environment: env}, nil
	} else {
		return nil, fmt.Errorf("CreateEnvironment: name already exists")
	}
}

func (r *Resolver) UpdateEnvironment(ctx context.Context, args *struct{ Environment *EnvironmentInput }) (*EnvironmentResolver, error) {
	var existingEnv Environment
	if r.DB.Where("id = ?", args.Environment.ID).Find(&existingEnv).RecordNotFound() {
		return nil, fmt.Errorf("UpdateEnv: couldn't find environment: %s", *args.Environment.ID)
	} else {
		existingEnv.Name = args.Environment.Name
		existingEnv.Color = args.Environment.Color

		// Check if this is the only default env.
		if existingEnv.IsDefault {
			var defaultEnvs []Environment
			r.DB.Where("is_default = ?", true).Find(&defaultEnvs)
			// Update IsDefault as long as the current is false or
			// if there are more than 1 default env
			if len(defaultEnvs) > 1 {
				existingEnv.IsDefault = args.Environment.IsDefault
			}
		} else {
			// If IsDefault is false, then no harm in updating
			existingEnv.IsDefault = args.Environment.IsDefault
		}

		r.DB.Save(&existingEnv)

		return &EnvironmentResolver{DB: r.DB, Environment: existingEnv}, nil
	}
}

func (r *Resolver) DeleteEnvironment(ctx context.Context, args *struct{ Environment *EnvironmentInput }) (*EnvironmentResolver, error) {
	var existingEnv Environment
	if r.DB.Where("id = ?", args.Environment.ID).Find(&existingEnv).RecordNotFound() {
		return nil, fmt.Errorf("DeleteEnv: couldn't find environment: %s", *args.Environment.ID)
	} else {
		// if this is the only default env, do not delete
		if existingEnv.IsDefault {
			var defaultEnvs []Environment
			r.DB.Where("is_default = ?", true).Find(&defaultEnvs)
			if len(defaultEnvs) == 1 {
				return nil, fmt.Errorf("Cannot delete since this is the only default env. Must be one at all times")
			}
		}

		// Only delete env. if no child services exist, else return err
		childServices := []Service{}
		r.DB.Where("environment_id = ?", args.Environment.ID).Find(&childServices)
		if len(childServices) == 0 {
			existingEnv.Name = args.Environment.Name
			secrets := []Secret{}

			r.DB.Delete(&existingEnv)
			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Find(&secrets)
			for _, secret := range secrets {
				r.DB.Delete(&secret)
				r.DB.Where("secret_id = ?", secret.Model.ID).Delete(SecretValue{})
			}

			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Delete(Release{})
			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Delete(ReleaseExtension{})
			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Delete(ProjectExtension{})
			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Delete(ProjectSettings{})
			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Delete(Extension{})

			//r.EnvironmentDeleted(&existingEnv)

			return &EnvironmentResolver{DB: r.DB, Environment: existingEnv}, nil
		} else {
			return nil, fmt.Errorf("Delete all project-services in environment before deleting environment.")
		}
	}
}

func (r *Resolver) CreateSecret(ctx context.Context, args *struct{ Secret *SecretInput }) (*SecretResolver, error) {

	projectID := uuid.UUID{}
	var environmentID uuid.UUID
	var secretScope SecretScope

	if args.Secret.ProjectID != nil {
		// Check if project can create secret
		if r.DB.Where("environment_id = ? and project_id = ?", args.Secret.EnvironmentID, args.Secret.ProjectID).Find(&ProjectEnvironment{}).RecordNotFound() {
			return nil, errors.New("Project not allowed to create secret in given environment")
		}

		projectID = uuid.FromStringOrNil(*args.Secret.ProjectID)
	}

	secretScope = GetSecretScope(args.Secret.Scope)
	if secretScope == SecretScope("unknown") {
		return nil, fmt.Errorf("Invalid env var scope.")
	}

	environmentID, err := uuid.FromString(args.Secret.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse environmentID. Invalid format.")
	}

	userIDString, err := CheckAuth(ctx, []string{})
	if err != nil {
		return &SecretResolver{}, err
	}

	userID, err := uuid.FromString(userIDString)
	if err != nil {
		return &SecretResolver{}, err
	}

	var existingEnvVar Secret

	if r.DB.Where("key = ? and project_id = ? and deleted_at is null and environment_id = ?", args.Secret.Key, projectID, environmentID).Find(&existingEnvVar).RecordNotFound() {
		secret := Secret{
			Key:           args.Secret.Key,
			ProjectID:     projectID,
			Type:          plugins.GetType(args.Secret.Type),
			Scope:         secretScope,
			EnvironmentID: environmentID,
			IsSecret:      args.Secret.IsSecret,
		}
		r.DB.Create(&secret)

		secretValue := SecretValue{
			SecretID: secret.Model.ID,
			Value:    args.Secret.Value,
			UserID:   userID,
		}
		r.DB.Create(&secretValue)

		//r.SecretCreated(&secret)

		return &SecretResolver{DB: r.DB, Secret: secret, SecretValue: secretValue}, nil
	} else {
		return nil, fmt.Errorf("CreateSecret: key already exists")
	}

}

func (r *Resolver) UpdateSecret(ctx context.Context, args *struct{ Secret *SecretInput }) (*SecretResolver, error) {
	var secret Secret

	userIDString, err := CheckAuth(ctx, []string{})
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
		secretValue := SecretValue{
			SecretID: secret.Model.ID,
			Value:    args.Secret.Value,
			UserID:   userID,
		}
		r.DB.Create(&secretValue)

		//r.SecretUpdated(&secret)

		return &SecretResolver{DB: r.DB, Secret: secret, SecretValue: secretValue}, nil
	}
}

func (r *Resolver) DeleteSecret(ctx context.Context, args *struct{ Secret *SecretInput }) (*SecretResolver, error) {
	var secret Secret

	if r.DB.Where("id = ?", args.Secret.ID).Find(&secret).RecordNotFound() {
		return nil, fmt.Errorf("DeleteSecret: key doesn't exist.")
	} else {
		// check if any configs are using the secret
		extensions := []Extension{}
		r.DB.Where(`config @> '{"config": [{"value": "?"}]}'"`, secret.Model.ID.String()).Find(&extensions)
		if len(extensions) == 0 {
			versions := []SecretValue{}

			r.DB.Delete(&secret)
			r.DB.Where("secret_id = ?", secret.Model.ID).Delete(&versions)

			//r.SecretDeleted(&secret)

			return &SecretResolver{DB: r.DB, Secret: secret}, nil
		} else {
			return nil, fmt.Errorf("Remove Config values from Extensions where Secret is used before deleting.")
		}
	}
}

func (r *Resolver) CreateExtension(args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	environmentID, err := uuid.FromString(args.Extension.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Missing argument EnvironmentID")
	}

	ext := Extension{
		Name:          args.Extension.Name,
		Component:     args.Extension.Component,
		Type:          plugins.Type(args.Extension.Type),
		Key:           args.Extension.Key,
		EnvironmentID: environmentID,
		Config:        postgres.Jsonb{[]byte(args.Extension.Config.RawMessage)},
	}

	r.DB.Create(&ext)
	//r.ExtensionCreated(&ext)

	return &ExtensionResolver{DB: r.DB, Extension: ext}, nil
}

func (r *Resolver) UpdateExtension(args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	ext := Extension{}
	if r.DB.Where("id = ?", args.Extension.ID).Find(&ext).RecordNotFound() {
		log.InfoWithFields("could not find extensionspec with id", log.Fields{
			"id": args.Extension.ID,
		})
		return &ExtensionResolver{DB: r.DB, Extension: Extension{}}, fmt.Errorf("could not find extensionspec with id")
	}

	environmentID, err := uuid.FromString(args.Extension.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Missing argument EnvironmentID")
	}

	// update extensionspec properties
	ext.Name = args.Extension.Name
	ext.Key = args.Extension.Key
	ext.Type = plugins.Type(args.Extension.Type)
	ext.Component = args.Extension.Component
	ext.EnvironmentID = environmentID
	ext.Config = postgres.Jsonb{args.Extension.Config.RawMessage}

	r.DB.Save(&ext)

	//r.ExtensionUpdated(&ext)

	return &ExtensionResolver{DB: r.DB, Extension: ext}, nil
}

func (r *Resolver) DeleteExtension(args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	ext := Extension{}
	extensions := []ProjectExtension{}
	extID, err := uuid.FromString(*args.Extension.ID)
	if err != nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	if r.DB.Where("id=?", extID).Find(&ext).RecordNotFound() {
		return nil, fmt.Errorf("Extension not found with given argument id")
	}

	// delete all extensions using extension spec
	if r.DB.Where("extension_id = ?", extID).Find(&extensions).RecordNotFound() {
		log.InfoWithFields("no extensions using this extension spec", log.Fields{
			"extension spec": ext,
		})
	}

	if len(extensions) > 0 {
		return nil, fmt.Errorf("You must delete all extensions using this extension spec in order to delete this extension spec.")
	} else {
		r.DB.Delete(&ext)

		//r.ExtensionDeleted(&ext)

		return &ExtensionResolver{DB: r.DB, Extension: ext}, nil
	}
}

func (r *Resolver) CreateProjectExtension(ctx context.Context, args *struct{ ProjectExtension *ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var projectExtension ProjectExtension

	// Check if project can create project extension in environment
	if r.DB.Where("environment_id = ? and project_id = ?", args.ProjectExtension.EnvironmentID, args.ProjectExtension.ProjectID).Find(&ProjectEnvironment{}).RecordNotFound() {
		return nil, errors.New("Project not allowed to install extensions in given environment")
	}

	extension := Extension{}
	if r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return nil, errors.New("No extension found.")
	}

	project := Project{}
	if r.DB.Where("id = ?", args.ProjectExtension.ProjectID).Find(&project).RecordNotFound() {
		log.InfoWithFields("no project found", log.Fields{
			"id": args.ProjectExtension.ProjectID,
		})
		return nil, errors.New("No project found.")
	}

	env := Environment{}
	if r.DB.Where("id = ?", args.ProjectExtension.EnvironmentID).Find(&env).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.ProjectExtension.EnvironmentID,
		})
		return nil, errors.New("No environment found.")
	}

	// check if extension already exists with project
	// ignore if the extension type is 'once' (installable many times)
	if extension.Type == plugins.GetType("once") || r.DB.Where("project_id = ? and extension_id = ? and environment_id = ?", args.ProjectExtension.ProjectID, args.ProjectExtension.ExtensionID, args.ProjectExtension.EnvironmentID).Find(&projectExtension).RecordNotFound() {
		if extension.Key == "route53" {
			// HOTFIX: check for existing subdomains for route53
			unmarshaledCustomConfig := make(map[string]interface{})
			err := json.Unmarshal(args.ProjectExtension.CustomConfig.RawMessage, &unmarshaledCustomConfig)
			if err != nil {
				return &ProjectExtensionResolver{}, errors.New("Could not unmarshal custom config")
			}

			artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
			if err != nil {
				return &ProjectExtensionResolver{}, err
			}

			hostedZoneId := ""
			for _, artifact := range artifacts {
				if artifact.Key == "HOSTED_ZONE_ID" {
					hostedZoneId = strings.ToUpper(artifact.Value.(string))
					break
				}
			}

			existingProjectExtensions := GetProjectExtensionsWithRoute53Subdomain(strings.ToUpper(unmarshaledCustomConfig["subdomain"].(string)), r.DB)
			for _, existingProjectExtension := range existingProjectExtensions {
				if existingProjectExtension.Model.ID.String() != "" {
					// check if HOSTED_ZONE_ID is the same
					var tmpExtension Extension

					r.DB.Where("id = ?", existingProjectExtension.ExtensionID).First(&tmpExtension)

					tmpExtensionArtifacts, err := ExtractArtifacts(existingProjectExtension, tmpExtension, r.DB)
					if err != nil {
						return &ProjectExtensionResolver{}, err
					}

					for _, artifact := range tmpExtensionArtifacts {
						if artifact.Key == "HOSTED_ZONE_ID" &&
							strings.ToUpper(artifact.Value.(string)) == hostedZoneId {
							errMsg := "There is a route53 project extension with inputted subdomain already."
							log.InfoWithFields(errMsg, log.Fields{
								"project_extension_id":          projectExtension.Model.ID.String(),
								"existing_project_extension_id": existingProjectExtension.Model.ID.String(),
								"environment_id":                projectExtension.EnvironmentID.String(),
								"hosted_zone_id":                hostedZoneId,
							})
							return &ProjectExtensionResolver{}, errors.New(errMsg)
						}
					}
				}
			}
		}

		projectExtension = ProjectExtension{
			ExtensionID:   extension.Model.ID,
			ProjectID:     project.Model.ID,
			EnvironmentID: env.Model.ID,
			Config:        postgres.Jsonb{[]byte(args.ProjectExtension.Config.RawMessage)},
			CustomConfig:  postgres.Jsonb{[]byte(args.ProjectExtension.CustomConfig.RawMessage)},
		}

		r.DB.Save(&projectExtension)

		artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
		if err != nil {
			log.Info(err.Error())
		}

		projectExtensionEvent := plugins.ProjectExtension{
			ID:   projectExtension.Model.ID.String(),
			Slug: extension.Key,
			Project: plugins.Project{
				ID:         project.Model.ID.String(),
				Slug:       project.Slug,
				Repository: project.Repository,
			},
			Environment: env.Key,
		}
		ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("project:%s", extension.Key)), plugins.GetAction("create"), projectExtensionEvent)
		ev.Artifacts = artifacts
		r.Events <- ev

		return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: projectExtension}, nil
	}

	return nil, errors.New("This extension is already installed in this project.")
}

func (r *Resolver) UpdateProjectExtension(args *struct{ ProjectExtension *ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var projectExtension ProjectExtension

	if r.DB.Where("id = ?", args.ProjectExtension.ID).First(&projectExtension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.ProjectExtension,
		})
		return &ProjectExtensionResolver{}, nil
	}

	extension := Extension{}
	if r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return nil, errors.New("No extension found.")
	}

	project := Project{}
	if r.DB.Where("id = ?", args.ProjectExtension.ProjectID).Find(&project).RecordNotFound() {
		log.InfoWithFields("no project found", log.Fields{
			"id": args.ProjectExtension.ProjectID,
		})
		return nil, errors.New("No project found.")
	}

	env := Environment{}
	if r.DB.Where("id = ?", args.ProjectExtension.EnvironmentID).Find(&env).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.ProjectExtension.EnvironmentID,
		})
		return nil, errors.New("No environment found.")
	}

	if extension.Key == "route53" {
		// HOTFIX: check for existing subdomains for route53
		unmarshaledCustomConfig := make(map[string]interface{})
		err := json.Unmarshal(args.ProjectExtension.CustomConfig.RawMessage, &unmarshaledCustomConfig)
		if err != nil {
			return &ProjectExtensionResolver{}, errors.New("Could not unmarshal custom config")
		}

		artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
		if err != nil {
			return &ProjectExtensionResolver{}, err
		}

		hostedZoneId := ""
		for _, artifact := range artifacts {
			if artifact.Key == "HOSTED_ZONE_ID" {
				hostedZoneId = strings.ToUpper(artifact.Value.(string))
				break
			}
		}

		existingProjectExtensions := GetProjectExtensionsWithRoute53Subdomain(strings.ToUpper(unmarshaledCustomConfig["subdomain"].(string)), r.DB)
		for _, existingProjectExtension := range existingProjectExtensions {
			if existingProjectExtension.Model.ID.String() != "" {
				// check if HOSTED_ZONE_ID is the same
				var tmpExtension Extension

				r.DB.Where("id = ?", existingProjectExtension.ExtensionID).First(&tmpExtension)

				tmpExtensionArtifacts, err := ExtractArtifacts(existingProjectExtension, tmpExtension, r.DB)
				if err != nil {
					return &ProjectExtensionResolver{}, err
				}

				for _, artifact := range tmpExtensionArtifacts {
					if artifact.Key == "HOSTED_ZONE_ID" &&
						strings.ToUpper(artifact.Value.(string)) == hostedZoneId {
						errMsg := "There is a route53 project extension with inputted subdomain already."
						log.InfoWithFields(errMsg, log.Fields{
							"project_extension_id":          projectExtension.Model.ID.String(),
							"existing_project_extension_id": existingProjectExtension.Model.ID.String(),
							"environment_id":                projectExtension.EnvironmentID.String(),
							"hosted_zone_id":                hostedZoneId,
						})
						return &ProjectExtensionResolver{}, errors.New(errMsg)
					}
				}
			}
		}
	}

	projectExtension.Config = postgres.Jsonb{args.ProjectExtension.Config.RawMessage}
	projectExtension.CustomConfig = postgres.Jsonb{args.ProjectExtension.CustomConfig.RawMessage}

	r.DB.Save(&projectExtension)

	artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
	if err != nil {
		log.Info(err.Error())
	}

	projectExtensionEvent := plugins.ProjectExtension{
		ID:   projectExtension.Model.ID.String(),
		Slug: extension.Key,
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Environment: env.Key,
	}

	ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("project:%s", extension.Key)), plugins.GetAction("update"), projectExtensionEvent)
	ev.Artifacts = artifacts

	r.Events <- ev

	return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: projectExtension}, nil
}

func (r *Resolver) DeleteProjectExtension(args *struct{ ProjectExtension *ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var projectExtension ProjectExtension
	var res []ReleaseExtension

	if r.DB.Where("id = ?", args.ProjectExtension.ID).First(&projectExtension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.ProjectExtension,
		})
		return &ProjectExtensionResolver{}, nil
	}

	extension := Extension{}
	if r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return nil, errors.New("No extension found.")
	}

	project := Project{}
	if r.DB.Where("id = ?", args.ProjectExtension.ProjectID).Find(&project).RecordNotFound() {
		log.InfoWithFields("no project found", log.Fields{
			"id": args.ProjectExtension.ProjectID,
		})
		return nil, errors.New("No project found.")
	}

	env := Environment{}
	if r.DB.Where("id = ?", args.ProjectExtension.EnvironmentID).Find(&env).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.ProjectExtension.EnvironmentID,
		})
		return nil, errors.New("No environment found.")
	}

	// delete all release extension objects with extension id
	if r.DB.Where("extension_id = ?", args.ProjectExtension.ID).Find(&res).RecordNotFound() {
		log.InfoWithFields("no release extensions found", log.Fields{
			"extension": extension,
		})
		return &ProjectExtensionResolver{}, nil
	}

	for _, re := range res {
		r.DB.Delete(&re)
	}

	r.DB.Delete(&projectExtension)

	artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
	if err != nil {
		log.Info(err.Error())
	}

	projectExtensionEvent := plugins.ProjectExtension{
		ID:   projectExtension.Model.ID.String(),
		Slug: extension.Key,
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Environment: env.Key,
	}
	ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("project:%s", extension.Key)), plugins.GetAction("destroy"), projectExtensionEvent)
	ev.Artifacts = artifacts
	r.Events <- ev

	return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: projectExtension}, nil
}

// UpdateUserPermissions
func (r *Resolver) UpdateUserPermissions(ctx context.Context, args *struct{ UserPermissions *UserPermissionsInput }) ([]string, error) {
	var err error
	var results []string

	if r.DB.Where("id = ?", args.UserPermissions.UserID).Find(&User{}).RecordNotFound() {
		return nil, errors.New("User not found")
	}

	for _, permission := range args.UserPermissions.Permissions {
		if _, err = CheckAuth(ctx, []string{permission.Value}); err != nil {
			return nil, err
		}
	}

	for _, permission := range args.UserPermissions.Permissions {
		if permission.Grant == true {
			userPermission := UserPermission{
				UserID: uuid.FromStringOrNil(args.UserPermissions.UserID),
				Value:  permission.Value,
			}
			r.DB.Where(userPermission).FirstOrCreate(&userPermission)
			results = append(results, permission.Value)
		} else {
			r.DB.Where("user_id = ? AND value = ?", args.UserPermissions.UserID, permission.Value).Delete(&UserPermission{})
		}
	}

	return results, nil
}

// UpdateProjectEnvironments
func (r *Resolver) UpdateProjectEnvironments(ctx context.Context, args *struct{ ProjectEnvironments *ProjectEnvironmentsInput }) ([]*EnvironmentResolver, error) {
	var results []*EnvironmentResolver

	project := Project{}
	if r.DB.Where("id = ?", args.ProjectEnvironments.ProjectID).Find(&project).RecordNotFound() {
		return nil, errors.New("No project found with inputted projectID")
	}

	for _, permission := range args.ProjectEnvironments.Permissions {
		// Check if environment object exists
		environment := Environment{}
		if r.DB.Where("id = ?", permission.EnvironmentID).Find(&environment).RecordNotFound() {
			return nil, errors.New(fmt.Sprintf("No environment found for environmentID %s", permission.EnvironmentID))
		}

		if permission.Grant {
			// Grant permission by adding ProjectEnvironment row
			projectEnvironment := ProjectEnvironment{
				EnvironmentID: environment.Model.ID,
				ProjectID:     project.Model.ID,
			}
			r.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID, project.Model.ID).FirstOrCreate(&projectEnvironment)
			results = append(results, &EnvironmentResolver{DB: r.DB, Environment: environment})
		} else {
			r.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID, project.Model.ID).Delete(&ProjectEnvironment{})
		}
	}

	return results, nil
}

func (r *Resolver) BookmarkProject(ctx context.Context, args *struct{ ID graphql.ID }) (bool, error) {
	var projectBookmark ProjectBookmark

	_userID, err := CheckAuth(ctx, []string{})
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
		projectBookmark = ProjectBookmark{
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

func GetProjectExtensionsWithRoute53Subdomain(subdomain string, db *gorm.DB) []ProjectExtension {
	var existingProjectExtensions []ProjectExtension

	if db.Where("custom_config ->> 'subdomain' ilike ?", subdomain).Find(&existingProjectExtensions).RecordNotFound() {
		return []ProjectExtension{}
	}

	return existingProjectExtensions
}

/* fills in Config by querying config ids and getting the actual value */
func ExtractArtifacts(projectExtension ProjectExtension, extension Extension, db *gorm.DB) ([]transistor.Artifact, error) {
	var artifacts []transistor.Artifact
	var err error

	type ExtConfig struct {
		Key           string `json:"key"`
		Value         string `json:"value"`
		Secret        bool   `json:"secret"`
		AllowOverride bool   `json:"allowOverride"`
	}

	extensionConfig := []ExtConfig{}
	err = json.Unmarshal(extension.Config.RawMessage, &extensionConfig)
	if err != nil {
		log.Info(err.Error())
	}

	projectConfig := []ExtConfig{}
	err = json.Unmarshal(projectExtension.Config.RawMessage, &projectConfig)
	if err != nil {
		log.Info(err.Error())
	}

	existingArtifacts := []transistor.Artifact{}
	err = json.Unmarshal(projectExtension.Artifacts.RawMessage, &existingArtifacts)
	if err != nil {
		log.Info(err.Error())
	}

	for i, ec := range extensionConfig {
		for _, pc := range projectConfig {
			if ec.AllowOverride && ec.Key == pc.Key && pc.Value != "" {
				extensionConfig[i].Value = pc.Value
			}
		}
	}

	for _, ec := range extensionConfig {
		var artifact transistor.Artifact
		// check if val is UUID. If so, query in environment variables for id
		secretID := uuid.FromStringOrNil(ec.Value)
		if secretID != uuid.Nil {
			secret := SecretValue{}
			if db.Where("secret_id = ?", secretID).Order("created_at desc").First(&secret).RecordNotFound() {
				log.InfoWithFields("secret not found", log.Fields{
					"secret_id": secretID,
				})
			}
			artifact.Key = ec.Key
			artifact.Value = secret.Value
		} else {
			artifact.Key = ec.Key
			artifact.Value = ec.Value
		}
		artifacts = append(artifacts, artifact)
	}

	for _, ea := range existingArtifacts {
		artifacts = append(artifacts, ea)
	}

	projectCustomConfig := make(map[string]interface{})
	err = json.Unmarshal(projectExtension.CustomConfig.RawMessage, &projectCustomConfig)
	if err != nil {
		log.Info(err.Error())
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
