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

func userHasPermission(claims *model.Claims, scopes []string) bool {
	for _, scope := range scopes {
		level := 0
		levels := strings.Count(scope, "/")

		if levels > 0 {
			for level < levels {
				if transistor.SliceContains(scope, claims.Permissions) {
					return true
				}
				scope = scope[0:strings.LastIndexByte(scope, '/')]
				level += 1
			}
		} else {
			if transistor.SliceContains(scope, claims.Permissions) {
				return true
			}
		}
	}
	return false
}

func CheckAuth(ctx context.Context, scopes []string) (string, error) {
	claims, err := getJWTToken(ctx)
	if err != nil {
		return "", err
	}

	if isAdmin(claims) == false {
		if userHasPermission(claims, scopes) == false {
			return claims.UserID, errors.New("You dont have permission to access this resource")
		}
	}

	return claims.UserID, nil
}
