// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package types defines oauth2 reusable types.
package types

import (
	"fmt"
	"net/url"
	"time"
)

// Client defines client information required by oauth2 to:
//   * Show an authorization form to a resource owner
//   * Validate that the provided request_uri parameter matches the one previously
//     registered for the client.
type Client struct {
	// Client's identifier.
	ID string
	// Client's name.
	Name string
	// Client's description.
	Description string
	// Profile image URL used when showing authorization form to resource owner
	ProfileImgURL *url.URL
	// Client's homepage URL to allow resource owners to verify client's authenticity by themselves.
	HomepageURL *url.URL
	// Redirect URL registered for this client.
	RedirectURL *url.URL
}

// Scope defines a type for manipulating OAuth2 scopes.
type Scope struct {
	// Scope's identifier. Example: read
	ID string
	// Scope's description
	Description string
}

// GrantCode represents an authorization grant code.
type GrantCode struct {
	// Authorization code value.
	Value string
	// Expiration time for this authorization code.
	ExpiresIn time.Duration
	// Client's identifier to which this code was emitted to.
	ClientID string
	// Redirect URL associated with the authorization code.
	RedirectURL *url.URL
	// List of authorization scopes for which this authorization code was generated.
	Scope []Scope
	// Whether or not this code was revoked.
	IsRevoked bool
	// Whether or not this code was expired.
	IsExpired bool
	// Whether or not this code was already used.
	IsUsed bool
}

// Token represents an access token.
type Token struct {
	// client associated to this token
	ClientID string `json:"-"`
	// The actual token value
	Value string `json:"access_token"`
	// Whether it is a bearer, MAC, SAML, etc
	Type string `json:"token_type"`
	// Expiration time for this token
	ExpiresIn string `json:"expires_in"`
	// Refresh token optionally emitted along with access token
	RefreshToken string `json:"refresh_token,omitempty"`
	// Authorization scoped allowed for this token
	Scope []Scope `json:"-"`
	// Whether or not this token was revoked.
	IsRevoked bool `json:"-"`
	// Whether or not this token was expired.
	IsExpired bool `json:"-"`
}

type AuthzError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
	URI         string `json:"error_uri,omitempty"`
	State       string `json:"state,omitempty"`
}

func (a *AuthzError) Error() string {
	str := fmt.Sprintf(`error="%s"`, a.Code)
	if a.Description != "" {
		str += fmt.Sprintf(`,error_description="%s"`, a.Description)
	}

	if a.URI != "" {
		str += fmt.Sprintf(`,error_uri="%s"`, a.URI)
	}
	return str
}
