// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package oauth2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hooklift/oauth2/providers/test"
	"github.com/hooklift/oauth2/types"
)

func AuthzGrantTokenRequestTest(t *testing.T, grantType, authzCode string) *http.Request {
	// http://tools.ietf.org/html/rfc6749#section-4.1.3
	queryStr := url.Values{
		"grant_type":   {grantType},
		"code":         {authzCode},
		"redirect_uri": {"https://example.com/oauth2/callback"},
		"client_id":    {"test_client_id"},
	}

	buffer := bytes.NewBufferString(queryStr.Encode())
	req, err := http.NewRequest("POST", "https://example.com/oauth2/tokens", buffer)
	ok(t, err)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	return req
}

// TestAccessToken tests a happy path for getting access tokens in accordance with
// http://tools.ietf.org/html/rfc6749#section-4.1.3 and
// http://tools.ietf.org/html/rfc6749#section-4.1.4
func TestAuthzGrantTokenRequest(t *testing.T) {
	cfg, authzCode := getTestAuthzCode(t)

	req := AuthzGrantTokenRequestTest(t, "authorization_code", authzCode)
	req.SetBasicAuth("testclient", "testclient")

	w := httptest.NewRecorder()
	IssueToken(w, req, cfg)

	// http://tools.ietf.org/html/rfc6749#section-4.1.4
	accessToken := types.Token{}
	err := json.Unmarshal(w.Body.Bytes(), &accessToken)
	ok(t, err)

	//log.Printf("%s", w.Body.String())
	equals(t, "bearer", accessToken.Type)
	equals(t, "600", accessToken.ExpiresIn)

	assert(t, accessToken.RefreshToken != "", "we were expecting a refresh token.")

	// Tests that cache headers are being sent when generating access tokens using
	// authorization grant codes.
	equals(t, "no-store", w.Header().Get("Cache-Control"))
	equals(t, "no-cache", w.Header().Get("Pragma"))
	equals(t, "0", w.Header().Get("Expires"))
}

// TestClientAuthRequired tests that client is required to always authenticate in order
// to request access tokens.
func TestAuthzGrantClientAuthRequired(t *testing.T) {
	cfg, authzCode := getTestAuthzCode(t)

	req := AuthzGrantTokenRequestTest(t, "authorization_code", authzCode)

	w := httptest.NewRecorder()
	IssueToken(w, req, cfg)
	// Tests for a 400 instead of 401 in accordance to http://tools.ietf.org/html/rfc6749#section-5.1
	equals(t, http.StatusBadRequest, w.Code)

	appErr := types.AuthzError{}
	err := json.Unmarshal(w.Body.Bytes(), &appErr)
	ok(t, err)
	equals(t, "unauthorized_client", appErr.Code)
}

// TestResourceOwnerCredentialsGrant tests happy path for http://tools.ietf.org/html/rfc6749#section-4.3
func TestResourceOwnerCredentialsGrant(t *testing.T) {
	cfg := setupTest()
	cfg.provider = test.NewProvider(true)

	queryStr := url.Values{
		"grant_type": {"password"},
		"username":   {"test_user"},
		"password":   {"test_password"},
	}

	buffer := bytes.NewBufferString(queryStr.Encode())
	req, err := http.NewRequest("POST", "https://example.com/oauth2/tokens", buffer)
	ok(t, err)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("testclient", "testclient")

	w := httptest.NewRecorder()
	IssueToken(w, req, cfg)

	accessToken := types.Token{}
	err = json.Unmarshal(w.Body.Bytes(), &accessToken)
	ok(t, err)

	//log.Printf("%s", w.Body.String())
	equals(t, "bearer", accessToken.Type)
	equals(t, "600", accessToken.ExpiresIn)

	assert(t, accessToken.RefreshToken != "", "we were expecting a refresh token.")

	// Tests that cache headers are being sent when generating tokens using
	// resource owner credentials.
	equals(t, "no-store", w.Header().Get("Cache-Control"))
	equals(t, "no-cache", w.Header().Get("Pragma"))
	equals(t, "0", w.Header().Get("Expires"))
}

// TestClientCredentialsGrant tests happy path for http://tools.ietf.org/html/rfc6749#section-4.4
func TestClientCredentialsGrant(t *testing.T) {
	cfg := setupTest()
	cfg.provider = test.NewProvider(true)

	queryStr := url.Values{
		"grant_type": {"client_credentials"},
	}

	buffer := bytes.NewBufferString(queryStr.Encode())
	req, err := http.NewRequest("POST", "https://example.com/oauth2/tokens", buffer)
	ok(t, err)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("testclient", "testclient")

	w := httptest.NewRecorder()
	IssueToken(w, req, cfg)

	accessToken := types.Token{}
	err = json.Unmarshal(w.Body.Bytes(), &accessToken)
	ok(t, err)

	//log.Printf("%s", w.Body.String())
	equals(t, "bearer", accessToken.Type)
	equals(t, "600", accessToken.ExpiresIn)

	// A refresh token SHOULD NOT be included.
	equals(t, "", accessToken.RefreshToken)

	// Tests that cache headers are being sent when generating tokens using
	// client credentials.
	equals(t, "no-store", w.Header().Get("Cache-Control"))
	equals(t, "no-cache", w.Header().Get("Pragma"))
	equals(t, "0", w.Header().Get("Expires"))
}

// TestRefreshToken tests happy path for http://tools.ietf.org/html/rfc6749#section-6
func TestRefreshToken(t *testing.T) {
	cfg := setupTest()
	provider := test.NewProvider(true)
	cfg.provider = provider

	noAuthzGrant := types.Grant{
		Scopes: types.Scopes{
			types.Scope{ID: "identity"},
		},
	}
	accessToken, err := provider.GenToken(noAuthzGrant, types.Client{
		ID: "test_client_id",
	}, true, cfg.tokenExpiration)
	ok(t, err)

	queryStr := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {accessToken.RefreshToken},
		"scope":         {"identity"},
	}

	buffer := bytes.NewBufferString(queryStr.Encode())
	req, err := http.NewRequest("POST", "https://example.com/oauth2/tokens", buffer)
	ok(t, err)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("testclient", "testclient")

	w := httptest.NewRecorder()
	IssueToken(w, req, cfg)

	token := types.Token{}
	err = json.Unmarshal(w.Body.Bytes(), &token)
	ok(t, err)

	//log.Printf("%s", w.Body.String())
	equals(t, "bearer", token.Type)
	equals(t, "600", token.ExpiresIn)
	assert(t, accessToken.Value != token.Value, "We got the same access token, it should be different!")
	assert(t, token.Value != "", "We were expecting to get a token and instead we got: %s", token.Value)
	assert(t, token.RefreshToken != "", "we were expecting a refresh token.")
	assert(t, token.RefreshToken != accessToken.RefreshToken, "We got the same refresh token, it should be different!")

	// Tests that cache headers are being sent when refreshing tokens
	equals(t, "no-store", w.Header().Get("Cache-Control"))
	equals(t, "no-cache", w.Header().Get("Pragma"))
	equals(t, "0", w.Header().Get("Expires"))
}

// TestAuthzCodeOwnership tests that the authorization code was issued to the client
// requesting the access token.
func TestAuthzCodeOwnership(t *testing.T) {
	cfg, authzCode := getTestAuthzCode(t)

	req := AuthzGrantTokenRequestTest(t, "authorization_code", authzCode)
	req.SetBasicAuth("boo", "boo")

	w := httptest.NewRecorder()
	IssueToken(w, req, cfg)

	// http://tools.ietf.org/html/rfc6749#section-4.1.4
	authzErr := types.AuthzError{}
	//log.Printf("%s", w.Body.String())
	err := json.Unmarshal(w.Body.Bytes(), &authzErr)
	ok(t, err)
	equals(t, "invalid_grant", authzErr.Code)
	equals(t, "Grant code was generated for a different redirect URI.", authzErr.Description)
}

// TestRevokeToken tests happy path for revoking refresh and access tokens.
// In accordance with https://tools.ietf.org/html/rfc7009
func TestRevokeToken(t *testing.T) {
	cfg, authzCode := getTestAuthzCode(t)

	r1 := AuthzGrantTokenRequestTest(t, "authorization_code", authzCode)
	r1.SetBasicAuth("testclient", "testclient")

	w := httptest.NewRecorder()
	IssueToken(w, r1, cfg)

	// http://tools.ietf.org/html/rfc6749#section-4.1.4
	accessToken := types.Token{}
	err := json.Unmarshal(w.Body.Bytes(), &accessToken)
	ok(t, err)

	r2, err := http.NewRequest("DELETE", "https://example.com/oauth2/tokens/"+accessToken.Value, nil)
	ok(t, err)
	r2.Header.Set("Content-type", "application/x-www-form-urlencoded")
	r2.SetBasicAuth("testclient", "testclient")

	w2 := httptest.NewRecorder()
	RevokeToken(w2, r2, cfg)
	equals(t, http.StatusOK, w2.Code)
}
