package codeamp_resolvers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/extemporalgenome/slug"
	"github.com/jinzhu/gorm/dialects/postgres"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
)

// CreateProject Create project
func (r *Resolver) CreateProject(args *struct {
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
			EnvironmentId: env.Model.ID,
			ProjectId:     project.Model.ID,
			GitBranch:     "master",
		})
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

	projectId, err := uuid.FromString(*args.Project.ID)
	if err != nil {
		log.InfoWithFields("Could not convert argument id", log.Fields{
			"id":  args.Project.ID,
			"err": err,
		})
		return nil, fmt.Errorf("Invalid argument id")
	}

	if r.DB.Where("id = ?", args.Project.ID).First(&project).RecordNotFound() {
		log.InfoWithFields("Project not found", log.Fields{
			"id": args.Project.ID,
		})
		return nil, fmt.Errorf("Project not found.")
	}

	protocol := "HTTPS"
	switch args.Project.GitProtocol {
	case "private", "PRIVATE", "ssh", "SSH":
		protocol = "SSH"
	case "public", "PUBLIC", "https", "HTTPS":
		protocol = "HTTPS"
	}

	res := plugins.GetRegexParams("(?P<host>(git@|https?:\\/\\/)([\\w\\.@]+)(\\/|:))(?P<owner>[\\w,\\-,\\_]+)\\/(?P<repo>[\\w,\\-,\\_]+)(.git){0,1}((\\/){0,1})", args.Project.GitUrl)
	repository := fmt.Sprintf("%s/%s", res["owner"], res["repo"])

	project.GitUrl = args.Project.GitUrl

	// Check if project already exists with same name
	if r.DB.Unscoped().Where("id != ? and repository = ?", projectId, repository).First(&Project{}).RecordNotFound() == false {
		return nil, fmt.Errorf("Project with repository name already exists.")
	}

	project.GitUrl = args.Project.GitUrl
	project.GitProtocol = protocol
	project.Repository = repository
	project.Name = repository
	project.Slug = slug.Slug(repository)
	r.DB.Save(&project)

	return &ProjectResolver{DB: r.DB, Project: project}, nil
}

// CreateRelease
func (r *Resolver) CreateRelease(ctx context.Context, args *struct{ Release *ReleaseInput }) (*ReleaseResolver, error) {
	var tailFeatureId uuid.UUID
	var currentRelease Release
	var releaseFromId Release

	if args.Release.ID != nil {
		if r.DB.Where("id = ?", *args.Release.ID).Find(&releaseFromId).RecordNotFound() {
			return &ReleaseResolver{}, errors.New(fmt.Sprintf("release from id does not exist %s", args.Release.ID))
		} else {
			snapshot, err := createSnapshot(r.DB, args)
			if err != nil {
				return &ReleaseResolver{}, err
			}

			marshalledSnapshot, err := json.Marshal(snapshot)
			if err != nil {
				log.Info(err.Error())
				return nil, err
			}

			forkedRelease := Release{
				ProjectId:     releaseFromId.ProjectId,
				EnvironmentId: releaseFromId.EnvironmentId,
				UserID:        releaseFromId.UserID,
				HeadFeatureID: releaseFromId.HeadFeatureID,
				TailFeatureID: releaseFromId.TailFeatureID,
				State:         plugins.GetState("waiting"),
				StateMessage:  "Release created",
				Snapshot:      postgres.Jsonb{marshalledSnapshot},
			}
			r.DB.Create(&forkedRelease)
			r.ReleaseCreated(&forkedRelease)
			return &ReleaseResolver{DB: r.DB, Release: forkedRelease}, nil
		}
	} else {
		projectId, err := uuid.FromString(args.Release.ProjectId)
		if err != nil {
			log.InfoWithFields("Couldn't parse projectId", log.Fields{
				"args": args,
			})
			return nil, fmt.Errorf("Couldn't parse projectId")
		}
		headFeatureId, err := uuid.FromString(args.Release.HeadFeatureId)
		if err != nil {
			log.InfoWithFields("Couldn't parse headFeatureId", log.Fields{
				"args": args,
			})
			return nil, fmt.Errorf("Couldn't parse headFeatureId")
		}
		environmentId, err := uuid.FromString(args.Release.EnvironmentId)
		if err != nil {
			log.InfoWithFields("Couldn't parse environmentId", log.Fields{
				"args": args,
			})
			return nil, fmt.Errorf("Couldn't parse environmentId")
		}

		// the tail feature id is the current release's head feature id
		if r.DB.Where("state = ? and project_id = ? and environment_id = ?", plugins.GetState("complete"), args.Release.ProjectId, environmentId).Find(&currentRelease).Order("created_at desc").Limit(1).RecordNotFound() {
			// get first ever feature in project if current release doesn't exist yet
			var firstFeature Feature
			if r.DB.Where("project_id = ?", args.Release.ProjectId).Find(&firstFeature).Order("created_at asc").Limit(1).RecordNotFound() {
				log.InfoWithFields("CreateRelease", log.Fields{
					"release": r,
				})
				return nil, fmt.Errorf("No features found.")
			}
			tailFeatureId = firstFeature.ID
		} else {
			tailFeatureId = currentRelease.HeadFeatureID
		}

		userIdString, err := CheckAuth(ctx, []string{})
		if err != nil {
			return &ReleaseResolver{}, err
		}
		userId := uuid.FromStringOrNil(userIdString)

		snapshot, err := createSnapshot(r.DB, args)
		if err != nil {
			return &ReleaseResolver{}, err
		}

		marshalledSnapshot, err := json.Marshal(snapshot)
		if err != nil {
			log.Info(err.Error())
			return nil, err
		}

		release := Release{
			ProjectId:     projectId,
			EnvironmentId: environmentId,
			UserID:        userId,
			HeadFeatureID: headFeatureId,
			TailFeatureID: tailFeatureId,
			State:         plugins.GetState("waiting"),
			StateMessage:  "Release created",
			Snapshot:      postgres.Jsonb{marshalledSnapshot},
		}

		r.DB.Create(&release)
		r.ReleaseCreated(&release)

		return &ReleaseResolver{DB: r.DB, Release: release}, nil
	}
}

// RollbackRelease
func (r *Resolver) RollbackRelease(ctx context.Context, args *struct{ ReleaseId graphql.ID }) (*ReleaseResolver, error) {
	/*
		Rollback's purpose is to deploy a feature with a previous configuration state of the project.
		We find the corresponding release object, get the Snapshot var to get the configuration of the project at the moment
		the release was created. We then create a new release object and insert the old release's info into the new release.
	*/
	release := Release{}
	if r.DB.Where("id = ?", string(args.ReleaseId)).Find(&release).RecordNotFound() {
		errMsg := fmt.Sprintf("Could not find release with given id %s", string(args.ReleaseId))
		log.Info(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	// create new release object with snapshot from found release
	newRelease := Release{
		ProjectId:     release.ProjectId,
		EnvironmentId: release.EnvironmentId,
		UserID:        release.UserID,
		HeadFeatureID: release.HeadFeatureID,
		TailFeatureID: release.TailFeatureID,
		State:         plugins.GetState("waiting"),
		StateMessage:  "Release created and rolled back.",
		Snapshot:      release.Snapshot,
	}
	r.DB.Create(&newRelease)
	r.ReleaseCreated(&newRelease)

	return &ReleaseResolver{DB: r.DB, Release: release}, nil
}

// CreateService Create service
func (r *Resolver) CreateService(args *struct {
	Service *ServiceInput
}) *ServiceResolver {
	return nil
}

// UpdateService Update Service
func (r *Resolver) UpdateService(args *struct {
	Service *ServiceInput
}) *ServiceResolver {
	return nil
}

// DeleteService Delete service
func (r *Resolver) DeleteService(args *struct {
	Service *ServiceInput
}) *ServiceResolver {
	return nil
}

// CreateServiceSpec Create service spec
func (r *Resolver) CreateServiceSpec(args *struct {
	ServiceSpec *ServiceSpecInput
}) *ServiceSpecResolver {
	return nil
}

// UpdateServiceSpec Update service spec
func (r *Resolver) UpdateServiceSpec(args *struct {
	ServiceSpec *ServiceSpecInput
}) *ServiceSpecResolver {
	return nil
}

// DeleteServiceSpec Delete service spec
func (r *Resolver) DeleteServiceSpec(args *struct {
	ServiceSpec *ServiceSpecInput
}) *ServiceSpecResolver {
	return nil
}

// CreateEnvironment Create environment
func (r *Resolver) CreateEnvironment(args *struct {
	Environment *EnvironmentInput
}) *EnvironmentResolver {
	return nil
}

// UpdateEnvironment Update environment
func (r *Resolver) UpdateEnvironment(args *struct {
	Environment *EnvironmentInput
}) *EnvironmentResolver {
	return nil
}

// DeleteEnvironment Delete environment
func (r *Resolver) DeleteEnvironment(args *struct {
	Environment *EnvironmentInput
}) *EnvironmentResolver {
	return nil
}

// CreateEnvironmentVariable Create environment variable
func (r *Resolver) CreateEnvironmentVariable(args *struct {
	EnvironmentVariable *EnvironmentVariableInput
}) *EnvironmentVariableResolver {
	return nil
}

// UpdateEnvironmentVariable Update environment variable
func (r *Resolver) UpdateEnvironmentVariable(args *struct {
	EnvironmentVariable *EnvironmentVariableInput
}) *EnvironmentVariableResolver {
	return nil
}

// DeleteEnvironmentVariable Delete environment variable
func (r *Resolver) DeleteEnvironmentVariable(args *struct {
	EnvironmentVariable *EnvironmentVariableInput
}) *EnvironmentVariableResolver {
	return nil
}

// CreateExtensionSpec Create extension spec
func (r *Resolver) CreateExtensionSpec(args *struct {
	ExtensionSpec *ExtensionSpecInput
}) *ExtensionSpecResolver {
	return nil
}

// UpdateExtensionSpec Update extension spec
func (r *Resolver) UpdateExtensionSpec(args *struct {
	ExtensionSpec *ExtensionSpecInput
}) *ExtensionSpecResolver {
	return nil
}

// DeleteExtensionSpec Delete extension spec
func (r *Resolver) DeleteExtensionSpec(args *struct {
	ExtensionSpec *ExtensionSpecInput
}) *ExtensionSpecResolver {
	return nil
}

// CreateExtension Create extension
func (r *Resolver) CreateExtension(args *struct {
	Extension *ExtensionInput
}) *ExtensionResolver {
	return nil
}

// UpdateExtension Update extension
func (r *Resolver) UpdateExtension(args *struct {
	Extension *ExtensionInput
}) *ExtensionResolver {
	return nil
}

// DeleteExtension Delete extesion
func (r *Resolver) DeleteExtension(args *struct {
	Extension *ExtensionInput
}) *ExtensionResolver {
	return nil
}
