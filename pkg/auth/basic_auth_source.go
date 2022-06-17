package auth

import (
	b64 "encoding/base64"
	"sync"

	oauth2 "golang.org/x/oauth2"
)

type BasicTokenSource struct {
	Username string
	Password string
	mu       sync.Mutex
}

func (b *BasicTokenSource) Token() (*oauth2.Token, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	password := b.Username + ":" + b.Password
	return &oauth2.Token{
		TokenType:   "Basic",
		AccessToken: b64.StdEncoding.EncodeToString([]byte(password)),
	}, nil
}
