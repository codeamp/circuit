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

func hasScopePermission(claims *model.Claims, scope string) bool {
	for _, scope := range strings.Split(scope, "/") {
		if transistor.SliceContains(scope, claims.Permissions) {
			return true
		}
	}

	return false
}

func userHasPermission(claims *model.Claims, scopes []string) bool {
	for _, scope := range scopes {
		if hasScopePermission(claims, scope) == true {
			return true
		}
	}

	// If they don't have any scope, then return true
	return len(scopes) == 0
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
