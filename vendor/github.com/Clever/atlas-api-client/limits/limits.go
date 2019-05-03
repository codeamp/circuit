package limits

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// AtlasRateLimit is the rate limit on the Atlas API.
// https://docs.atlas.mongodb.com/api/
var AtlasRateLimit = rate.Every(time.Minute / 100)

// RateLimitedRoundTripper imposes a rate limit on requests made with it.
type RateLimitedRoundTripper struct {
	http.RoundTripper
	limiter *rate.Limiter
}

var _ http.RoundTripper = RateLimitedRoundTripper{}

func NewRateLimitedRoundTripper(tripper http.RoundTripper, limiter *rate.Limiter) RateLimitedRoundTripper {
	return RateLimitedRoundTripper{
		RoundTripper: tripper,
		limiter:      limiter,
	}
}

func (r RateLimitedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	err := r.limiter.Wait(req.Context())
	if err != nil {
		return nil, fmt.Errorf("rate limit: %s", err.Error())
	}
	return r.RoundTripper.RoundTrip(req)
}
