package graphql_resolver

import (
	"context"
	"fmt"

	auth "github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// User Resolver Query
type ProjectResolverQuery struct {
	DB *gorm.DB
}

func (u *ProjectResolverQuery) Project(ctx context.Context, args *struct {
	ID            *graphql.ID
	Slug          *string
	Name          *string
	EnvironmentID *string
}) (*ProjectResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	resolver := ProjectResolver{DBProjectResolver: &db_resolver.ProjectResolver{DB: u.DB}}
	var query *gorm.DB

	var identifier string
	if args.ID != nil {
		query = u.DB.Where("id = ?", *args.ID)
		identifier = string(*args.ID)
	} else if args.Slug != nil {
		query = u.DB.Where("slug = ?", *args.Slug)
		identifier = string(*args.Slug)
	} else if args.Name != nil {
		query = u.DB.Where("name = ?", *args.Name)
		identifier = string(*args.Name)
	} else {
		return nil, fmt.Errorf("Missing argument id or slug")
	}

	if err := query.First(&resolver.DBProjectResolver.Project).Error; err != nil {
		return nil, err
	}

	if args.EnvironmentID != nil {	

		environmentID, err := uuid.FromString(*args.EnvironmentID)
		if err != nil {
			return nil, fmt.Errorf("Environment ID should be of type uuid")
		}

		// check if project has permissions to requested environment
		var permission model.ProjectEnvironment
		if u.DB.Where("project_id = ? AND environment_id = ?", resolver.DBProjectResolver.Project.Model.ID, environmentID).Find(&permission).RecordNotFound() {
			log.InfoWithFields("Environment not found", log.Fields{
				"environment": environmentID,
				"identifier":  identifier,
			})
			return nil, fmt.Errorf("Environment not found")
		}

		// get environment
		if u.DB.Where("id = ?", environmentID).Find(&resolver.DBProjectResolver.Environment).RecordNotFound() {
			log.InfoWithFields("Environment not found", log.Fields{
				"environment": environmentID,
				"identifier":  identifier,
			})
			return nil, fmt.Errorf("Environment not found")
		}

	}

	return &resolver, nil
}

func (u *ProjectResolverQuery) Projects(ctx context.Context, args *struct {
	ProjectSearch *model.ProjectSearchInput
	Params        *model.PaginatorInput
}) (*ProjectListResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	db := u.DB
	if args.ProjectSearch != nil {
		if args.ProjectSearch.Repository != nil && *args.ProjectSearch.Repository != "" {
			db = db.Where("repository like ?", fmt.Sprintf("%%%s%%", *args.ProjectSearch.Repository))
		}

		if args.ProjectSearch.Bookmarked == true {
			var projectBookmarks []model.ProjectBookmark

			// Make a query quickly for bookmarks owned by user
			u.DB.Where("user_id = ?", ctx.Value("jwt").(model.Claims).UserID).Find(&projectBookmarks)

			var projectIds []uuid.UUID
			for _, bookmark := range projectBookmarks {
				projectIds = append(projectIds, bookmark.ProjectID)
			}

			// If the above repository search was included in the paginator input
			// then this should be an "AND" for the where operation
			if len(projectIds) > 0 {
				db = db.Where("id in (?)", projectIds)
			}
		}
	}

	return &ProjectListResolver{
		DBProjectListResolver: &db_resolver.ProjectListResolver{
			DB:             db.Order("name asc"),
			PaginatorInput: args.Params,
		},
	}, nil
}
