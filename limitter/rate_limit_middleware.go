package limitter

import (
	"net"
	"strconv"
	"time"

	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/etcinit/speedbump"
	"github.com/hellofresh/janus/errors"
)

// RateLimitMiddleware prevents requests to an API from exceeding a specified rate limit.
type RateLimitMiddleware struct {
	limiter *speedbump.RateLimiter
	hasher  speedbump.RateHasher
	limit   int64
}

// NewRateLimitMiddleware creates a new instance of RateLimitMiddleware
func NewRateLimitMiddleware(limiter *speedbump.RateLimiter, hasher speedbump.RateHasher, limit int64) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter, hasher, limit}
}

// Handler is the middleware method.
func (m *RateLimitMiddleware) Handler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Rate Limit middleware started")

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		ok, err := m.limiter.Attempt(ip)
		if err != nil {
			panic(err)
		}

		nextTime := time.Now().Add(m.hasher.Duration())
		left, err := m.limiter.Left(ip)
		if err != nil {
			panic(err)
		}

		w.Header().Set("X-Rate-Limit-Limit", strconv.FormatInt(m.limit, 10))
		w.Header().Set("X-Rate-Limit-Remaining", strconv.FormatInt(left, 10))
		w.Header().Set("X-Rate-Limit-Reset", nextTime.String())

		if !ok {
			panic(errors.New(http.StatusTooManyRequests, "rate limit exceeded. Try again in "+nextTime.String()))
		}

		handler.ServeHTTP(w, r)
	})
}