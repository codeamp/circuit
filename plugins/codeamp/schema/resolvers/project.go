package resolvers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	log "github.com/codeamp/logger"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/extemporalgenome/slug"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	"golang.org/x/crypto/ssh"
)

type ProjectInput struct {
	ID            *string
	GitProtocol   string
	GitBranch     *string
	GitUrl        string
	Bookmarked    *bool
	EnvironmentId string
}

func (r *Resolver) Project(ctx context.Context, args *struct {
	ID            *graphql.ID
	Slug          *string
	Name          *string
	EnvironmentId *string
}) (*ProjectResolver, error) {
	var project models.Project
	var environment models.Environment
	var query *gorm.DB

	// get environment
	if r.db.Where("id = ?", *args.EnvironmentId).Find(&environment).RecordNotFound() {
		log.InfoWithFields("Environment doesn't exist.", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("Environment doesn't exist.")
	}

	if args.ID != nil {
		query = r.db.Where("id = ?", *args.ID)
	} else if args.Slug != nil {
		query = r.db.Where("slug = ?", *args.Slug)
	} else if args.Name != nil {
		query = r.db.Where("name = ?", *args.Name)
	} else {
		return nil, fmt.Errorf("Missing argument id or slug")
	}

	if err := query.First(&project).Error; err != nil {
		return nil, err
	}

	return &ProjectResolver{db: r.db, Project: project, Environment: environment}, nil
}

func (r *Resolver) UpdateProject(args *struct{ Project *ProjectInput }) (*ProjectResolver, error) {
	if args.Project.ID == nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	var project *models.Project
	if r.db.Where("id = ?", args.Project.ID).First(project).RecordNotFound() {
		log.InfoWithFields("Project not found", log.Fields{
			"id": args.Project.ID,
		})
		return nil, fmt.Errorf("Project not found.")
	}

	project, err := r.cleanRepoInfo(args)
	if err != nil {
		return nil, err
	}

	r.db.Save(project)
	return &ProjectResolver{db: r.db, Project: *project}, nil
}

/*
ResetProject removes all project-related objects as well as
updating the project's GitUrl if the user has selected a different
Repository Type
*/
func (r *Resolver) ResetProject(args *struct{ Project *ProjectInput }) (*ProjectResolver, error) {
	project, err := r.cleanRepoInfo(args)
	if err != nil {
		return nil, err
	}

	r.db.Save(project)

	// Cascade delete all features and releases related to old git url
	r.db.Where("project_id = ?", project.ID).Delete(models.Feature{})
	r.db.Where("project_id = ?", project.ID).Delete(models.Release{})

	return &ProjectResolver{db: r.db, Project: *project}, nil
}

func (r *Resolver) cleanRepoInfo(args *struct{ Project *ProjectInput }) (*models.Project, error) {
	var project models.Project
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
	project.GitProtocol = protocol
	project.Repository = repository
	project.Name = repository
	project.Slug = slug.Slug(repository)
	return &project, nil
}

func (r *Resolver) CreateProject(args *struct{ Project *ProjectInput }) (*ProjectResolver, error) {
	project, err := r.cleanRepoInfo(args)
	if err != nil {
		return nil, err
	}

	// Check if project already exists with same name
	if r.db.Where("repository = ?", project.Repository).First(&models.Project{}).RecordNotFound() == false {
		return nil, fmt.Errorf("Project with repository name already exists.")
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
	r.db.Create(&project)
	environments := []models.Environment{}
	if r.db.Find(&environments).RecordNotFound() {
		log.InfoWithFields("Environment doesn't exist.", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("No environments initialized.")
	}

	for _, env := range environments {
		r.db.Create(&models.ProjectSettings{
			EnvironmentId: env.Model.ID,
			ProjectId:     project.Model.ID,
			GitBranch:     "master",
		})
	}

	return &ProjectResolver{db: r.db, Project: *project, Environment: environments[0]}, nil
}

type ProjectResolver struct {
	db          *gorm.DB
	Project     models.Project
	Environment models.Environment
}

func (r *ProjectResolver) ID() graphql.ID {
	return graphql.ID(r.Project.Model.ID.String())
}

func (r *ProjectResolver) Name() string {
	return r.Project.Name
}

func (r *ProjectResolver) Slug() string {
	return r.Project.Slug
}

func (r *ProjectResolver) Repository() string {
	return r.Project.Repository
}

func (r *ProjectResolver) Secret() string {
	return r.Project.Secret
}

func (r *ProjectResolver) GitUrl() string {
	return r.Project.GitUrl
}

func (r *ProjectResolver) GitProtocol() string {
	return r.Project.GitProtocol
}

func (r *ProjectResolver) RsaPrivateKey() string {
	return r.Project.RsaPrivateKey
}

func (r *ProjectResolver) RsaPublicKey() string {
	return r.Project.RsaPublicKey
}

func (r *ProjectResolver) CurrentRelease() (*ReleaseResolver, error) {
	var currentRelease models.Release

	if r.db.Where("state = ? and project_id = ? and environment_id = ?", plugins.GetState("complete"), r.Project.Model.ID, r.Environment.Model.ID).Order("created_at desc").First(&currentRelease).RecordNotFound() {
		log.InfoWithFields("CurrentRelease does not exist", log.Fields{
			"state":          plugins.GetState("complete"),
			"project_id":     r.Project.Model.ID,
			"environment_id": r.Environment.Model.ID,
		})
		return &ReleaseResolver{db: r.db, Release: currentRelease}, fmt.Errorf("Current release does not exist.")
	}
	return &ReleaseResolver{db: r.db, Release: currentRelease}, nil
}

func (r *ProjectResolver) Features(ctx context.Context) ([]*FeatureResolver, error) {
	var rows []models.Feature
	var results []*FeatureResolver

	r.db.Where("project_id = ? and ref = ?", r.Project.ID, fmt.Sprintf("refs/heads/%s", r.GitBranch())).Order("created desc").Find(&rows)

	for _, feature := range rows {
		results = append(results, &FeatureResolver{db: r.db, Feature: feature})
	}

	return results, nil
}

func (r *ProjectResolver) Services(ctx context.Context) ([]*ServiceResolver, error) {
	var rows []models.Service
	var results []*ServiceResolver

	r.db.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Find(&rows)

	for _, service := range rows {
		results = append(results, &ServiceResolver{db: r.db, Service: service})
	}

	return results, nil
}

func (r *ProjectResolver) Releases(ctx context.Context) ([]*ReleaseResolver, error) {
	var rows []models.Release
	var results []*ReleaseResolver
	r.db.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Order("created_at desc").Find(&rows)
	for _, release := range rows {
		results = append(results, &ReleaseResolver{db: r.db, Release: release})
	}
	return results, nil
}

func (r *ProjectResolver) EnvironmentVariables(ctx context.Context) ([]*EnvironmentVariableResolver, error) {
	var rows []models.EnvironmentVariable
	var results []*EnvironmentVariableResolver
	r.db.Select("key, id, created_at, type, project_id, environment_id, deleted_at, is_secret").Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Order("key, created_at desc").Find(&rows)
	for _, envVar := range rows {
		results = append(results, &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar})
	}
	return results, nil
}

func (r *ProjectResolver) Extensions(ctx context.Context) ([]*ExtensionResolver, error) {
	var rows []models.Extension
	var results []*ExtensionResolver
	r.db.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Find(&rows)
	for _, extension := range rows {
		results = append(results, &ExtensionResolver{db: r.db, Extension: extension})
	}
	return results, nil
}

func (r *ProjectResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Project.Model.CreatedAt}
}

func (r *ProjectResolver) GitBranch() string {
	var projectSettings models.ProjectSettings
	if r.db.Where("project_id = ? and environment_id = ?", r.Project.Model.ID.String(), r.Environment.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		// if no project settings is found, we should create an
		// entry for its environment with default settings .e.g branch => "master"
		defaultProjectSettings := models.ProjectSettings{
			EnvironmentId: r.Environment.Model.ID,
			ProjectId:     r.Project.Model.ID,
			GitBranch:     "master",
		}
		r.db.Create(&defaultProjectSettings)
		return defaultProjectSettings.GitBranch
	} else {
		return projectSettings.GitBranch
	}
}
