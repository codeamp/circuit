package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/transistor"
	oidc "github.com/coreos/go-oidc"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserId     string   `json:"userId"`
	Email      string   `json:"email"`
	Verified   bool     `json:"email_verified"`
	Groups     []string `json:"groups"`
	TokenError string   `json:"tokenError"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func AuthMiddleware(next http.Handler, db *gorm.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := Claims{}
		authString := r.Header.Get("Authorization")
		ctx := context.Context(context.Background())

		if len(authString) < 8 {
			claims.TokenError = "invalid access token"
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		bearerToken := authString[7:len(authString)]

		// Initialize a provider by specifying dex's issuer URL.
		provider, err := oidc.NewProvider(ctx, viper.GetString("plugins.codeamp.oidc_uri"))
		if err != nil {
			claims.TokenError = err.Error()
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		// Create an ID token parser, but only trust ID tokens issued to "example-app"
		idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: "example-app"})

		idToken, err := idTokenVerifier.Verify(ctx, bearerToken)
		if err != nil {
			claims.TokenError = fmt.Sprintf("could not verify bearer token: %v", err.Error())
			w.Header().Set("Www-Authenticate", "Bearer token_type=\"JWT\"")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		if err := idToken.Claims(&claims); err != nil {
			claims.TokenError = fmt.Sprintf("failed to parse claims: %v", err.Error())
			w.Header().Set("Www-Authenticate", "Bearer token_type=\"JWT\"")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		if !claims.Verified {
			claims.TokenError = fmt.Sprintf("email (%q) in returned claims was not verified", claims.Email)
			w.Header().Set("Www-Authenticate", "Bearer token_type=\"JWT\"")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		user := models.User{}
		if db.Where("email = ?", claims.Email).Find(&user).RecordNotFound() {
			user.Email = claims.Email
			db.Create(&user)

			userPermission := models.UserPermission{
				UserId: user.Model.ID,
				Value:  "admin",
			}
			db.Create(&userPermission)
		}

		claims.UserId = user.ID.String()

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow cross domain AJAX requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Cache-Control, Content-Language, Content-Type, Expires, Last-Modified, Pragma, WWW-Authenticate")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		if r.Method == "OPTIONS" {
			//handle preflight in here
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func CheckAuth(ctx context.Context, scopes []string) (string, error) {
	claims := ctx.Value("jwt").(Claims)

	if claims.UserId == "" {
		return "", errors.New(claims.TokenError)
	}

	if len(scopes) == 0 {
		return claims.UserId, nil
	} else {
		for _, scope := range scopes {
			if transistor.SliceContains(scope, claims.Groups) {
				return claims.UserId, nil
			}
		}
		return "", errors.New("you dont have permission to access this resource")
	}

	return claims.UserId, nil
}
