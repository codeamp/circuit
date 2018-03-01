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

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/extemporalgenome/slug"
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
	if r.DB.Unscoped().Where("repository = ?", repository).First(&existingProject).RecordNotFound() {
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
	}

	if userId, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	} else {
		// Create user permission for project
		userPermission := UserPermission{
			UserID: uuid.FromStringOrNil(userId),
			Value:  fmt.Sprintf("user/%s", project.Repository),
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
		var projectSettings ProjectSettings
		if r.DB.Where("environment_id = ? and project_id = ?", args.Project.EnvironmentID, args.Project.ID).First(&projectSettings).RecordNotFound() {
			log.InfoWithFields("Project settings not found", log.Fields{})
		} else {
			projectSettings.GitBranch = *args.Project.GitBranch
			r.DB.Save(&projectSettings)
		}
	}

	r.DB.Save(&project)

	return &ProjectResolver{DB: r.DB, Project: project}, nil
}

// CreateRelease
func (r *Resolver) CreateRelease(ctx context.Context, args *struct{ Release *ReleaseInput }) (*ReleaseResolver, error) {
	var project Project
	var secrets []Secret
	var services []Service
	var extensions []ProjectExtension
	var secretsJsonb postgres.Jsonb
	var servicesJsonb postgres.Jsonb
	var extensionsJsonb postgres.Jsonb

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

		if r.DB.Where("project_id = ? and environment_id = ?", args.Release.ProjectID, args.Release.EnvironmentID).Find(&services).RecordNotFound() {
			log.InfoWithFields("no services found", log.Fields{
				"project_id": args.Release.ProjectID,
			})
			return &ReleaseResolver{}, fmt.Errorf("no services found")
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

		if r.DB.Where("project_id = ? AND environment_id = ? AND state = ?", args.Release.ProjectID, args.Release.EnvironmentID, plugins.GetState("complete")).Find(&extensions).RecordNotFound() {
			log.InfoWithFields("project has no extensions", log.Fields{
				"project_id":     args.Release.ProjectID,
				"environment_id": args.Release.EnvironmentID,
			})
		}

		extensionsMarshaled, err := json.Marshal(extensions)
		if err != nil {
			return &ReleaseResolver{}, err
		}

		extensionsJsonb = postgres.Jsonb{extensionsMarshaled}
	} else {
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
		extensionsJsonb = release.ProjectExtensions
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
		State:             plugins.GetState("waiting"),
		StateMessage:      "Release created",
		Secrets:           secretsJsonb,
		Services:          servicesJsonb,
		ProjectExtensions: extensionsJsonb,
		Artifacts:         postgres.Jsonb{[]byte("{}")},
	}

	r.DB.Create(&release)

	if r.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": release.ProjectID,
		})
		return &ReleaseResolver{}, errors.New("Project not found")
	}

	// get all branches relevant for the projec
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

		pluginServices = append(pluginServices, plugins.Service{
			ID:        service.Model.ID.String(),
			Action:    plugins.GetAction("create"),
			State:     plugins.GetState("waiting"),
			Name:      service.Name,
			Command:   service.Command,
			Listeners: []plugins.Listener{},
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

	releaseEvent := plugins.Release{
		ID:          release.Model.ID.String(),
		Action:      plugins.GetAction("create"),
		State:       plugins.GetState("waiting"),
		Environment: environment.Name,
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

	r.Events <- transistor.NewEvent(releaseEvent, nil)

	// Create/Emit Release ProjectExtensions
	for _, extension := range extensions {
		ext := Extension{}
		if r.DB.Where("id= ?", extension.ExtensionID).Find(&ext).RecordNotFound() {
			log.ErrorWithFields("extension spec not found", log.Fields{
				"id": extension.ExtensionID,
			})
			return &ReleaseResolver{}, errors.New("extension spec not found")
		}

		if plugins.Type(ext.Type) == plugins.GetType("workflow") {
			var headFeature Feature
			if r.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
				log.ErrorWithFields("head feature not found", log.Fields{
					"id": release.HeadFeatureID,
				})
				return &ReleaseResolver{}, errors.New("head feature not found")
			}

			secretsSha1 := sha1.New()
			secretsSha1.Write(secretsJsonb.RawMessage)
			secretsSig := secretsSha1.Sum(nil)

			servicesSha1 := sha1.New()
			servicesSha1.Write(servicesJsonb.RawMessage)
			servicesSig := servicesSha1.Sum(nil)

			// create ReleaseExtension
			releaseExtension := ReleaseExtension{
				ReleaseID:          release.Model.ID,
				FeatureHash:        headFeature.Hash,
				ServicesSignature:  fmt.Sprintf("%x", servicesSig),
				SecretsSignature:   fmt.Sprintf("%x", secretsSig),
				ProjectExtensionID: extension.Model.ID,
				State:              plugins.GetState("waiting"),
				Type:               plugins.GetType("workflow"),
			}

			r.DB.Create(&releaseExtension)

			r.Events <- transistor.NewEvent(plugins.ReleaseExtension{
				ID:        releaseExtension.Model.ID.String(),
				Action:    plugins.GetAction("create"),
				Slug:      ext.Key,
				State:     releaseExtension.State,
				Release:   releaseEvent,
				Secrets:   map[string]string{},
				Config:    map[string]interface{}{},
				Artifacts: map[string]interface{}{},
			}, nil)
		}
	}

	return &ReleaseResolver{DB: r.DB, Release: Release{}}, nil
}

// CreateService Create service
func (r *Resolver) CreateService(args *struct{ Service *ServiceInput }) (*ServiceResolver, error) {
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

	serviceSpecID, err := uuid.FromString(*args.ServiceSpec.ID)
	if err != nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	if r.DB.Where("id=?", serviceSpecID).Find(&serviceSpec).RecordNotFound() {
		return nil, fmt.Errorf("ServiceSpec not found with given argument id")
	}

	r.DB.Delete(serviceSpec)

	//r.ServiceSpecDeleted(&serviceSpec)

	return &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}, nil
}

func (r *Resolver) CreateEnvironment(ctx context.Context, args *struct{ Environment *EnvironmentInput }) (*EnvironmentResolver, error) {

	var existingEnv Environment
	if r.DB.Where("name = ?", args.Environment.Name).Find(&existingEnv).RecordNotFound() {
		env := Environment{
			Name:  args.Environment.Name,
			Color: args.Environment.Color,
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

		r.DB.Save(&existingEnv)

		//r.EnvironmentUpdated(&existingEnv)

		return &EnvironmentResolver{DB: r.DB, Environment: existingEnv}, nil
	}
}

func (r *Resolver) DeleteEnvironment(ctx context.Context, args *struct{ Environment *EnvironmentInput }) (*EnvironmentResolver, error) {
	var existingEnv Environment
	if r.DB.Where("id = ?", args.Environment.ID).Find(&existingEnv).RecordNotFound() {
		return nil, fmt.Errorf("DeleteEnv: couldn't find environment: %s", *args.Environment.ID)
	} else {
		existingEnv.Name = args.Environment.Name
		r.DB.Delete(&existingEnv)

		//r.EnvironmentDeleted(&existingEnv)

		return &EnvironmentResolver{DB: r.DB, Environment: existingEnv}, nil
	}
}

func (r *Resolver) CreateSecret(ctx context.Context, args *struct{ Secret *SecretInput }) (*SecretResolver, error) {

	projectID := uuid.UUID{}
	var environmentID uuid.UUID
	var secretScope SecretScope

	if args.Secret.ProjectID != nil {
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
		var rows []Secret

		r.DB.Where("project_id = ? and key = ? and environment_id = ?", secret.ProjectID, secret.Key, secret.EnvironmentID).Find(&rows)
		for _, ev := range rows {
			r.DB.Unscoped().Delete(&ev)
		}

		//r.SecretDeleted(&secret)

		return &SecretResolver{DB: r.DB, Secret: secret}, nil
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
	if r.DB.Where("extension_spec_id = ?", extID).Find(&extensions).RecordNotFound() {
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
	var extension ProjectExtension

	extID, err := uuid.FromString(args.ProjectExtension.ExtensionID)
	if err != nil {
		log.InfoWithFields("couldn't parse ExtensionID", log.Fields{
			"extension": args.ProjectExtension,
		})
		return nil, errors.New("Could not parse ExtensionID. Invalid Format.")
	}

	projectID, err := uuid.FromString(args.ProjectExtension.ProjectID)
	if err != nil {
		log.InfoWithFields("couldn't parse ProjectID", log.Fields{
			"extension": args.ProjectExtension,
		})
		return nil, errors.New("Could not parse ProjectID. Invalid format.")
	}

	environmentID, err := uuid.FromString(args.ProjectExtension.EnvironmentID)
	if err != nil {
		log.InfoWithFields("couldn't parse EnvironmentID", log.Fields{
			"extension": args.ProjectExtension,
		})
		return nil, errors.New("Could not parse EnvironmentID. Invalid format.")
	}

	// get extensionspec
	var ext Extension
	if r.DB.Where("id = ?", extID).Find(&ext).RecordNotFound() {
		log.InfoWithFields("Could not find an extension spec while trying to CreateProjectExtension", log.Fields{
			"id": extID,
		})
	}

	// check if extension already exists with project
	if ext.Type == plugins.GetType("once") || r.DB.Where("project_id = ? and extension_spec_id = ? and environment_id = ?", projectID, extID, environmentID).Find(&extension).RecordNotFound() {
		extension = ProjectExtension{
			ExtensionID:   extID,
			ProjectID:     projectID,
			EnvironmentID: environmentID,
			Config:        postgres.Jsonb{[]byte(args.ProjectExtension.Config.RawMessage)},
			State:         plugins.GetState("waiting"),
			Artifacts:     postgres.Jsonb{},
		}
		r.DB.Save(&extension)

		//r.ProjectExtensionCreated(&extension)

		return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension}, nil
	}

	return nil, errors.New("This extension is already installed in this project.")
}

func (r *Resolver) UpdateProjectExtension(args *struct{ ProjectExtension *ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var extension ProjectExtension

	if r.DB.Where("id = ?", args.ProjectExtension.ID).First(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.ProjectExtension,
		})
		return &ProjectExtensionResolver{}, nil
	}
	extension.Config = postgres.Jsonb{args.ProjectExtension.Config.RawMessage}
	extension.State = plugins.GetState("waiting")

	r.DB.Save(&extension)

	//r.ProjectExtensionUpdated(&extension)

	return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension}, nil
}

func (r *Resolver) DeleteProjectExtension(args *struct{ ProjectExtension *ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var extension ProjectExtension
	var res []ReleaseExtension

	if r.DB.Where("id = ?", args.ProjectExtension.ID).First(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.ProjectExtension,
		})
		return &ProjectExtensionResolver{}, nil
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

	r.DB.Delete(&extension)

	//r.ProjectExtensionDeleted(&extension)

	return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension}, nil
}

// UpdateUserPermissions
func (r *Resolver) UpdateUserPermissions(ctx context.Context, args *struct{ UserPermissionsInput *UserPermissionsInput }) ([]string, error) {
	var err error
	var results []string

	if r.DB.Where("id = ?", args.UserPermissionsInput.UserID).Find(User{}).RecordNotFound() {
		return nil, errors.New("User not found")
	}

	for _, permission := range args.UserPermissionsInput.Permissions {
		if _, err = CheckAuth(ctx, []string{permission.Value}); err != nil {
			return nil, err
		}
	}

	for _, permission := range args.UserPermissionsInput.Permissions {
		if permission.Grant == true {
			userPermission := UserPermission{
				UserID: uuid.FromStringOrNil(args.UserPermissionsInput.UserID),
				Value:  permission.Value,
			}
			r.DB.Where(userPermission).FirstOrCreate(&userPermission)
			results = append(results, permission.Value)
		} else {
			r.DB.Where("user_id = ? AND value = ?", args.UserPermissionsInput.UserID, permission.Value).Delete(UserPermission{})
		}
	}

	return results, nil
}
