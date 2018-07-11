package graphql_resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	oidc "github.com/coreos/go-oidc"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type Middleware struct {
	Resolver *Resolver
}

func (middleware *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := model.Claims{}
		authString := r.Header.Get("Authorization")
		ctx := context.Context(context.Background())

		if len(authString) < 8 {
			claims.TokenError = "invalid access token"
			log.Error(claims.TokenError)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		bearerToken := authString[7:]

		// Initialize a provider by specifying dex's issuer URL.
		provider, err := oidc.NewProvider(ctx, viper.GetString("plugins.codeamp.oidc_uri"))
		if err != nil {
			claims.TokenError = err.Error()
			log.Error(claims.TokenError)
			//w.WriteHeader(http.StatusForbidden)

			// ADB There needs to be a better way to deal with this
			// when testing outside of a dex environment
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		// Create an ID token parser, but only trust ID tokens issued to "example-app"
		idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: viper.GetString("plugins.codeamp.oidc_client_id")})

		idToken, err := idTokenVerifier.Verify(ctx, bearerToken)
		if err != nil {
			claims.TokenError = fmt.Sprintf("could not verify bearer token: %v", err.Error())
			log.Error(claims.TokenError)
			w.Header().Set("Www-Authenticate", "Bearer token_type=\"JWT\"")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		if err := idToken.Claims(&claims); err != nil {
			log.Error(claims.TokenError)
			claims.TokenError = fmt.Sprintf("failed to parse claims: %v", err.Error())
			w.Header().Set("Www-Authenticate", "Bearer token_type=\"JWT\"")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		if !claims.Verified {
			log.Error(claims.TokenError)
			claims.TokenError = fmt.Sprintf("email (%q) in returned claims was not verified", claims.Email)
			w.Header().Set("Www-Authenticate", "Bearer token_type=\"JWT\"")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		c, err := middleware.Resolver.Redis.Get(fmt.Sprintf("%s_%s", idToken.Nonce, claims.Email)).Result()
		if err == redis.Nil {
			user := model.User{}
			if middleware.Resolver.DB.Where("email = ?", claims.Email).Find(&user).RecordNotFound() {
				user.Email = claims.Email
				middleware.Resolver.DB.Create(&user)
			}

			middleware.Resolver.DB.Model(&user).Association("Permissions").Find(&user.Permissions)

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
				log.Error(err)
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
				return
			}

			err = middleware.Resolver.Redis.Set(fmt.Sprintf("%s_%s", idToken.Nonce, claims.Email), serializedClaims, 24*time.Hour).Err()
			if err != nil {
				log.Error(err)
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
				return
			}
		} else if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			err := json.Unmarshal([]byte(c), &claims)
			if err != nil {
				log.Error(err)
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
	})
}

func (middleware *Middleware) Cors(next http.Handler) http.Handler {
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
