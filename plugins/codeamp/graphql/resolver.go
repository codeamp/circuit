package graphql_resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	oidc "github.com/coreos/go-oidc"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
)

// Resolver is the main resolver for all queries
type Resolver struct {
	// DB
	DB *gorm.DB
	// Events
	Events chan transistor.Event
	// Redis
	Redis *redis.Client
}

func (resolver *Resolver) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := db_resolver.Claims{}
		authString := r.Header.Get("Authorization")
		ctx := context.Context(context.Background())

		if len(authString) < 8 {
			claims.TokenError = "invalid access token"
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		bearerToken := authString[7:]

		// Initialize a provider by specifying dex's issuer URL.
		provider, err := oidc.NewProvider(ctx, viper.GetString("plugins.codeamp.oidc_uri"))
		if err != nil {
			claims.TokenError = err.Error()
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		// Create an ID token parser, but only trust ID tokens issued to "example-app"
		idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: viper.GetString("plugins.codeamp.oidc_client_id")})

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

		c, err := resolver.Redis.Get(fmt.Sprintf("%s_%s", idToken.Nonce, claims.Email)).Result()
		if err == redis.Nil {
			user := User{}
			if resolver.DB.Where("email = ?", claims.Email).Find(&user).RecordNotFound() {
				user.Email = claims.Email
				resolver.DB.Create(&user)
			}

			resolver.DB.Model(&user).Association("Permissions").Find(&user.Permissions)

			var permissions []string
			for _, permission := range user.Permissions {
				permissions = append(permissions, permission.Value)
			}

			// Add user scope
			permissions = append(permissions, fmt.Sprintf("user/%s", user.ID.String()))
			claims.UserID = user.ID.String()
			claims.Permissions = permissions

			serializedClaims, err := json.Marshal(claims)
			if err != nil {
				log.Panic(err)
			}

			err = resolver.Redis.Set(fmt.Sprintf("%s_%s", idToken.Nonce, claims.Email), serializedClaims, 24*time.Hour).Err()
			if err != nil {
				log.Panic(err)
			}
		} else if err != nil {
			log.Panic(err)
		} else {
			err := json.Unmarshal([]byte(c), &claims)
			if err != nil {
				log.Panic(err)
			}
		}

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
