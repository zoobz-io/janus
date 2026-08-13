package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Endpoints is the subset of an OIDC discovery document janus needs to run
// the authorization-code flow against any compliant issuer.
type Endpoints struct {
	AuthURL     string `json:"authorization_endpoint"`
	TokenURL    string `json:"token_endpoint"`
	UserinfoURL string `json:"userinfo_endpoint"`
}

// Discover fetches `{issuer}/.well-known/openid-configuration` and returns
// the endpoints janus uses. Endpoints are not assumed to live under the
// issuer prefix — providers like Google spread them across hosts, which is
// why discovery exists.
func Discover(ctx context.Context, issuer string) (*Endpoints, error) {
	url := strings.TrimSuffix(issuer, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating discovery request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discovery returned %d: %s", resp.StatusCode, body)
	}

	var endpoints Endpoints
	if err := json.NewDecoder(resp.Body).Decode(&endpoints); err != nil {
		return nil, fmt.Errorf("decoding discovery document: %w", err)
	}

	if endpoints.AuthURL == "" || endpoints.TokenURL == "" || endpoints.UserinfoURL == "" {
		return nil, fmt.Errorf("discovery document at %s is missing endpoints", url)
	}
	return &endpoints, nil
}
