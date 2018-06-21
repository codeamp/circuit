package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

var authDisabled bool
var enabledMap = map[bool]string{
	true:  "ENABLED",
	false: "DISABLED",
}

func SetAuthEnabled(enabled bool) {
	log.Warn(fmt.Sprintf("AUTHENTICATION %s", enabledMap[enabled]))
	authDisabled = !enabled
}

func checkAuthentication(ctx context.Context, scopes []string) (string, error) {
	claims := ctx.Value("jwt").(model.Claims)

	if claims.UserID == "" {
		return "", errors.New(claims.TokenError)
	}

	if transistor.SliceContains("admin", claims.Permissions) {
		return claims.UserID, nil
	}

	if len(scopes) == 0 {
		return claims.UserID, nil
	} else {
		for _, scope := range scopes {
			level := 0
			levels := strings.Count(scope, "/")

			if levels > 0 {
				for level < levels {
					if transistor.SliceContains(scope, claims.Permissions) {
						return claims.UserID, nil
					}
					scope = scope[0:strings.LastIndexByte(scope, '/')]
					level += 1
				}
			} else {
				if transistor.SliceContains(scope, claims.Permissions) {
					return claims.UserID, nil
				}
			}
		}
		return claims.UserID, errors.New("you dont have permission to access this resource")
	}
}

func CheckAuth(ctx context.Context, scopes []string) (string, error) {
	if authDisabled {
		return passAuthentication(ctx, scopes)
	}

	return checkAuthentication(ctx, scopes)
}

func passAuthentication(ctx context.Context, scopes []string) (string, error) {
	return "Testing Mode", nil
}
