package auth

import (
	"context"
	"net/http"
	"strings"

	oidc "github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type AuthProvider struct {
	tokenSource oauth2.TokenSource
}

type AuthProviderFactory func(ctx context.Context) (*AuthProvider, error)

// Returns a func that initialize a new auth provider which uses the client credentials flow
// Scopes must be provided as space separated: e.g.: "openid profile email"
// STS URL is expected to be the base url e.g: https://login.athosone.com
func NewClientCredentialsFactory(clientID string, clientSecret string, stsURL string, scopes string) AuthProviderFactory {
	return func(ctx context.Context) (*AuthProvider, error) {
		stsURL = strings.Trim(stsURL, "/")
		provider, err := oidc.NewProvider(ctx, stsURL)

		if err != nil {
			return nil, err
		}
		cfg := clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     provider.Endpoint().TokenURL,
			Scopes:       strings.Split(scopes, " "),
			AuthStyle:    oauth2.AuthStyleAutoDetect,
		}

		return &AuthProvider{
			tokenSource: cfg.TokenSource(ctx),
		}, nil
	}
}

// Returns a func that initialize a new auth provider which uses the basic auth flow
func NewBasicAuthFactory(username string, password string) AuthProviderFactory {
	return func(ctx context.Context) (*AuthProvider, error) {
		return &AuthProvider{
			tokenSource: &BasicTokenSource{Username: username, Password: password},
		}, nil
	}
}

// Returns a func that initialize a new auth provider that returns an empty token source.
// This is useful for testing.
func NewInsecure() *AuthProvider {
	return &AuthProvider{tokenSource: &InsecureTokenSource{}}
}

func (auth *AuthProvider) Authenticate(ctx context.Context, request *http.Request) error {
	token, err := auth.tokenSource.Token()
	if err != nil {
		return errors.Wrap(err, "Could not get an access token")
	}
	if request.Header == nil {
		request.Header = http.Header{}
	}
	token.SetAuthHeader(request)
	return nil
}
