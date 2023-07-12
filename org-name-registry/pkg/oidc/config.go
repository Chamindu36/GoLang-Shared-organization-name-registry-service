package oidc

import (
	"errors"
)

// Structure to extract the OIDC related IDP configurations
type OidcIdpConfig struct {
	Issuer           string `yaml:"issuer"`
	ClientID         string `yaml:"clientId"`
	ClientSecret     string `yaml:"clientSecret"`
	RedirectURL      string `yaml:"redirectUrl"`
	TokenURL         string `yaml:"tokenUrl"`
	AuthorizationURL string `yaml:"authorizationUrl"`
	IntrospectURL    string `yaml:"introspectUrl"`
	Admin            struct {
		UserName string `yaml:"username"`
		PassWord string `yaml:"password"`
	} `yaml:"admin"`
}

// Validate method is used to validate the provided idp configurations
// @param cfg Provided configuration object
// @return error if validation failed
func (cfg *OidcIdpConfig) Validate() error {
	if isEmpty(cfg.Issuer) {
		return createErr("Identity provider not found in OIDC config")
	}
	if isEmpty(cfg.ClientID) {
		return createErr("Client id not found in OIDC config")
	}
	if isEmpty(cfg.ClientSecret) {
		return createErr("Client Secret not found in OIDC config")
	}
	if isEmpty(cfg.RedirectURL) {
		return createErr("Redirect Url not found in OIDC config")
	}
	if isEmpty(cfg.IntrospectURL) {
		return createErr("Introspect Url not found in OIDC config")
	}
	if isEmpty(cfg.Admin.UserName) {
		return createErr("Admin username cannot be empty")
	}
	if isEmpty(cfg.Admin.PassWord) {
		return createErr("Admin password cannot be empty")
	}

	return nil
}

func isEmpty(str string) bool {
	return len(str) == 0
}

func createErr(err string) error {
	return errors.New(err)
}

// Structure to extract response from the introspect call
type IntrospectResponse struct {
	Nbf       int    `json:"nbf"`
	Scopes    string `json:"scope"`
	Active    bool   `json:"active"`
	TokenType string `json:"token_type"`
	Exp       int    `json:"exp"`
	Iat       int    `json:"iat"`
	ClientId  string `json:"client_id"`
	Username  string `json:"username"`
}
