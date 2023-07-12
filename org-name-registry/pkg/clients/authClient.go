package clients

import "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/oidc"

// Auth client Interface to to perform OIDC validation
type Interface interface {
	AuthenticationClient() oidc.Interface
}

type AuthClient struct {
	authenticationClient oidc.Interface
}

func (a *AuthClient) AuthenticationClient() oidc.Interface {
	return a.authenticationClient
}

func NewAuthClient(authClient oidc.Interface) (*AuthClient, error) {
	return &AuthClient{
		authenticationClient: authClient,
	}, nil
}
