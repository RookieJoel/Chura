package keycloak

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
)

// tokenRefreshMargin is how long before expiry the cached token is replaced.
const tokenRefreshMargin = 30 * time.Second

// tokenSource caches the service-account access token obtained through the
// client-credentials grant. Refreshes are serialised by the mutex.
type tokenSource struct {
	client       *gocloak.GoCloak
	realm        string
	clientID     string
	clientSecret string
	now          func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (s *tokenSource) get(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.token != "" && s.now().Before(s.expires.Add(-tokenRefreshMargin)) {
		return s.token, nil
	}
	jwt, err := s.client.LoginClient(ctx, s.clientID, s.clientSecret, s.realm)
	if err != nil {
		return "", err
	}
	if jwt == nil || jwt.AccessToken == "" {
		return "", errors.New("empty access token in client-credentials response")
	}
	s.token = jwt.AccessToken
	s.expires = s.now().Add(time.Duration(jwt.ExpiresIn) * time.Second)
	return s.token, nil
}

// invalidate drops the cached token, e.g. after Keycloak answered 401.
func (s *tokenSource) invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = ""
}
