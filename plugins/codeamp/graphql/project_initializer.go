package graphql_resolver

import (
	"context"
	"fmt"

	auth "github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// User Resolver Initializer
type ProjectResolverInitializer struct {
	DB *gorm.DB
}

func (u *ProjectResolverInitializer) Project(ctx context.Context, identifier map[string]string, environmentID uuid.UUID) (*ProjectResolver, error) {
	log.Info("project_initializer.go project")
	log.Info(environmentID)
	log.Info(identifier)

	resolver := ProjectResolver{DBProjectResolver: &db_resolver.ProjectResolver{DB: u.DB}}
	var query *gorm.DB

	if _, ok := identifier["ID"]; ok {
		query = u.DB.Where("id = ?", identifier["ID"])
	} else if _, ok := identifier["Slug"]; ok {
		query = u.DB.Where("slug = ?", identifier["Slug"])
	} else if _, ok := identifier["Name"]; ok {
		query = u.DB.Where("name = ?", identifier["Name"])
	} else {
		return nil, fmt.Errorf("Missing argument id or slug")
	}

	if err := query.First(&resolver.DBProjectResolver.Project).Error; err != nil {
		return nil, err
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

	return &resolver, nil
}

func (u *ProjectResolverInitializer) Projects(ctx context.Context, projectSearchInput *model.ProjectSearchInput) ([]*ProjectResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.Project
	var results []*ProjectResolver
	if projectSearchInput.Repository != nil {
		u.DB.Where("repository like ?", fmt.Sprintf("%%%s%%", *projectSearchInput.Repository)).Find(&rows)
	} else {
		var projectBookmarks []model.ProjectBookmark

		u.DB.Where("user_id = ?", ctx.Value("jwt").(model.Claims).UserID).Find(&projectBookmarks)

		var projectIds []uuid.UUID
		for _, bookmark := range projectBookmarks {
			projectIds = append(projectIds, bookmark.ProjectID)
		}
		u.DB.Where("id in (?)", projectIds).Find(&rows)
	}

	for _, project := range rows {
		results = append(results, &ProjectResolver{DBProjectResolver: &db_resolver.ProjectResolver{DB: u.DB, Project: project}})
	}

	return results, nil
}
