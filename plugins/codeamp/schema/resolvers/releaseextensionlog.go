package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type ReleaseExtensionLogResolver struct {
	db                  *gorm.DB
	ReleaseExtensionLog models.ReleaseExtensionLog
}

func (r *Resolver) ReleaseExtensionLog(ctx context.Context, args *struct{ ID graphql.ID }) (*ReleaseExtensionLogResolver, error) {
	releaseExtensionLog := models.ReleaseExtensionLog{}
	if err := r.db.Where("id = ?", args.ID).First(&releaseExtensionLog).Error; err != nil {
		return nil, err
	}

	return &ReleaseExtensionLogResolver{db: r.db, ReleaseExtensionLog: releaseExtensionLog}, nil
}

func (r *ReleaseExtensionLogResolver) ID() graphql.ID {
	return graphql.ID(r.ReleaseExtensionLog.Model.ID.String())
}

func (r *ReleaseExtensionLogResolver) Msg() string {
	return r.ReleaseExtensionLog.Msg
}

func (r *ReleaseExtensionLogResolver) ReleaseExtension(ctx context.Context) (*ReleaseExtensionResolver, error) {
	var releaseExtension models.ReleaseExtension
	r.db.Model(r.ReleaseExtensionLog).Related(&releaseExtension)
	return &ReleaseExtensionResolver{db: r.db, ReleaseExtension: releaseExtension}, nil
}
