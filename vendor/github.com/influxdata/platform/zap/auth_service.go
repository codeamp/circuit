package zap

import (
	"context"

	"go.uber.org/zap"

	"github.com/influxdata/platform"
)

var _ platform.AuthorizationService = (*AuthorizationService)(nil)

// AuthorizationService manages authorizations.
type AuthorizationService struct {
	Logger               *zap.Logger
	AuthorizationService platform.AuthorizationService
}

// FindAuthorizationByID returns an authorization given an id, and logs any errors.
func (s *AuthorizationService) FindAuthorizationByID(ctx context.Context, id platform.ID) (a *platform.Authorization, err error) {
	defer func() {
		if err != nil {
			s.Logger.Info("error finding authorization by id", zap.Error(err))
		}
	}()

	return s.AuthorizationService.FindAuthorizationByID(ctx, id)
}

// FindAuthorizationByToken returns an authorization given a token, and logs any errors.
func (s *AuthorizationService) FindAuthorizationByToken(ctx context.Context, t string) (a *platform.Authorization, err error) {
	defer func() {
		if err != nil {
			s.Logger.Info("error finding authorization by token", zap.Error(err))
		}
	}()

	return s.AuthorizationService.FindAuthorizationByToken(ctx, t)
}

// FindAuthorizations returns authorizations given a filter, and logs any errors.
func (s *AuthorizationService) FindAuthorizations(ctx context.Context, filter platform.AuthorizationFilter, opt ...platform.FindOptions) (as []*platform.Authorization, i int, err error) {
	defer func() {
		if err != nil {
			s.Logger.Info("error finding authorizations", zap.Error(err))
		}
	}()

	return s.AuthorizationService.FindAuthorizations(ctx, filter, opt...)
}

// CreateAuthorization creates an authorization, and logs any errors.
func (s *AuthorizationService) CreateAuthorization(ctx context.Context, a *platform.Authorization) (err error) {
	defer func() {
		if err != nil {
			s.Logger.Info("error creating authorization", zap.Error(err))
		}
	}()

	return s.AuthorizationService.CreateAuthorization(ctx, a)
}

// DeleteAuthorization deletes an authorization, and logs any errors.
func (s *AuthorizationService) DeleteAuthorization(ctx context.Context, id platform.ID) (err error) {
	defer func() {
		if err != nil {
			s.Logger.Info("error deleting authorization", zap.Error(err))
		}
	}()

	return s.AuthorizationService.DeleteAuthorization(ctx, id)
}

// SetAuthorizationStatus updates an authorization's status and logs any errors.
func (s *AuthorizationService) SetAuthorizationStatus(ctx context.Context, id platform.ID, status platform.Status) (err error) {
	defer func() {
		if err != nil {
			s.Logger.Info("error updating authorization", zap.Error(err))
		}
	}()

	return s.AuthorizationService.SetAuthorizationStatus(ctx, id, status)
}
