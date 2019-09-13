package atlas

import (
	"golang.org/x/time/rate"

	"github.com/Clever/atlas-api-client/digestauth"
	"github.com/Clever/atlas-api-client/gen-go/client"
	"github.com/Clever/atlas-api-client/limits"
)

// New creates an Atlas client give the username, password, and URL.
func New(atlasUsername, atlasPassword, atlasURL string) *client.WagClient {
	atlasAPI := client.New(atlasURL)
	digestT := digestauth.NewTransport(atlasUsername, atlasPassword)
	atlasAPI.SetTransport(limits.NewRateLimitedRoundTripper(&digestT, rate.NewLimiter(limits.AtlasRateLimit, 5)))
	return atlasAPI
}
