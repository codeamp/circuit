package db_resolver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type Resolver struct {
	DB *gorm.DB
}

// User
func (r *Resolver) User(ctx context.Context, args *struct {
	ID *graphql.ID
}) (*UserResolver, error) {
	var userID string
	var err error
	var user model.User

	if args.ID != nil {
		userID = string(*args.ID)
	} else {
		claims := ctx.Value("jwt").(model.Claims)
		userID = claims.UserID
	}

	if _, err = auth.CheckAuth(ctx, []string{fmt.Sprintf("user/%s", userID)}); err != nil {
		return nil, err
	}

	if err = r.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &UserResolver{DB: r.DB, User: user}, nil
}

// Users
func (r *Resolver) Users(ctx context.Context) ([]*UserResolver, error) {
	var rows []model.User
	var results []*UserResolver

	r.DB.Order("created_at desc").Find(&rows)

	for _, user := range rows {
		results = append(results, &UserResolver{DB: r.DB, User: user})
	}

	return results, nil
}

// Project
func (r *Resolver) Project(ctx context.Context, args *struct {
	ID            *graphql.ID
	Slug          *string
	Name          *string
	EnvironmentID *string
}) (*ProjectResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var project model.Project
	var environment model.Environment
	var query *gorm.DB

	if args.ID != nil {
		query = r.DB.Where("id = ?", *args.ID)
	} else if args.Slug != nil {
		query = r.DB.Where("slug = ?", *args.Slug)
	} else if args.Name != nil {
		query = r.DB.Where("name = ?", *args.Name)
	} else {
		return nil, fmt.Errorf("Missing argument id or slug")
	}

	if err := query.First(&project).Error; err != nil {
		return nil, err
	}

	if args.EnvironmentID == nil {
		return nil, fmt.Errorf("Missing environment id")
	}

	environmentID, err := uuid.FromString(*args.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Environment ID should be of type uuid")
	}

	// check if project has permissions to requested environment
	var permission model.ProjectEnvironment
	if r.DB.Where("project_id = ? AND environment_id = ?", project.Model.ID, environmentID).Find(&permission).RecordNotFound() {
		log.InfoWithFields("Environment not found", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("Environment not found")
	}

	// get environment
	if r.DB.Where("id = ?", *args.EnvironmentID).Find(&environment).RecordNotFound() {
		log.InfoWithFields("Environment not found", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("Environment not found")
	}

	return &ProjectResolver{DB: r.DB, Project: project, Environment: environment}, nil
}

// Projects
func (r *Resolver) Projects(ctx context.Context, args *struct {
	ProjectSearch *model.ProjectSearchInput
	Params        *model.PaginatorInput
}) (*ProjectListResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.Project
	var query *gorm.DB

	if args.ProjectSearch != nil && args.ProjectSearch.Repository != nil {

		r.DB.Where("repository like ?", fmt.Sprintf("%%%s%%", *args.ProjectSearch.Repository)).Find(&rows)

	} else {
		var projectBookmarks []model.ProjectBookmark

		r.DB.Where("user_id = ?", ctx.Value("jwt").(model.Claims).UserID).Find(&projectBookmarks)

		var projectIds []uuid.UUID
		for _, bookmark := range projectBookmarks {
			projectIds = append(projectIds, bookmark.ProjectID)
		}
		query = r.DB.Where("id in (?)", projectIds)
	}

	return &ProjectListResolver{
		Query:          query,
		PaginatorInput: args.Params,
	}, nil
}
func (r *Resolver) Features(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (*FeatureListResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	query := r.DB.Order("created_at desc")

	return &FeatureListResolver{
		Query:          query,
		PaginatorInput: args.Params,
	}, nil
}

func (r *Resolver) Services(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (*ServiceListResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	query := r.DB.Order("created_at desc")

	return &ServiceListResolver{
		Query:          query,
		PaginatorInput: args.Params,
	}, nil
}

func (r *Resolver) ServiceSpecs(ctx context.Context) ([]*ServiceSpecResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.ServiceSpec
	var results []*ServiceSpecResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, serviceSpec := range rows {
		results = append(results, &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec})
	}

	return results, nil
}

func (r *Resolver) Releases(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (*ReleaseListResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	return &ReleaseListResolver{
		PaginatorInput: args.Params,
	}, nil
}

func (r *Resolver) Environments(ctx context.Context, args *struct{ ProjectSlug *string }) ([]*EnvironmentResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var environments []model.Environment
	var results []*EnvironmentResolver

	if args.ProjectSlug != nil {
		var project model.Project
		var permissions []model.ProjectEnvironment

		if err := r.DB.Where("slug = ?", *args.ProjectSlug).First(&project).Error; err != nil {
			return nil, err
		}

		r.DB.Where("project_id = ?", project.Model.ID).Find(&permissions)
		for _, permission := range permissions {
			var environment model.Environment
			r.DB.Where("id = ?", permission.EnvironmentID).Find(&environment)
			results = append(results, &EnvironmentResolver{DB: r.DB, Environment: environment})
		}

		return results, nil
	}

	r.DB.Order("created_at desc").Find(&environments)
	for _, environment := range environments {
		results = append(results, &EnvironmentResolver{DB: r.DB, Environment: environment})
	}

	return results, nil
}

func (r *Resolver) Secrets(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (*SecretListResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{"admin"}); err != nil {
		return nil, err
	}

	query := r.DB.Where("scope != ?", "project").Order("created_at desc")

	return &SecretListResolver{
		PaginatorInput: args.Params,
		Query:          query,
	}, nil
}

func (r *Resolver) Extensions(ctx context.Context, args *struct{ EnvironmentID *string }) ([]*ExtensionResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.Extension
	var results []*ExtensionResolver

	if args.EnvironmentID != nil {
		r.DB.Where("extensions.environment_id = ?", args.EnvironmentID).Order(`
			CASE extensions.type
				WHEN 'workflow' THEN 1
				WHEN 'deployment' THEN 2
				ELSE 3
			END, extensions.key ASC`).Find(&rows)
	} else {
		r.DB.Order(`
			CASE extensions.type
				WHEN 'workflow' THEN 1
				WHEN 'deployment' THEN 2
				ELSE 3
			END, extensions.key ASC`).Find(&rows)
	}

	for _, ext := range rows {
		results = append(results, &ExtensionResolver{DB: r.DB, Extension: ext})
	}

	return results, nil
}

func (r *Resolver) ProjectExtensions(ctx context.Context) ([]*ProjectExtensionResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.ProjectExtension
	var results []*ProjectExtensionResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, extension := range rows {
		results = append(results, &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension})
	}

	return results, nil
}

func (r *Resolver) ReleaseExtensions(ctx context.Context) ([]*ReleaseExtensionResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.ReleaseExtension
	var results []*ReleaseExtensionResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, releaseExtension := range rows {
		results = append(results, &ReleaseExtensionResolver{DB: r.DB, ReleaseExtension: releaseExtension})
	}

	return results, nil
}

// Permissions
func (r *Resolver) Permissions(ctx context.Context) (model.JSON, error) {
	var rows []model.UserPermission
	var results = make(map[string]bool)

	r.DB.Unscoped().Select("DISTINCT(value)").Find(&rows)

	for _, userPermission := range rows {
		if _, err := auth.CheckAuth(ctx, []string{userPermission.Value}); err != nil {
			results[userPermission.Value] = false
		} else {
			results[userPermission.Value] = true
		}
	}

	bytes, err := json.Marshal(results)
	if err != nil {
		return model.JSON{}, err
	}

	return model.JSON{bytes}, nil
}
