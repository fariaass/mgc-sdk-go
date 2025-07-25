package containerregistry

import (
	"context"
	"net/http"

	mgc_http "github.com/MagaluCloud/mgc-sdk-go/internal/http"
)

type (
	// CredentialsService provides methods for managing container registry credentials
	CredentialsService interface {
		Get(ctx context.Context) (*CredentialsResponse, error)
		ResetPassword(ctx context.Context) (*CredentialsResponse, error)
	}

	// credentialsService implements the CredentialsService interface
	credentialsService struct {
		client *ContainerRegistryClient
	}

	// CredentialsResponse represents the response containing registry credentials
	CredentialsResponse struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
)

// Get retrieves the current container registry credentials
func (c *credentialsService) Get(ctx context.Context) (*CredentialsResponse, error) {
	path := "/v0/credentials"

	res, err := mgc_http.ExecuteSimpleRequestWithRespBody[CredentialsResponse](ctx, c.client.newRequest, c.client.GetConfig(), http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ResetPassword resets the password for the container registry credentials
func (c *credentialsService) ResetPassword(ctx context.Context) (*CredentialsResponse, error) {
	path := "/v0/credentials/password"

	res, err := mgc_http.ExecuteSimpleRequestWithRespBody[CredentialsResponse](ctx, c.client.newRequest, c.client.GetConfig(), http.MethodPost, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}
