package graphql_resolver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/extemporalgenome/slug"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
)

// Project Resolver Mutation
type ProjectResolverMutation struct {
	DB *gorm.DB
}

// CreateProject Create project
func (r *ProjectResolverMutation) CreateProject(ctx context.Context, args *struct {
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

	// Check if project exists in github
	if protocol == "HTTPS" && !strings.Contains(args.Project.GitUrl, "@") { // now only check for public repo
		resp, err := http.Get(args.Project.GitUrl)
		if err != nil || resp.StatusCode != 200 {
			fmt.Printf("#############resp.StatusCode: %d\n", resp.StatusCode)
			return nil, fmt.Errorf("This repository does not exist on %s. Try again with a different git url.", res["host"])
		}
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

	// Create git branch for env per env
	environments := []model.Environment{}
	r.DB.Find(&environments)
	if len(environments) == 0 {
		log.InfoWithFields("No envs found.", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("No envs found")
	}

	r.DB.Create(&project)

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
func (r *ProjectResolverMutation) UpdateProject(ctx context.Context, args *struct {
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

		oldBranchName := ""
		var projectSettings model.ProjectSettings
		if r.DB.Where("environment_id = ? and project_id = ?", environmentID, projectID).First(&projectSettings).RecordNotFound() {
			projectSettings.EnvironmentID = environmentID
			projectSettings.ProjectID = projectID
			projectSettings.GitBranch = *args.Project.GitBranch
			projectSettings.ContinuousDeploy = *args.Project.ContinuousDeploy
		} else {
			oldBranchName = projectSettings.GitBranch
			projectSettings.GitBranch = *args.Project.GitBranch
			projectSettings.ContinuousDeploy = *args.Project.ContinuousDeploy
		}

		_userID, err := auth.CheckAuth(ctx, []string{})
		if err != nil {
			return nil, err
		}

		userID, err := uuid.FromString(_userID)
		if err != nil {
			return nil, err
		}

		log.WarnWithFields("[AUDIT] Updated Project Branch", log.Fields{
			"project":     project.Slug,
			"branch":      *args.Project.GitBranch,
			"oldBranch":   oldBranchName,
			"user":        userID,
			"environment": environmentID},
		)

		// Save after all error conditions have passed
		if err = r.DB.Save(&projectSettings).Error; err != nil {
			log.Error(err)
			return nil, err
		}
	}

	if err := r.DB.Save(&project).Error; err != nil {
		log.Error(err)
		return nil, err
	}

	return &ProjectResolver{DBProjectResolver: &db_resolver.ProjectResolver{DB: r.DB, Project: project}}, nil
}

func (r *ProjectResolverMutation) BookmarkProject(ctx context.Context, args *struct{ ID graphql.ID }) (bool, error) {
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
		r.DB.Unscoped().Delete(&projectBookmark)
		return false, nil
	}
}

// UpdateProjectEnvironments
func (r *ProjectResolverMutation) UpdateProjectEnvironments(ctx context.Context, args *struct {
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
