package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/transistor"
)

func getJWTToken(ctx context.Context) (*model.Claims, error) {
	if ctx == nil || ctx.Value("jwt") == nil {
		return nil, fmt.Errorf("No JWT was provided")
	}
	claims := ctx.Value("jwt").(model.Claims)
	if claims.UserID == "" {
		return nil, errors.New(claims.TokenError)
	}

	return &claims, nil
}

func isAdmin(claims *model.Claims) bool {
	return transistor.SliceContains("admin", claims.Permissions)
}

func hasScopePermission(scope string, permissions []string) bool {
	// Loop through each permission and see if it is a prefix
	// of the scope we desire access to
	for _, permission := range permissions {
		if strings.HasPrefix(scope, permission) {
			return true
		}
	}

	return false
}

func userHasPermission(claims *model.Claims, scopes []string) bool {
	// Loop through each scope and hand off to ask if scope has permission
	for _, scope := range scopes {
		if hasScopePermission(scope, claims.Permissions) == false {
			return false
		}
	}

	// If we made it this far and we haven't bailed then the above has found the necessary permissions
	// OR there were no scopes provided in which case it should return true
	return true
}

func CheckAuth(ctx context.Context, scopes []string) (string, error) {
	claims, err := getJWTToken(ctx)
	if err != nil {
		return "", err
	}

	if isAdmin(claims) == false {
		if userHasPermission(claims, scopes) == false {
			return claims.UserID, errors.New("You don't have permission to access this resource")
		}
	}

	return claims.UserID, nil
}
