package auth

import (
	"sync"

	oauth2 "golang.org/x/oauth2"
)

type InsecureTokenSource struct {
	mu sync.Mutex
}

func (i *InsecureTokenSource) Token() (*oauth2.Token, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return &oauth2.Token{AccessToken: ""}, nil
}
