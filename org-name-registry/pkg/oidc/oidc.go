package oidc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/errors"
	"github.com/Chamindu36/organization-name-registry-service/pkg/logging"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Interface interface {
	AuthenticateToken(token string) (bool, error)
}

const (
	TokenPrefix          string = "token="
	BasicAuthPrefix      string = "Basic "
	AuthorizationHeader  string = "Authorization"
	ContentTypeHeader    string = "Content-Type"
	EncodedFormTypeValue string = "application/x-www-form-urlencoded"
	OrgRegServiceScope   string = "org_reg"
)

type Authenticator struct {
	oauth2Config *oauth2.Config
	oidcConfig   *oidc.Config
	issuer       string
	config       *OidcIdpConfig
	ctx          context.Context
}

// Oidc Authenticator to perform authentication process
func NewAuthenticator(cfg *OidcIdpConfig) *Authenticator {
	ctx := context.Background()

	config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthorizationURL,
			TokenURL: cfg.TokenURL,
		},
		RedirectURL: cfg.RedirectURL,
		Scopes:      []string{oidc.ScopeOpenID},
	}

	oidcConfig := &oidc.Config{
		ClientID: cfg.ClientID,
	}

	return &Authenticator{
		oauth2Config: config,
		oidcConfig:   oidcConfig,
		issuer:       cfg.Issuer,
		ctx:          ctx,
		config:       cfg,
	}
}

func (a *Authenticator) allowedRedirect(redirectUrl string) bool {
	if a.config.RedirectURL == redirectUrl {
		return true
	}
	return false
}

// AuthenticateToken method is used to authenticate the token of the request call
// @param token Token received to the registry service
// @param ownerEmail owner's email address
// @return true if the token is a valid token, false if the token validation failed
// @return error if any error occurred oe token validation failed
func (a *Authenticator) AuthenticateToken(token string) (bool, error) {

	//Extract admin details
	s := fmt.Sprintf("%s:%s", a.config.Admin.UserName, a.config.Admin.PassWord)
	// Encode admin username and password
	encodedCredentials := base64.URLEncoding.EncodeToString([]byte(s))
	payload := strings.NewReader(TokenPrefix + token)

	request, err := http.NewRequest(http.MethodPost, a.config.IntrospectURL, payload)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Introspect request call failed")
		return false, errors.Newf(errors.Error_INTERNAL, err, "Introspect request call failed")
	}
	request.Header.Set(ContentTypeHeader, EncodedFormTypeValue)
	request.Header.Set(AuthorizationHeader, BasicAuthPrefix+encodedCredentials)
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	// Introspect call to validate the token
	resp, err := client.Do(request)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Introspect request call failed")
		return false, errors.Newf(errors.Error_INTERNAL, err, "Introspect request call failed")
	}

	if resp.StatusCode != http.StatusOK {
		logging.NewDefaultLogger().Errorf("Introspect validation failed")
		return false, errors.Newf(errors.Error_INTERNAL, nil, "Introspect validation failed")
	}

	defer resp.Body.Close()
	// Read json http response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Response extraction failed with introspect call")
		return false, errors.Newf(errors.Error_INTERNAL, err, "Response extraction failed with introspect call")
	}
	var introspectData IntrospectResponse

	err = json.Unmarshal([]byte(body), &introspectData)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Response extraction failed with introspect call")
		return false, errors.Newf(errors.Error_INTERNAL, err, "Response extraction failed with introspect call")
	}

	if !introspectData.Active {
		return false, errors.Newf(errors.Error_NOT_AUTHORIZED, nil, "Token is not active")
	} else {
		if !strings.Contains(introspectData.Scopes, OrgRegServiceScope) {
			return false, errors.Newf(errors.Error_NOT_AUTHORIZED, nil, "Scope Validation failed")
		}
	}
	return true, nil
}
